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
	FitnessFn           FitnessFunc
	NumSpeciesReproduce int // How many species reproduce (and die off).

	// Nested configs
	Gconf GenerationConfig
}

type Playground struct {
	config PlaygroundConfig
	codes  map[IDType]*DNA
	rnd    *rand.Rand
}

type SpeciesScore struct {
	id    IDType
	score ScoreType
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
		dna := NewDNA()
		for i := 0; i < p.config.DnaSeedSnippets; i++ {
			dna.AddSnippet(p.rnd.Intn(NumOps))
		}

		numVision := len(p.config.GenInputsFn(0))
		for i := 0; i < numVision; i++ {
			dna.VisionIDs.InsertID(i)
		}
		dna.MotorIDs.InsertID(p.config.DnaSeedSnippets - 1)

		for i := 0; i < p.config.DnaSeedMutations; i++ {
			p.mutateDNA(dna)
		}

		p.codes[id] = dna
	}
}

func (p *Playground) SimulatePlayground() {
	// Each generation gets the set of inputs, competes, reproduces, and mutates.
	for gen := 0; gen < p.config.Generations; gen++ {
		g := NewGeneration(p.config.Gconf, p.codes)

		scores := make([]SpeciesScore, p.config.NumSpecies)
		for round := 0; round < p.config.RoundsPerGen; round++ {
			inputs := p.config.GenInputsFn(round)
			results := g.FireBrains(inputs)

			for id, result := range results {
				// Need to store the id as well because the slice gets sorted.
				scores[id].id = id
				scores[id].score += p.scoreResult(id, result, inputs)
			}
		}

		// Sorts low to high (lower scores are better).
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score < scores[j].score
		})

		fmt.Printf("Scores for gen %d: Min=%d 25th=%d 50th=%d 75th=%d Max=%d\n", gen,
			scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
			scores[3*len(scores)/4].score, scores[len(scores)-1].score)

		for i := 0; i < p.config.NumSpeciesReproduce; i++ {
			winnerID := scores[i].id
			idToReplace := scores[p.config.NumSpecies-i-1].id

			// Create a copy of the underlying DNA struct to have different references
			// at each index even though the (pointer to the) source DNA is the same.
			cc := p.codes[winnerID].DeepCopy()
			p.mutateDNA(cc)
			p.codes[idToReplace] = cc
		}
	}

}
func (p *Playground) scoreResult(
	id IDType, result *BrainResult, inputs []SignalType) ScoreType {
	score := 10000 * p.config.FitnessFn(inputs, result.outputs)
	score += 100 * ScoreType(result.steps)
	score += 1 * ScoreType(dnaComplexity(p.codes[id]))
	return score
}

func dnaComplexity(dna *DNA) int {
	complexity := 0
	for _, snip := range dna.Snippets {
		complexity += 1 + len(snip.Synapses)
	}
	return complexity
}

func (p *Playground) mutateDNA(dna *DNA) {
	// Somewhat high chance of adding a new snippet.
	if p.rnd.Float32() < 0.3 {
		dna.AddSnippet(p.rnd.Intn(NumOps))
	}

	for snipID, snip := range dna.Snippets {
		// Low chance of deleting a snippet (as long as its not vision or motor).
		if !dna.VisionIDs.HasID(snipID) && !dna.MotorIDs.HasID(snipID) && p.rnd.Float32() < 0.01 {
			dna.DeleteSnippet(snipID)
			continue
		}

		// Chance of changing the operation
		if p.rnd.Float32() < 0.10 {
			snip.SetOp(p.rnd.Intn(NumOps))
		}

		// Chance of a bit flip to create or remove a synapse to each other neuron.
		for possibleSynID := range dna.Snippets {
			if p.rnd.Float32() < 0.05 {
				if _, exists := snip.Synapses[possibleSynID]; exists {
					snip.RemoveSynapse(possibleSynID)
				} else {
					snip.AddSynapse(possibleSynID)
				}
			}
		}
	}
}
