package neuron

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type EvolutionConfig struct {
	// Number of parents to crossover for each offspring.
	Parents int
	// Percent of species that die off each generation.
	BottomTierPercent float32

	DistanceThreshold int
}

type MutationConfig struct {
	NeuronExpansion  float32
	SynapseExpansion float32

	AddNeuron  float32
	AddSynapse float32

	ChangeOp  float32
	SetSeed   float32
	UnsetSeed float32
}

type PlaygroundConfig struct {
	// Initialization
	NumInputs  int
	NumOutputs int

	// Running the playground
	NumVariants int
	Generations int

	// Nested configs
	Econf EvolutionConfig
	Gconf GenerationConfig
	Mconf MutationConfig
}

type Species struct {
	rep     *DNA
	scores  []BrainScore
	fitness ScoreType
}

func (s *Species) Size() int {
	return len(s.scores)
}

// Playground handles the organization and evolution of DNA.
type Playground struct {
	config  PlaygroundConfig
	source  *Conglomerate
	codes   map[IDType]*DNA
	species map[IDType]*Species
	rnd     *rand.Rand
}

func NewPlayground(config PlaygroundConfig) *Playground {
	return &Playground{
		config:  config,
		source:  NewConglomerate(),
		codes:   make(map[IDType]*DNA),
		species: make(map[IDType]*Species),
		rnd:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Playground) InitDNA() {
	p.source.AddVisionAndMotor(p.config.NumInputs, p.config.NumOutputs)
	for id := 0; id < p.config.NumVariants; id++ {
		dna := NewDNA(p.source)

		for _, nType := range neuronTypes {
			for i := 0; i < p.source.NeuronIDs[nType].Length(); i++ {
				dna.AddNeuron(p.source.NeuronIDs[nType].GetId(i), OR)
			}
		}
		for synID := range p.source.Synapses.idMap {
			dna.AddSynapse(synID)
		}

		for i := 0; i < 10; i++ {
			p.mutateNeurons(dna)
		}
		p.codes[id] = dna
	}
}

func (p *Playground) SimulatePlayground() {
	for gen := 0; gen < p.config.Generations; gen++ {
		g := NewGeneration(p.config.Gconf, p.codes)

		scores := g.FireBrains()

		// fmt.Printf("Gen %d scores: Max=%d 75th=%d 50th=%d 25th=%d Min=%d\n", gen,
		// 	scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
		// 	scores[3*len(scores)/4].score, scores[len(scores)-1].score)

		p.shiftConglomerate()

		speciesOffspring := p.speciation(scores)

		newCodes := make(map[IDType]*DNA, p.config.NumVariants)
		currentMaxID := 0
		for speciesID, species := range p.species {
			childCodes := p.reproduction(species, speciesOffspring[speciesID])
			for id, child := range childCodes {

				p.mutateDNAStructure(child)
				p.mutateNeurons(child)

				newCodes[currentMaxID+id] = child
			}

			currentMaxID = len(newCodes)
		}

		for _, species := range p.species {
			// Include one DNA from this generation to represent the species for the
			// next gen.
			species.rep = p.codes[species.scores[0].id]
			// Clear all members from the species since they are no longer needed.
			species.scores = make([]BrainScore, 0)
		}

		for id, code := range newCodes {
			p.codes[id] = code
		}
	}
}

// Break DNA into species based on the distance between their structures.
func (p *Playground) speciation(scores []BrainScore) map[IDType]int {
	// Figure out which species this genome belongs in.
	for _, score := range scores {
		foundSpecies := false
		nextSpeciesID := 0
		for speciesID, species := range p.species {
			if nextSpeciesID <= speciesID {
				nextSpeciesID = speciesID + 1
			}

			if p.dnaDistance(p.codes[score.id], species.rep) > p.config.Econf.DistanceThreshold {
				continue
			}
			foundSpecies = true
			species.scores = append(species.scores, score)
		}
		if !foundSpecies {
			p.species[nextSpeciesID] = &Species{
				rep:    p.codes[score.id],
				scores: []BrainScore{score},
			}
		}
	}

	// Adjust the fitness score for each member.
	totalGenerationFitness := ScoreType(0)
	for speciesID, species := range p.species {
		if species.Size() == 0 {
			delete(p.species, speciesID)
			continue
		}
		for index, score := range species.scores {
			adjustedFitness := score.score / ScoreType(species.Size())
			species.scores[index].score = adjustedFitness
			species.fitness += adjustedFitness
		}
		totalGenerationFitness += species.fitness
	}

	// Use the total fitness of the species to determine how many offspring
	// in the next generation are from each species.
	offspringPerSpecies := make(map[IDType]int, len(p.species))
	for speciesID, species := range p.species {
		offspringPerSpecies[speciesID] = percentageOfWithMin1(p.config.NumVariants,
			float32(species.fitness)/float32(totalGenerationFitness))
	}
	return offspringPerSpecies
}

func (p *Playground) dnaDistance(a, b *DNA) int {
	matchingEdges := 0
	for synID := range a.Synpases.idMap {
		if _, ok := b.Synpases.idMap[synID]; ok {
			matchingEdges++
		}
	}

	nonMatchingEdges := len(a.Synpases.idMap) + len(b.Synpases.idMap) - (2 * matchingEdges)

	// Add a factor that includes the neuron ops and seeds

	return nonMatchingEdges
}

func (p *Playground) reproduction(species *Species, numOffspring int) map[IDType]*DNA {
	// Sorts high to low (higher scores are better).
	sort.Slice(species.scores, func(i, j int) bool {
		return species.scores[i].score > species.scores[j].score
	})

	dieOff := percentageOfWithMin1(species.Size(), p.config.Econf.BottomTierPercent)
	species.scores = species.scores[:species.Size()-dieOff]

	newCodes := make(map[IDType]*DNA, numOffspring)

	// Can't reproduce without enough parents.
	if species.Size() < p.config.Econf.Parents {
		return newCodes
	}

	for id := 0; id < numOffspring; id++ {
		scoreIndices := make(IDSet, p.config.Econf.Parents)
		for {
			// Get N unique random numbers.
			rndIndex := p.rnd.Intn(species.Size())
			if _, ok := scoreIndices[rndIndex]; !ok {
				scoreIndices[rndIndex] = member
			}
			if len(scoreIndices) == p.config.Econf.Parents {
				break
			}
		}

		parentScores := make([]BrainScore, 0)
		for i := 0; i < species.Size(); i++ {
			if _, ok := scoreIndices[i]; ok {
				parentScores = append(parentScores, species.scores[i])
			}
		}

		newCodes[id] = p.createOffspring(parentScores)
	}

	return newCodes
}

// Overlay DNA on the conglomerate to line up genes.
func (p *Playground) createOffspring(parentScores []BrainScore) *DNA {
	child := NewDNA(p.source)

	seenEdges := make(IDSet, p.source.Synapses.nextID)
	for v := 0; v < p.source.NeuronIDs[SENSE].Length(); v++ {
		visionID := p.source.NeuronIDs[SENSE].GetId(v)
		p.traverseEdges(visionID, parentScores, child, seenEdges)
	}

	return child
}

func (p *Playground) traverseEdges(neuronID IDType, parentScores []BrainScore, child *DNA, seenEdges IDSet) {
	// Any parent that has the source neuron is a contender.
	synContenders := make([]BrainScore, 0)
	for _, parentScore := range parentScores {
		if _, ok := p.codes[parentScore.id].Neurons[neuronID]; !ok {
			continue
		}
		synContenders = append(synContenders, parentScore)
	}

	for synID := range p.source.Synapses.srcMap[neuronID] {
		// This edge has already been evaluated in this run.
		if _, ok := seenEdges[synID]; ok {
			continue
		}
		seenEdges[synID] = member

		// Compute a percentage chance for this edge to be included in the child.
		inclusionChance := float32(0.0)
		synGeneChance := geneChance(synContenders)
		dstContenders := make([]BrainScore, 0)
		for parentIndex, parentScore := range synContenders {
			if _, ok := p.codes[parentScore.id].Synpases.idMap[synID]; ok {
				inclusionChance += synGeneChance[parentIndex]
				dstContenders = append(dstContenders, parentScore)
			}
		}
		if !p.mutationOccurs(inclusionChance) {
			continue
		}

		// Add the synapse to the child.
		syn := p.source.Synapses.idMap[synID]
		child.Synpases.TrackSynapse(synID, syn.src, syn.dst)

		// If the dst neuron hasn't been added already, pick a random parent
		// with this synapse to pass on the neuron.
		if _, ok := child.Neurons[syn.dst]; ok {
			continue
		}
		rndVal := p.rnd.Float32()
		var dstIndex int
		for index, chance := range geneChance(dstContenders) {
			if rndVal < chance {
				dstIndex = index
				break
			}
			rndVal -= chance
		}
		child.SetNeuron(syn.dst, p.codes[dstContenders[dstIndex].id].Neurons[syn.dst])

		p.traverseEdges(syn.dst, parentScores, child, seenEdges)
	}
}

func (p *Playground) shiftConglomerate() {
	// Increase the number of neurons by the expansion percentage.
	neuronsToAdd := percentageOfWithMin1(p.source.NeuronIDs[INTER].Length(), p.config.Mconf.NeuronExpansion)
	for i := 0; i < neuronsToAdd; i++ {
		// Okay to add a neuron on the same synapse more than once.
		synID := p.rnd.Intn(p.source.Synapses.nextID)
		p.source.AddInterNeuron(synID)
	}

	// Increase the number of synapses by the expansion percentage.
	newSynapses := percentageOfWithMin1(p.source.Synapses.nextID, p.config.Mconf.SynapseExpansion)
	// dstCandidates := make(map[IDType]IDSet, 0)

	// Repurpose newSynapses to also represent an approximate clump size, so
	// new synapses are generally created with pretty close srcs and dsts.
	// Find nearby neurons that are newSynapses+1 away.
	nearbyNeurons := p.nearbyNeurons(newSynapses + 1)

	synCandidates := make([]Synapse, 0)
	for src, dsts := range nearbyNeurons {
		for dst := range dsts {
			if p.source.GetNeuronType(src) == MOTOR || p.source.GetNeuronType(dst) == SENSE {
				continue
			}

			_, err := p.source.Synapses.FindID(src, dst)
			if err == nil {
				// The synapse already exists, so continue without editing the candidates.
				continue
			}

			synCandidates = append(synCandidates, Synapse{src: src, dst: dst})
		}
	}

	// for _, srcType := range neuronTypes {
	// 	if srcType == MOTOR {
	// 		continue
	// 	}

	// 	for i := 0; i < p.source.NeuronIDs[srcType].Length(); i++ {
	// 		src := p.source.NeuronIDs[srcType].GetId(i)

	// 		dsts := p.downstreamNeurons(src, newSynapses+1)
	// 		for dst := range dsts {
	// 			dstType := p.source.GetNeuronType(dst)
	// 			p.addSynIfNotExists(src, dst, srcType, dstType, dstCandidates)
	// 			p.addSynIfNotExists(dst, src, srcType, dstType, dstCandidates)
	// 		}
	// 	}
	// }

	// synCandidates := make([]Synapse, 0)
	// for src, dsts := range dstCandidates {
	// 	for dst := range dsts {
	// 		synCandidates = append(synCandidates, Synapse{src: src, dst: dst})
	// 	}
	// }

	for i := 0; i < newSynapses; i++ {
		if len(synCandidates) == 0 {
			break
		}

		rndIndex := p.rnd.Intn(len(synCandidates))
		syn := synCandidates[rndIndex]
		p.source.Synapses.AddNewSynapse(syn.src, syn.dst)

		// Remove the synapse from the list so it isn't chosen again.
		synCandidates = removeIndexFromSynSlice(synCandidates, rndIndex)
	}
}

func (p *Playground) nearbyNeurons(hops int) map[IDType]IDSet {
	downstream := make(map[IDType]IDSet)
	for _, srcType := range neuronTypes {
		for i := 0; i < p.source.NeuronIDs[srcType].Length(); i++ {
			src := p.source.NeuronIDs[srcType].GetId(i)
			downstream[src] = p.source.Synapses.AllDsts(src)
		}
	}
	fmt.Printf("Downstream %v\n", downstream)

	upstream := make(map[IDType]IDSet)
	for src, dsts := range downstream {
		for dst := range dsts {
			if _, ok := upstream[dst]; !ok {
				upstream[dst] = make(IDSet)
			}
			upstream[dst][src] = member
		}
	}
	fmt.Printf("Upstream %v\n", upstream)

	nearbyNeurons := make(map[IDType]IDSet, 0)
	for _, srcType := range neuronTypes {
		for i := 0; i < p.source.NeuronIDs[srcType].Length(); i++ {
			src := p.source.NeuronIDs[srcType].GetId(i)

			nearbyNeurons[src] = make(IDSet)
			nearbyNeurons[src][src] = member

			for hop := 0; hop < hops; hop++ {
				pendingNearby := make(IDSet)
				for nearby := range nearbyNeurons[src] {
					for downstream := range downstream[nearby] {
						pendingNearby[downstream] = member
					}
					for upstream := range upstream[nearby] {
						pendingNearby[upstream] = member
					}
				}

				for nearby := range pendingNearby {
					nearbyNeurons[src][nearby] = member
				}
			}

			delete(nearbyNeurons[src], src)
		}
	}

	/*
		// First find all neurons #hops away.
		downstream := make(map[IDType]IDSet)
		for _, srcType := range neuronTypes {
			for i := 0; i < p.source.NeuronIDs[srcType].Length(); i++ {
				src := p.source.NeuronIDs[srcType].GetId(i)
				downstream[src] = p.downstreamNeurons(src, hops)
			}
		}
		fmt.Printf("Downstream %v\n", downstream)

		// The downstream map can be flipped to get upstream neurons.
		upstream := make(map[IDType]IDSet, 0)
		for src, dsts := range downstream {
			for dst := range dsts {
				if _, ok := upstream[dst]; !ok {
					upstream[dst] = make(IDSet)
				}
				upstream[dst][src] = member
			}
		}
		fmt.Printf("Upstream %v\n", upstream)

		nearbyNeurons := make(map[IDType]IDSet, 0)
		// For every one of my downstream, add their upstream.
		for src, downstreams := range downstream {
			if _, ok := nearbyNeurons[src]; !ok {
				nearbyNeurons[src] = make(IDSet)
			}
			for downstream := range downstreams {
				nearbyNeurons[src][downstream] = member
				for upstream := range upstream[downstream] {
					if src == upstream {
						continue
					}
					nearbyNeurons[src][upstream] = member
				}
			}
		}

	*/
	return nearbyNeurons
}

// func (p *Playground) downstreamNeurons(src IDType, hops int) IDSet {
// 	if hops == 0 {
// 		return make(IDSet, 0)
// 	}

// 	downstream := make(IDSet)
// 	for dst := range p.source.Synapses.AllDsts(src) {
// 		// Add all neurons immediately downstream of this one.
// 		downstream[dst] = member

// 		// Recurse to add all further downstream neurons.
// 		for downstreamDst := range p.downstreamNeurons(dst, hops-1) {
// 			downstream[downstreamDst] = member
// 		}
// 	}

// 	return downstream
// }

// Take a new offspring and (maybe) give it some new structure from the source.
// The only mutations that can occur on the conglomerate involve at least one
// INTER neuron, so all neuron and synapse candidates are based on those.
func (p *Playground) mutateDNAStructure(dna *DNA) {
	// Find every neuron in the conglomerate that's between two neurons that
	// the DNA has. So the DNA needs the src and dst but not the middle neuron.
	neuronCandidates := make([]IDType, 0)
	newSyn1 := make([]Synapse, 0)
	newSyn2 := make([]Synapse, 0)
	oldSyn := make([]Synapse, 0)
	for src := range p.source.Synapses.srcMap {
		if _, hasSrc := dna.Neurons[src]; !hasSrc {
			continue
		}

		for mid := range p.source.Synapses.AllDsts(src) {
			if _, hasMid := dna.Neurons[mid]; hasMid {
				continue
			}

			for dst := range p.source.Synapses.AllDsts(mid) {
				if _, hasDst := dna.Neurons[dst]; !hasDst {
					continue
				}

				// The same neuron ID may be added multiple times, but the surrounding
				// synapses will be different.
				neuronCandidates = append(neuronCandidates, mid)
				newSyn1 = append(newSyn1, Synapse{src: src, dst: mid})
				newSyn2 = append(newSyn2, Synapse{src: mid, dst: dst})
				oldSyn = append(oldSyn, Synapse{src: src, dst: dst})
			}
		}
	}

	neuronsToAdd := percentageOfWithMin1(len(dna.Neurons), p.config.Mconf.NeuronExpansion)
	for i := 0; i < neuronsToAdd; i++ {
		if len(neuronCandidates) == 0 {
			break
		}
		if !p.mutationOccurs(p.config.Mconf.AddNeuron) {
			continue
		}

		// Randomly pick which neuron will be added.
		// Then, remove that index from each list.
		rndIndex := p.rnd.Intn(len(neuronCandidates))
		neuronID := neuronCandidates[rndIndex]
		dna.AddNeuron(neuronID, p.randomOp())
		neuronCandidates = removeIndexFromIDSlice(neuronCandidates, rndIndex)

		newID1, _ := p.source.Synapses.FindID(newSyn1[rndIndex].src, newSyn1[rndIndex].dst)
		dna.AddSynapse(newID1)
		newSyn1 = removeIndexFromSynSlice(newSyn1, rndIndex)

		newID2, _ := p.source.Synapses.FindID(newSyn2[rndIndex].src, newSyn2[rndIndex].dst)
		dna.AddSynapse(newID2)
		newSyn2 = removeIndexFromSynSlice(newSyn2, rndIndex)

		oldID, _ := p.source.Synapses.FindID(oldSyn[rndIndex].src, oldSyn[rndIndex].dst)
		dna.RemoveSynapse(oldID)
		oldSyn = removeIndexFromSynSlice(oldSyn, rndIndex)
	}

	synCandidates := make([]IDType, 0)
	for synID, syn := range p.source.Synapses.idMap {
		// Already has this synapse, so skip it.
		if _, hasSyn := dna.Synpases.idMap[synID]; hasSyn {
			continue
		}

		_, hasSrc := dna.Neurons[syn.src]
		_, hasDst := dna.Neurons[syn.dst]
		// Can add this synapse because it has both the src and destination.
		if hasSrc && hasDst {
			synCandidates = append(synCandidates, synID)
		}
	}

	synsToAdd := percentageOfWithMin1(len(dna.Synpases.idMap), p.config.Mconf.SynapseExpansion)
	for i := 0; i < synsToAdd; i++ {
		if len(synCandidates) == 0 {
			break
		}
		if !p.mutationOccurs(p.config.Mconf.AddSynapse) {
			continue
		}

		rndIndex := p.rnd.Intn(len(synCandidates))
		dna.AddSynapse(synCandidates[rndIndex])
		synCandidates = removeIndexFromIDSlice(synCandidates, rndIndex)
	}
}

func (p *Playground) mutateNeurons(dna *DNA) {
	for _, neuron := range dna.Neurons {
		if p.mutationOccurs(p.config.Mconf.ChangeOp) {
			neuron.op = p.randomOp()
		}

		// Consider just making seeds 0 or MaxSignal?
		// Like I don't really see how adding 168 is going to be helpful

		if p.mutationOccurs(p.config.Mconf.SetSeed) {
			neuron.SetSeed(SignalType(p.rnd.Intn(int(MaxSignal()))))
		} else if p.mutationOccurs(p.config.Mconf.UnsetSeed) {
			neuron.RemoveSeed()
		}
	}
}

func (p *Playground) mutationOccurs(chance float32) bool {
	return p.rnd.Float32() <= chance
}

func (p *Playground) randomOp() OperatorType {
	return interpretOp(p.rnd.Intn(NumOps))
}

func geneChance(scores []BrainScore) []float32 {
	scoreTotal := ScoreType(0)
	for _, score := range scores {
		scoreTotal += score.score
	}

	geneChance := make([]float32, len(scores))
	for i := 0; i < len(scores); i++ {
		// The max uint64 (ScoreType) is less than the max float32.
		geneChance[i] = float32(scores[i].score) / float32(scoreTotal)
	}
	return geneChance
}

func percentageOfWithMin1(val int, percent float32) int {
	out := val * int(100*percent) / 100
	if out == 0 {
		out = 1
	}
	return out
}

// Assumes that the order of the slice doesn't matter.
func removeIndexFromIDSlice(s []IDType, index int) []IDType {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}
func removeIndexFromSynSlice(s []Synapse, index int) []Synapse {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}
