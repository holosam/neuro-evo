package neuron

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

type EvolutionConfig struct {
	// Number of parents to crossover for each offspring.
	Parents int
	// Percent of species that die off each generation.
	BottomTierPercent float32
}

type MutationConfig struct {
	NeuronExpansion  float32
	SynapseExpansion float32

	AddNeuron  float32
	AddSynapse float32
	ChangeOp   float32
	SetSeed    float32
	UnsetSeed  float32
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

// Playground handles the organization and evolution of DNA.
type Playground struct {
	config PlaygroundConfig
	source *Conglomerate
	codes  map[IDType]*DNA
	rnd    *rand.Rand

	winner *DNA
}

func NewPlayground(config PlaygroundConfig) *Playground {
	return &Playground{
		config: config,
		source: NewConglomerate(),
		codes:  make(map[IDType]*DNA),
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Playground) InitDNA() {
	p.source.AddVisionAndMotor(p.config.NumInputs, p.config.NumOutputs)
	for id := 0; id < p.config.NumVariants; id++ {
		dna := NewDNA(p.source)
		p.mutateNeurons(dna)
		p.codes[id] = dna
	}
}

func (p *Playground) GetWinner() *DNA {
	return p.winner
}

func (p *Playground) SimulatePlayground() {
	// Each generation gets the set of inputs, competes, reproduces, and mutates.
	for gen := 0; gen < p.config.Generations; gen++ {
		g := NewGeneration(p.config.Gconf, p.codes)

		scores := g.FireBrains()

		// Sorts high to low (higher scores are better).
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score > scores[j].score
		})
		fmt.Printf("Gen %d scores: Max=%d 75th=%d 50th=%d 25th=%d Min=%d\n", gen,
			scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
			scores[3*len(scores)/4].score, scores[len(scores)-1].score)

		p.winner = p.codes[scores[0].id] //.DeepCopy()

		species := p.speciation(scores)

		newCodes := make(map[IDType]*DNA, p.config.NumVariants)
		currentMaxID := 0
		for speciesID, speciesScores := range species {
			speciesCodes := p.reproduction(speciesScores)
			for id, code := range speciesCodes {

				// Add dna mutation here

				// Also track species now to be used for the next generation

				newCodes[currentMaxID+id] = code
			}
			currentMaxID = len(newCodes)
		}

		for id, code := range newCodes {
			p.codes[id] = code
		}

		fmt.Printf("Winning DNA: %s\n", p.winner.PrettyPrint())
	}
}

// Break DNA into species based on the distance between their structures.
func (p *Playground) speciation(scores []BrainScore) map[IDType][]BrainScore {

}

func (p *Playground) dnaDistance(id1, id2 IDType) int {
	// Implement this for species.
	return 0
}

func (p *Playground) reproduction(scores []BrainScore) map[IDType]*DNA {
	dieOff := int(float32(len(scores)) * p.config.Econf.BottomTierPercent)
	scores = scores[:len(scores)-dieOff]

	newCodes := make(map[IDType]*DNA, len(scores))
	for id := 0; id < len(scores); id++ {

		scoreIndices := make(map[int]void, p.config.Econf.Parents)
		for {
			rndIndex := p.rnd.Intn(len(scores))
			if _, ok := scoreIndices[rndIndex]; !ok {
				scoreIndices[rndIndex] = member
			}
			if len(scoreIndices) == p.config.Econf.Parents {
				break
			}
		}

		parentScores := make([]BrainScore, 0)
		for i := 0; i < len(scores); i++ {
			if _, ok := scoreIndices[i]; ok {
				parentScores = append(parentScores, scores[i])
			}
		}

		newCodes[id] = p.createOffspring(parentScores)
	}

	return newCodes
}

// Overlay DNA on the conglomerate to line up genes.
func (p *Playground) createOffspring(dnaIDs []IDType) *DNA {
	child := NewDNA(p.source)

	seenEdges := make(IDSet, p.source.Synapses.nextID)
	for v := 0; v < p.source.NeuronIDs[SENSE].Length(); v++ {
		visionID := p.source.NeuronIDs[SENSE].GetId(v)
		p.traverseEdges(visionID, dnaIDs, child, seenEdges)
	}

	return child
}

func (p *Playground) traverseEdges(neuronID IDType, parentIDs []IDType, child *DNA, seenEdges IDSet) {
	// Any parent that has the source neuron is a contender.
	synContenders := make([]IDType, 0)
	for _, parentID := range parentIDs {
		if _, ok := p.codes[parentID].Neurons[neuronID]; !ok {
			continue
		}
		synContenders = append(synContenders, parentID)
	}

	for synID := range p.source.Synapses.srcMap[neuronID] {
		// This edge has already been evaluated in this run.
		if _, ok := seenEdges[synID]; ok {
			continue
		}
		seenEdges[synID] = member

		// Compute a percentage chance for this edge to be included in the child.
		inclusionChance := float32(0.0)
		synGeneChance := p.geneChance(len(synContenders))
		dstContenders := make([]IDType, 0)
		for parentIndex, parentID := range synContenders {
			if _, ok := p.codes[parentID].Synpases.idMap[synID]; ok {
				inclusionChance += synGeneChance[parentIndex]
				dstContenders = append(dstContenders, parentID)
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
		for index, chance := range p.geneChance(len(dstContenders)) {
			if rndVal < chance {
				dstIndex = index
				break
			}
			rndVal -= chance
		}
		child.SetNeuron(syn.dst, p.codes[dstContenders[dstIndex]].Neurons[syn.dst])

		p.traverseEdges(syn.dst, parentIDs, child, seenEdges)
	}
}

func (p *Playground) shiftConglomerate() {
	// Increase the number of neurons by the expansion percentage.
	neuronsToAdd := percentageOfWithMin1(p.source.Synapses.nextID, p.config.Mconf.NeuronExpansion)
	for i := 0; i < neuronsToAdd; i++ {
		synID := p.rnd.Intn(p.source.Synapses.nextID)
		// Okay to add a neuron on the same synapse more than once.
		p.source.AddInterNeuron(synID)
	}

	// Increase the number of synapses by the expansion percentage.
	newSynapses := percentageOfWithMin1(p.source.NeuronIDs[INTER].Length(), p.config.Mconf.SynapseExpansion)
	synCandidates := make(map[IDType]IDSet, 0)
	for i := 0; i < p.source.NeuronIDs[INTER].Length(); i++ {
		interID := p.source.NeuronIDs[INTER].GetId(p.rnd.Intn(p.source.NeuronIDs[INTER].Length()))

		// Repurpose newSynapses to also represent an approximate clump size.
		// Find nearby neurons that are newSynapses+1 away from this one.
		nearbyIDs := p.nearbyNeurons(interID, interID, newSynapses+1)
		for nearbyID := range nearbyIDs {
			nType := p.source.GetNeuronType(nearbyID)
			switch nType {
			case SENSE:
				p.addSynIfNotExists(nearbyID, interID, synCandidates)
			case INTER:
				p.addSynIfNotExists(nearbyID, interID, synCandidates)
				p.addSynIfNotExists(interID, nearbyID, synCandidates)
			case MOTOR:
				p.addSynIfNotExists(interID, nearbyID, synCandidates)
			}
		}
	}

	synCandidateList := make([]Synapse, 0)
	for src, dsts := range synCandidates {
		for dst := range dsts {
			synCandidateList = append(synCandidateList, Synapse{src: src, dst: dst})
		}
	}

	for i := 0; i < newSynapses; i++ {
		if len(synCandidateList) == 0 {
			break
		}

		rndIndex := p.rnd.Intn(len(synCandidateList))
		syn := synCandidateList[rndIndex]
		p.source.Synapses.AddNewSynapse(syn.src, syn.dst)

		// Remove the synapse from the list so it isn't chosen again.
		synCandidateList[rndIndex] = synCandidateList[len(synCandidateList)-1]
		synCandidateList = synCandidateList[:len(synCandidateList)-1]
	}
}

func (p *Playground) nearbyNeurons(startID, src IDType, hops int) IDSet {
	if hops == 1 {

		// Currently not working as intended
		// This doesn't yet contain upstream syns
		// Or, could change usage above to loop through all neurons, not just INTER

		return p.source.Synapses.dstMap[src]
	}

	nearby := make(IDSet)
	for dst := range p.source.Synapses.dstMap[src] {
		// Avoid going down the same pathways multiple times.
		if dst == startID {
			continue
		}
		for synID := range p.nearbyNeurons(startID, dst, hops-1) {
			if synID == startID {
				continue
			}
			nearby[synID] = member
		}
	}

	return nearby
}

func (p *Playground) addSynIfNotExists(src, dst IDType, synCandidates map[IDType]IDSet) {
	if dsts, ok := p.source.Synapses.dstMap[src]; ok {
		if _, ok = dsts[dst]; ok {
			return
		}
	}

	if _, ok := synCandidates[src]; !ok {
		synCandidates[src] = make(IDSet)
	}
	synCandidates[src][dst] = member
}

// Take a new offspring and (maybe) give it some new structure from the source.
func (p *Playground) mutateDNAStructure(dna *DNA) {
	// The only mutations that can occur on the conglomerate involve at least one
	// INTER neuron, so all neuron and synapse candidates are based on those.
	if p.mutationOccurs(p.config.Mconf.AddNeuron) {
		neuronCandidates := make([]IDType, 0)
		firstSyn := make([]Synapse, 0)
		secondSyn := make([]Synapse, 0)
		removeSyn := make([]Synapse, 0)

		// Find every neuron in the conglomerate that's between two neurons that
		// the DNA has.
		for src := range p.source.Synapses.dstMap {
			for mid := range p.source.Synapses.dstMap[src] {
				for dst := range p.source.Synapses.dstMap[mid] {
					_, hasSrc := dna.Neurons[src]
					_, hasMid := dna.Neurons[mid]
					_, hasDst := dna.Neurons[dst]
					if hasSrc && !hasMid && hasDst {
						neuronCandidates = append(neuronCandidates, mid)
						index := len(neuronCandidates) - 1
						firstSyn[index] = Synapse{src: src, dst: mid}
						secondSyn[index] = Synapse{src: mid, dst: dst}
						removeSyn[index] = Synapse{src: src, dst: dst}
					}
				}
			}
		}

		if len(neuronCandidates) > 0 {
			rndIndex := p.rnd.Intn(len(neuronCandidates))
			neuronID := neuronCandidates[rndIndex]
			dna.AddNeuron(neuronID, p.randomOp())

			firstAdd := firstSyn[rndIndex]
			firstID := p.source.Synapses.FindID(firstAdd.src, firstAdd.dst)
			dna.AddSynapse(firstID)

			secondAdd := secondSyn[rndIndex]
			secondID := p.source.Synapses.FindID(secondAdd.src, secondAdd.dst)
			dna.AddSynapse(secondID)

			remove := removeSyn[rndIndex]
			removeID := p.source.Synapses.FindID(remove.src, remove.dst)
			dna.RemoveSynapse(removeID)
		}
	}

	if p.mutationOccurs(p.config.Mconf.AddSynapse) {
		synCandidateList := make([]IDType, 0)
		for synID, syn := range p.source.Synapses.idMap {
			// Already has this synapse, so skip it.
			if _, hasSyn := dna.Synpases.idMap[synID]; hasSyn {
				continue
			}

			_, hasSrc := dna.Neurons[syn.src]
			_, hasDst := dna.Neurons[syn.dst]
			// Can add this synapse because it has both the src and destination.
			if hasSrc && hasDst {
				synCandidateList = append(synCandidateList, synID)
			}
		}

		if len(synCandidateList) > 0 {
			dna.AddSynapse(synCandidateList[p.rnd.Intn(len(synCandidateList))])
		}
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

func (p *Playground) geneChance(numParents int) []float32 {

	// Might be better to try inputting the actual BrainScores and returning
	// chances based on the ratios of scores.

	switch numParents {
	case 1:
		return []float32{1.0}
	case 2:
		return []float32{0.66, 0.34}
	case 3:
		return []float32{0.44, 0.33, 0.23}
	case 4:
		return []float32{0.32, 0.27, 0.23, 0.18}
	default:
		log.Fatalf("Unsupported parent number")
		return []float32{}
	}
}

func percentageOfWithMin1(val int, percent float32) int {
	out := val * int(100*percent) / 100
	if out == 0 {
		out = 1
	}
	return out
}
