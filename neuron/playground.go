package neuron

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

type EvolutionConfig struct {
	// Number of parents to crossover for each offspring.
	Parents int
	// Percent of species that die off each generation.
	BottomTierPercent float32

	// Genome distance to be considered a different species.
	DistanceThreshold float32

	DistanceEdgeFactor      float32
	DistanceOperationFactor float32
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

	// Nested configs
	Econf EvolutionConfig
	// Gconf GenerationConfig
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

func (p *Playground) GetBrain(id IDType) *Brain {
	return Flourish(p.codes[id])
}

func (p *Playground) Evolve(scores []BrainScore) {
	fmt.Printf("Evolution beginning (at %v)\n", time.Now())
	p.shiftConglomerate()

	fmt.Printf("Beginning speciation at %v\n", time.Now())
	speciesOffspring := p.speciation(scores)
	fmt.Printf("Species offspring (at %v): %v\n", time.Now(), speciesOffspring)

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
	fmt.Printf("Done with reproduction (at %v)\n", time.Now())

	for speciesID, species := range p.species {
		if species.Size() == 0 {
			delete(p.species, speciesID)
			continue
		}

		// fmt.Printf("Species %d (size %d) has fitness %d, represented by \n%s\n",
		// 	speciesID, species.Size(), species.fitness, species.rep.PrettyPrint())

		// Include one DNA from this generation to represent the species for the
		// next gen.
		species.rep = p.codes[species.scores[0].id]
		// Clear all members from the species since they are no longer needed.
		species.scores = make([]BrainScore, 0)
		species.fitness = 0
	}

	for id, code := range newCodes {
		p.codes[id] = code
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
			// fmt.Printf("Adding score %+v: to existing species: %+v\n", score, species)
			break
		}

		// No existing species matched, so create a new one.
		if !foundSpecies {
			p.species[nextSpeciesID] = &Species{
				rep:    p.codes[score.id],
				scores: []BrainScore{score},
			}
			// fmt.Printf("Didn't find species, adding score %+v: to new species: %+v\n", score, p.species[nextSpeciesID])
		}
	}

	// Adjust the fitness score for each member.
	for speciesID, species := range p.species {
		// fmt.Printf("Adjusting species #%d (size %d) fitness: %+v\n", speciesID, species.Size(), species)
		if species.Size() == 0 {
			delete(p.species, speciesID)
			continue
		}
		for index, score := range species.scores {
			adjustedFitness := score.score / ScoreType(species.Size())
			species.scores[index].score = adjustedFitness
			species.fitness += adjustedFitness
			// fmt.Printf("Adjusted fitness for score %+v is %d\n", score, adjustedFitness)
		}
	}

	return p.partitionOffspring()
}

// Computes a number [0-1] for the distance between these two genomes based
// mostly on their structures and a bit on their neuron operations. Simply
// having matching neuronIDs is rather meaningless for how the genomes will
// operate, so this function attempts to compute the distance based on how
// different their outcomes will be.
func (p *Playground) dnaDistance(a, b *DNA) float32 {
	matchingEdges := 0
	matchingOperations := 0
	for synID := range a.Synpases.idMap {
		if syn, ok := b.Synpases.idMap[synID]; ok {
			matchingEdges++

			// If the src and dst neuron for this edge match, then count it.
			// This will naturally double count neurons, however it keeps with the
			// theme of computing genome distance based on edges.
			if a.Neurons[syn.src].IsEqual(b.Neurons[syn.src]) && a.Neurons[syn.dst].IsEqual(b.Neurons[syn.dst]) {
				matchingOperations++
			}
		}
	}

	totalEdges := len(a.Synpases.idMap) + len(b.Synpases.idMap)

	// The edge factor represents the distance between the structures of the two
	// genomes, by calculating the percentage of mismatched edges.
	nonMatchingEdges := totalEdges - (2 * matchingEdges)
	edgeFactor := float32(nonMatchingEdges) / float32(totalEdges)

	// The neuron factor represents how different the operations are on the
	// edges that do match.
	nonMatchingOperations := matchingEdges - matchingOperations
	neuronFactor := float32(nonMatchingOperations) / float32(matchingEdges)

	// The structure is weighted more than the operations.
	return p.config.Econf.DistanceEdgeFactor*edgeFactor + p.config.Econf.DistanceOperationFactor*neuronFactor
}

func (p *Playground) partitionOffspring() map[IDType]int {
	totalGenerationFitness := ScoreType(0)
	for _, species := range p.species {
		totalGenerationFitness += species.fitness
	}

	// Use the total fitness of the species to determine how many offspring
	// in the next generation are from each species.
	offspringPerSpecies := make(map[IDType]int, len(p.species))

	// It's possible for the sum of offspring to be different than NumVariants.
	// with simple rounding. For example, Variants = 3, species fitness =
	// [23, 3, 4], results become [2, 0, 0]. So instead, the remainder of the
	// fitness fraction gets added/subtracted to the next species.
	correction := 0.0

	baseValue := float64(p.config.NumVariants) / float64(totalGenerationFitness)
	for speciesID, species := range p.species {
		offspring := (float64(species.fitness) * baseValue) + correction
		offspringPerSpecies[speciesID] = int(math.Round(offspring))
		correction = offspring - float64(offspringPerSpecies[speciesID])
		// fmt.Printf("Species %d %+v gets %d offspring\n", speciesID, species, offspringPerSpecies[speciesID])
	}

	return offspringPerSpecies
}

func (p *Playground) reproduction(species *Species, numOffspring int) map[IDType]*DNA {
	// Sorts high to low (higher scores are better).
	sort.Slice(species.scores, func(i, j int) bool {
		return species.scores[i].score > species.scores[j].score
	})
	// fmt.Printf("Beginning reproduction: %+v\n", species)

	dieOff := percentageOfWithMin1(species.Size(), p.config.Econf.BottomTierPercent)
	species.scores = species.scores[:species.Size()-dieOff]

	newCodes := make(map[IDType]*DNA, numOffspring)

	// Can't reproduce without enough parents.
	if species.Size() < p.config.Econf.Parents {
		return newCodes
	}

	// The highest scoring variant of each species with more than 5 variants gets
	// directly copied to the next generation.
	if len(species.scores) >= 5 && numOffspring >= 5 {
		newCodes[numOffspring-1] = p.codes[species.scores[0].id].DeepCopy()
		numOffspring--
	}

	for id := 0; id < numOffspring; id++ {
		// fmt.Printf("-Making offspring %d\n", id)
		scoreIndices := make(IDSet, p.config.Econf.Parents)
		for {
			// Get #parents unique random numbers.
			rndIndex := p.rnd.Intn(species.Size())
			if _, ok := scoreIndices[rndIndex]; !ok {
				scoreIndices[rndIndex] = member
			}
			if len(scoreIndices) == p.config.Econf.Parents {
				break
			}
		}

		// Make a list of parents in decreasing score order.
		parentScores := make([]BrainScore, 0)
		for i := 0; i < species.Size(); i++ {
			if _, ok := scoreIndices[i]; ok {
				parentScores = append(parentScores, species.scores[i])
			}
		}
		// fmt.Printf("-Parents will be %v\n", parentScores)

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

		// Add the vision neuron to the child first.
		parentIndex := p.randomParentGene(parentScores)
		child.SetNeuron(visionID, p.codes[parentScores[parentIndex].id].Neurons[visionID])

		p.traverseEdges(visionID, parentScores, child, seenEdges)
	}

	return child
}

func (p *Playground) traverseEdges(neuronID IDType, parentScores []BrainScore, child *DNA, seenEdges IDSet) {
	// fmt.Printf("Evaluating neuron %d\n", neuronID)

	// Any parent that has the source neuron is a contender.
	synContenders := make([]BrainScore, 0)
	for _, parentScore := range parentScores {
		if _, ok := p.codes[parentScore.id].Neurons[neuronID]; !ok {
			continue
		}
		synContenders = append(synContenders, parentScore)
	}
	// fmt.Printf("-Parents %v have this neuron\n", synContenders)

	// Get all the possible synapses from each parent with this neuron.
	synCandidates := make(IDSet)
	for _, score := range synContenders {
		for synID := range p.codes[score.id].Synpases.srcMap[neuronID] {
			synCandidates[synID] = member
		}
	}

	for synID := range synCandidates {
		// This edge has already been evaluated in this run.
		if _, ok := seenEdges[synID]; ok {
			continue
		}
		seenEdges[synID] = member
		// fmt.Printf("--Evaluating syn %d from neuron %d\n", synID, neuronID)

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
		// fmt.Printf("--The inclusion chance is %v from parents %v\n", inclusionChance, dstContenders)
		if !p.mutationOccurs(inclusionChance) {
			continue
		}
		// fmt.Printf("--The mutation occurs and syn %d is added\n", synID)

		// Add the synapse to the child.
		syn := p.source.Synapses.idMap[synID]
		child.Synpases.TrackSynapse(synID, syn.src, syn.dst)

		// If the dst neuron hasn't been added already, pick a random parent
		// with this synapse to pass on the neuron.
		if _, ok := child.Neurons[syn.dst]; ok {
			continue
		}

		dstIndex := p.randomParentGene(dstContenders)
		child.SetNeuron(syn.dst, p.codes[dstContenders[dstIndex].id].Neurons[syn.dst])
		// fmt.Printf("--Adding neuron %d from parent %d\n", syn.dst, dstIndex)

		// Calling this function here makes this a DFS, which is already true
		// anyway because to be a BFS, each new neuron would need to be added to a
		// queue then popped off. However, the type of traversal doesn't affect the
		// outcome since it's the same chance of including each edge regardless of
		// the order it's evaluated in.
		p.traverseEdges(syn.dst, parentScores, child, seenEdges)
	}
}

func (p *Playground) shiftConglomerate() {
	// Increase the number of neurons by the expansion percentage.
	// neuronsToAdd := percentageOfWithMin1(p.source.NeuronIDs[INTER].Length(), p.config.Mconf.NeuronExpansion)
	neuronsToAdd := 1
	for i := 0; i < neuronsToAdd; i++ {
		// Okay to add a neuron on the same synapse more than once.
		synID := p.rnd.Intn(p.source.Synapses.nextID)
		newInterID := p.source.AddInterNeuron(synID)
		fmt.Printf("Shifting conglomerate: Adding new neuron %d on syn %d\n", newInterID, synID)
	}

	// Increase the number of synapses by the expansion percentage.
	// newSynapses := percentageOfWithMin1(p.source.Synapses.nextID, p.config.Mconf.SynapseExpansion)
	newSynapses := 4

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

	for i := 0; i < newSynapses; i++ {
		if len(synCandidates) == 0 {
			break
		}

		rndIndex := p.rnd.Intn(len(synCandidates))
		syn := synCandidates[rndIndex]
		newSynID := p.source.Synapses.AddNewSynapse(syn.src, syn.dst)
		fmt.Printf("Shifting conglomerate: Adding new synapse %+v with id %d\n", syn, newSynID)

		// Remove the candidate from the list so it isn't chosen again.
		synCandidates = removeIndexFromSynSlice(synCandidates, rndIndex)
	}
}

func (p *Playground) nearbyNeurons(hops int) map[IDType]IDSet {
	// Iterate through all the synapses to get every neighboring neuron,
	// regardless of the direction.
	neighbors := make(map[IDType]IDSet)
	for _, syn := range p.source.Synapses.idMap {
		if _, ok := neighbors[syn.src]; !ok {
			neighbors[syn.src] = make(IDSet)
		}
		if _, ok := neighbors[syn.dst]; !ok {
			neighbors[syn.dst] = make(IDSet)
		}
		neighbors[syn.src][syn.dst] = member
		neighbors[syn.dst][syn.src] = member
	}

	nearby := make(map[IDType]IDSet, 0)

	for src := range neighbors {
		nearby[src] = make(IDSet)
		// Seed the map with the src neuron. This is needed for the traversal
		// below, and is removed before returning.
		nearby[src][src] = member

		// The neighbors map has neurons that are only one hop away, so the number
		// of hops dictates how many times the neighbors are accessed.
		for i := 0; i < hops; i++ {
			// Build up a set of pending additions to the nearby map so its
			// not edited during the iteration.
			pendingNearby := make(IDSet)

			// For every nearby neuron, add all its neighbors.
			// On the first iteration (i=0), this just adds src's neighbors.
			// On subsequent iterations, this adds the neighbors' neighbors.
			for nearbyID := range nearby[src] {
				for neighbor := range neighbors[nearbyID] {
					pendingNearby[neighbor] = member
				}
			}

			// Now that nearby[src] has been fully iterated over, it's safe to add
			// the pending values.
			for pending := range pendingNearby {
				nearby[src][pending] = member
			}
		}

		// For the purposes of this function, a neuron is not near itself.
		delete(nearby[src], src)
	}

	return nearby
}

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

func (p *Playground) randomParentGene(parentScores []BrainScore) int {
	rndVal := p.rnd.Float32()
	var dstIndex int
	for index, chance := range geneChance(parentScores) {
		if rndVal < chance {
			dstIndex = index
			break
		}
		rndVal -= chance
	}
	return dstIndex
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
