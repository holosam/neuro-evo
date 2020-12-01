package neuron

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type GenInputsFunc func(round int) []SignalType

type ScoreType uint64

// FitnessFunc scores the output outputs. Closer to 0 is better.
type FitnessFunc func(inputs []SignalType, outputs []SignalType) ScoreType

type PlaygroundConfig struct {
	// Initialization
	DnaSeedSnippets  int
	DnaSeedMutations int

	// Running the playground
	NumSpecies   int
	Generations  int
	RoundsPerGen int
	GenInputsFn  GenInputsFunc

	// Evolution
	FitnessFn  FitnessFunc
	NumParents int

	// Nested configs
	Gconf GenerationConfig
}

type Playground struct {
	config PlaygroundConfig
	codes  map[IDType]*DNA
	rnd    *rand.Rand

	winner *DNA
}

type SpeciesScore struct {
	id         IDType
	score      ScoreType
	allOutputs [][]Signal
}

func NewPlayground(config PlaygroundConfig) *Playground {
	return &Playground{
		config: config,
		codes:  make(map[IDType]*DNA),
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Playground) SeedRandDNA() {
	for id := 0; id < p.config.NumSpecies; id++ {
		p.codes[id] = p.singleRandDNA()
	}
}

func (p *Playground) singleRandDNA() *DNA {
	dna := NewDNA()

	numVision := len(p.config.GenInputsFn(0))
	for i := 0; i < numVision; i++ {
		dna.AddSnippet(SENSE, p.randomOp())
	}

	for i := 0; i < p.config.DnaSeedSnippets-numVision-1; i++ {
		dna.AddSnippet(INTER, p.randomOp())
	}

	dna.AddSnippet(MOTOR, p.randomOp())

	for i := 0; i < p.config.DnaSeedMutations; i++ {
		p.mutateDNA(dna)
	}

	return dna
}

func (p *Playground) GetWinner() *DNA {
	return p.winner
}

func (p *Playground) SimulatePlayground() {
	// Each generation gets the set of inputs, competes, reproduces, and mutates.
	for gen := 0; gen < p.config.Generations; gen++ {
		g := NewGeneration(p.config.Gconf, p.codes)

		scores := make([]SpeciesScore, p.config.NumSpecies)
		for id := range p.codes {
			scores[id] = SpeciesScore{
				// Need to store the id as well because the slice gets sorted.
				id:         id,
				score:      0,
				allOutputs: make([][]Signal, p.config.RoundsPerGen),
			}
		}

		// Brains keep state between rounds.
		for round := 0; round < p.config.RoundsPerGen; round++ {
			inputs := p.config.GenInputsFn(round)
			results := g.FireBrains(inputs)
			// fmt.Printf("Gen %d, round %d, inputs = %v\n", gen+1, round+1, inputs)

			for id, result := range results {
				scores[id].score += p.scoreResult(id, result, inputs)
				scores[id].allOutputs = append(scores[id].allOutputs, result.Outputs)
				// fmt.Printf("ID %d got score %d, total is %d from output %v\n", id, p.scoreResult(id, result, inputs), scores[id].score, result.Outputs)
			}
		}

		// Sorts low to high (lower scores are better).
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score < scores[j].score
		})
		p.winner = p.codes[scores[0].id].DeepCopy()

		fmt.Printf("Gen %d scores: Min=%d 25th=%d 50th=%d 75th=%d Max=%d\n", gen,
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

func (p *Playground) scoreResult(id IDType, result *BrainResult, inputs []SignalType) ScoreType {
	outputs := make([]SignalType, len(result.Outputs))
	for i, signal := range result.Outputs {
		outputs[i] = signal.Output
	}

	score := p.config.FitnessFn(inputs, outputs)
	// score := 10000 * p.config.FitnessFn(inputs, outputs)
	// score -= 100 * ScoreType(result.steps)
	// score += 100 * ScoreType(result.steps)
	// score += 1 * ScoreType(dnaComplexity(p.codes[id]))
	return score
}

func dnaComplexity(dna *DNA) int {
	complexity := len(dna.Seeds)
	for _, snip := range dna.Snippets {
		complexity += 1 + len(snip.synapses)
	}
	return complexity
}

// Move to generation.go? Along with rounds and SpeciesScore
func (p *Playground) createOffspring(dnaIDs []IDType, scores []SpeciesScore) *DNA {
	child := NewDNA()

	// Seed the child DNA with starting IDs for each type of neuron, to avoid
	// collisions when traversing different parent DNAs.
	// Vision starts at 0.
	child.NeuronIDs[SENSE].InsertID(0)
	// Inter starts at len(vision) + 1.
	numVision := p.codes[0].NeuronIDs[SENSE].Length()
	child.NeuronIDs[INTER].InsertID(numVision)
	// Motor starts at the len(vision) + max(len(inter)) + 1.
	maxID := 0
	for _, parentID := range dnaIDs {
		numVisAndInter := numVision + p.codes[parentID].NeuronIDs[INTER].Length()
		if numVisAndInter > maxID {
			maxID = numVisAndInter
		}
	}
	child.NeuronIDs[MOTOR].InsertID(maxID)

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

func (p *Playground) mutateDNA(dna *DNA) {
	for i := 0; i < 3; i++ {
		nType := INTER
		if p.rnd.Float32() < 0.02 {
			nType = MOTOR
		}
		dna.AddSnippet(nType, p.randomOp())
	}

	for snipID, snip := range dna.Snippets {
		// Chance of changing the operation.
		if p.rnd.Float32() < 0.10 {
			snip.op = p.randomOp()
		}

		if p.rnd.Float32() < 0.05 {
			if _, exists := dna.Seeds[snipID]; exists {
				if p.rnd.Float32() < 0.50 {
					dna.RemoveSeed(snipID)
				} else {
					dna.SetSeed(snipID, SignalType(p.rnd.Intn(int(MaxSignal()))))
				}
			} else {
				dna.SetSeed(snipID, SignalType(p.rnd.Intn(int(MaxSignal()))))
			}
		}

		// Chance of a bit flip to create or remove a synapse to each other neuron.
		for possibleSynID := range dna.Snippets {
			if p.rnd.Float32() < 0.10 {
				if _, exists := snip.synapses[possibleSynID]; exists {
					dna.RemoveSynapse(snipID, possibleSynID)
				} else {
					// Try skipping direct vision->motor.
					// if (dna.NeuronIDs[SENSE].HasID(snipID) && dna.NeuronIDs[INTER].HasID(possibleSynID)) || (dna.NeuronIDs[INTER].HasID(snipID) && dna.NeuronIDs[MOTOR].HasID(possibleSynID)) {
					dna.AddSynapse(snipID, possibleSynID)
					// }
				}
			}
		}
	}
}

func (p *Playground) randomOp() OperatorType {
	return interpretOp(p.rnd.Intn(NumOps))
}
