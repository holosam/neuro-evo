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
	// When parents crossover, the fitter parent have their genes preferred.
	CrossoverPriority float32
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
		p.winner = p.codes[scores[0].id].DeepCopy()

		fmt.Printf("Gen %d scores: Max=%d 75th=%d 50th=%d 25th=%d Min=%d\n", gen,
			scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
			scores[3*len(scores)/4].score, scores[len(scores)-1].score)

		topTier := p.config.NumSpecies / 4
		newCodes := make(map[IDType]*DNA, topTier)
		for i := 0; i < topTier; i++ {
			parentScores := make([]SpeciesScore, 3)
			parentScores[0] = scores[p.rnd.Intn(topTier)]
			parentScores[1] = scores[p.rnd.Intn(topTier)]
			parentScores[2] = scores[p.rnd.Intn(topTier)+topTier]

			parents := make([]IDType, 3)
			parents[0] = parentScores[0].id
			parents[1] = parentScores[1].id
			parents[2] = parentScores[2].id

			child := p.createOffspring(parents, parentScores)
			if len(child.Snippets) == 0 {
				child = p.singleRandDNA()
			}
			p.mutateDNA(child)
			newCodes[scores[p.config.NumSpecies-i-1].id] = child
			// fmt.Printf("Breeding parent ids %v to get new child %s\n", parents, child.PrettyPrint())
		}

		for id, code := range newCodes {
			p.codes[id] = code
		}

		if gen%10 == 0 {
			fmt.Printf("Winning DNA: %s\n", p.winner.PrettyPrint())
		}
	}
}

func (p *Playground) createOffspring(dnaIDs []IDType, scores []SpeciesScore) *DNA {
	child := NewDNA()

	// Seed the child DNA with starting IDs for each type of neuron, to avoid
	// collisions when traversing different parent DNAs.
	// Vision starts at 0.
	child.NeuronIDs[SENSE].InsertID(0)
	// Motor starts at len(vision) + 1.
	numVision := p.codes[0].NeuronIDs[SENSE].Length()
	child.NeuronIDs[MOTOR].InsertID(numVision)
	// Inter starts at the len(vision) + max(len(inter) + 1.
	child.NeuronIDs[INTER].InsertID(numVision + p.codes[0].NeuronIDs[MOTOR].Length())

	//
	// Looked at up to here
	//

	for parentIndex, parentID := range dnaIDs {
		parent := p.codes[parentID]

		// Randomly choose the number of traversals to include for the child.
		// randNumTraversals := 1 + p.rnd.Intn(1+(parent.NumPathways()/(p.config.NumParents-1)))
		// p.config.RoundsPerGen is the same as len(scores[parentIndex].allOutputs)
		// totalTraversals := p.config.RoundsPerGen * len(scores[parentIndex].allOutputs[0])
		// A traversal will happen every `traversalCadence` iterations.
		// traversalCadence := totalTraversals / randNumTraversals

		if parent.NumPathways() == 0 { //|| totalTraversals == 0 {
			continue
		}

		paths := 0
		for _, outputs := range scores[parentIndex].allOutputs {
			for _, signal := range outputs {
				// fmt.Printf("i=%d, rpg=%d, j=%d, tc=%d\n", i, p.config.RoundsPerGen, j, traversalCadence)
				// if ((i*p.config.RoundsPerGen)+j)%traversalCadence == 0 {
				// fmt.Printf("Beginning traverse parent %s and child %s\n", parent.PrettyPrint(), child.PrettyPrint())

				if p.rnd.Intn(p.config.NumParents) == 0 {
					continue
				}

				p.randomTraversePathway(parent, child, &signal, -1)
				paths++
			}
		}
	}

	// Correct for the initial seeding (above).
	for _, nType := range neuronTypes {
		id := child.NeuronIDs[nType].GetId(0)
		if _, exists := child.Snippets[id]; !exists {
			child.NeuronIDs[nType].RemoveID(id)
		}
	}

	return child
}

// Switch to overlay
func (p *Playground) randomTraversePathway(parent, child *DNA, signal *Signal, prevChildID IDType) {
	// Base case: vision and seed signals don't have sources.
	if len(signal.sources) == 0 {
		return
	}

	// This child neuron's ID will become the index of the parent's neuron.
	// This helps overlay parent neurons that have different ID sets.
	nType := parent.GetNeuronType(signal.neuronID)
	parentNeuron := parent.Snippets[signal.neuronID]

	childID := child.NeuronIDs[nType].GetId(0) + parent.NeuronIDs[nType].GetIndex(signal.neuronID)

	// If the child doesn't yet have this neuron, create it.
	if snip, ok := child.Snippets[childID]; !ok {
		child.Snippets[childID] = NewNeuron(childID, parentNeuron.op)
		child.NeuronIDs[nType].InsertID(childID)
		if child.NextID <= childID {
			child.NextID = childID + 1
		}
	} else {
		// If the neuron exists, the op has a chance of being overridden.
		if p.rnd.Intn(p.config.NumParents) == 0 {
			snip.op = parentNeuron.op
		}
	}

	// Make a connection to the downstream neuron, maintaining the pathway.
	// Initial calls have -1 here, but motor neurons don't get synapses anyway.
	child.AddSynapse(childID, prevChildID)

	// Pick a random source to continue traversing.
	sourceVal := p.rnd.Intn(len(signal.sources))
	var sourceID IDType
	for id := range signal.sources {
		if sourceVal == 0 {
			sourceID = id
			break
		}
		sourceVal--
	}

	// If the source is negative then it could be a vision input or a seed.
	if sourceID < 0 {
		// If it's a seed, add it to the child's DNA.
		if seed, ok := parent.Seeds[signal.neuronID]; ok {
			child.Seeds[childID] = seed
		}
	} else {
		// Otherwise continue traversing up the tree.
		p.randomTraversePathway(parent, child, signal.sources[sourceID], childID)
	}
}

func (p *Playground) dnaSimilarity(dnaIDs []IDType) int {
	return 0
}

// Overlay DNA on the conglomerate to line up genes.
func (p *Playground) createOffspring(dnaIDs []IDType) *DNA {
	child := NewDNA(p.source)

	// Compute the percent chance to add for each parent.
	// The fittest parent (index 0) has the highest chance of passing on genes,
	// and so on.
	geneChance := make([]float32, p.config.Econf.Parents)
	// Base chance is the GCF of the parents and the priority chance.
	// For example, if there are 3 parents and the fitter parent is preferred
	// by 33%, then the base chance is 0.11, and the geneChance will end up as
	// [0.44, 0.33, 0.22]
	baseChance := 1.0 / (float32(p.config.Econf.Parents) * (1.0 / p.config.Econf.CrossoverPriority))

	// This math is wrong, the p.config.Econf.Parents+1 below is only right for 3

	// For example, if there are 4 parents and the fitter parent is preferred
	// by 20%, then the base chance is 1 / (4 * 5) 0.05,
	// and the geneChance will end up as [0.25, 0.20, 0.15, 0.10]

	totalGeneChance := float32(0.0)
	for i := 0; i < p.config.Econf.Parents; i++ {
		// geneChance[0] = 0.11 * (4-0) = 0.44
		// geneChance[1] = 0.11 * (4-1) = 0.33
		geneChance[i] = baseChance * float32(p.config.Econf.Parents+1-i)
		totalGeneChance += geneChance[i]
	}

	// Find out a way to ensure this with the code
	if totalGeneChance < 1.0 {
		log.Fatalf("total gene chance = %v", totalGeneChance)
	}

	// Track the percentage chance that this synapse gets passed on.
	synGeneChance := make(map[IDType]float32, p.source.Synapses.nextID)
	for dnaIndex, dnaID := range dnaIDs {
		for synID := range p.codes[dnaID].Synpases.idMap {
			synGeneChance[synID] = synGeneChance[synID] + geneChance[dnaIndex]
		}
	}

	// Pick up here
	// Trying to figure out how to do the crossover (yellow notepads)

	for synID, chance := range synGeneChance {
		if !p.mutationOccurs(chance) {
			continue
		}

		syn := p.source.Synapses.idMap[synID]
		child.AddSynapse(synID)
		if _, hasSrc := child.Neurons[syn.src]; !hasSrc {

		}
	}

	return child
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

func percentageOfWithMin1(val int, percent float32) int {
	out := val * int(100*percent) / 100
	if out == 0 {
		out = 1
	}
	return out
}
