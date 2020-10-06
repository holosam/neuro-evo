package neuron

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

// FitnessFunc scores the output outputs. Closer to 0 is better.
type FitnessFunc func(inputs []SignalType, outputs []SignalType) int

type PlaygroundConfig struct {
	// Running the playground
	NumSpecies  int
	Generations int

	// Competition
	RoundsOfCompetition int // How many times to play against another species.
	NumCompetitions     int // How many other species to play against.

	// Evolution
	SpeciesReproduce int // How many species reproduce (and die off).

	// Creating species
	DnaSeedSnippets  int
	DnaSeedMutations int

	// Scoring
	FitnessFn FitnessFunc

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
	score int
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
		dna := p.initRandDNA()
		p.codes[id] = dna
	}
}

func (p *Playground) SimulateCompetition(envInputs []SignalType) {
	// Each generation gets the set of inputs, competes, reproduces, and mutates.
	for gen := 0; gen < p.config.Generations; gen++ {
		g := NewGeneration(p.config.Gconf, p.codes)

		// Send the environment inputs to every species to begin.
		inputs := make(map[IDType][]SignalType, p.config.NumSpecies)
		for id := range p.codes {
			inputs[id] = envInputs
		}
		results := g.FireBrains(inputs)

		scores := make([]SpeciesScore, len(results))

		// Each competition pits two species against each other.
		// for comp := 0; comp < p.config.NumCompetitions; comp++ {
		for i := 0; i < p.config.NumSpecies; i++ {
			for j := i + 1; j < p.config.NumSpecies; j++ {
				// PICK UP HERE
				// To do every combination of brains, it seems like it makes more sense
				// to do `go FireBrain` and then have it write the results
				// It should be a lot easier here to have a struct (or add to playground)
				// with a mutex and have functions like
				// GetNextInput(id) which waits on the other brain
				// try using thread safe structs instead of passing channels
			}
		}

		for round := 0; round < p.config.RoundsOfCompetition; round++ {

			results = g.FireBrains(inputs)
			for id, result := range results {
				scores[id].score += p.config.FitnessFn(result.inputs, result.outputs)
			}
		}
		// }

		// Sorts low to high (lower scores are better).
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score < scores[j].score
		})

		fmt.Printf("Scores for gen %d: Min=%d 25th=%d 50th=%d 75th=%d Max=%d\n", gen,
			scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
			scores[3*len(scores)/4].score, scores[len(scores)-1].score)

		for i := 0; i < p.config.SpeciesReproduce; i++ {
			winnerID := scores[i].id
			idToReplace := scores[p.config.NumSpecies-i-1].id

			// Create a copy of the underlying DNA struct to have different references
			// at each index even though the (pointer to the) source DNA is the same.
			p.codes[idToReplace] = p.codes[winnerID].DeepCopy()
			p.mutateDNA(p.codes[idToReplace])
		}
	}

}

/*
// SimulatePlayground runs a full generation of DNA.
func (p *Playground) SimulatePlayground(envInputs []SignalType) {
	fmt.Printf("Beginning evolution with input %v\n", envInputs)
	for i := 0; i < p.config.MaxGensPerPlay; i++ {
		for id, code := range p.codes {
			// fmt.Printf("Gen %d: species #%d, start dna %s\n", i, id, p.codes[id].PrettyPrint())
			if id%2 == 0 {
				continue
			}
			for m := 0; m < 2; m++ {
				p.mutateDNA(code)
			}
			// fmt.Printf("Gen %d: species #%d, mutated dna %s\n", i, id, p.codes[id].PrettyPrint())
		}

		results := RunGeneration(p.codes, envInputs, p.config.MaxStepsPerGen)

		scores := make([]SpeciesScore, len(results))
		minScore := math.MaxInt32
		for _, result := range results {
			id := result.id
			scores[id] = p.scoreResult(id, result, envInputs)
			if scores[id].score < minScore {
				minScore = scores[id].score
			}
			// fmt.Printf("Gen %d: species #%d, result=%v, score=%d, and dna %s\n", i, id, result, scores[id].score, p.codes[id].PrettyPrint())
		}

		p.setNextGenCodes(scores)

		// if minScore < 980 {
		// 	break
		// }

		// fmt.Printf("Gen %d: best score %d\n", i, minScore)
		// if minScore < 1000000 && !p.config.ContinueAfterAccurate {
		// The best score possible for this playground has been acheived,
		// should move on to different playgrounds instead of spinning wheels.
		// break
		// }
	}
}

func (p *Playground) setNextGenCodes(scores []SpeciesScore) {
	// Sorts low to high (lower scores are better).
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	fmt.Printf("Min=%d 25th=%d 50th=%d 75th=%d Max=%d\n",
		scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score,
		scores[3*len(scores)/4].score, scores[len(scores)-1].score)

	numToCopy := 2

	newCodes := make(map[int]*DNA, p.config.NumSpecies)
	for i := 0; i < p.config.NumSpecies; i += numToCopy {
		// Create a copy of the underlying DNA struct to have different references
		// at each index even though the (pointer to the) source DNA is the same.
		// Tried this at first, but it didn't work because it copied over all the
		// Snippet pointers which were being modified by different references.
		// copiedDNA := *p.codes[scores[winnerIndex].id] // (didn't work)
		for j := i; j < i+numToCopy; j++ {
			newCodes[j] = p.codes[scores[i/numToCopy].id].DeepCopy()
		}
	}
	p.codes = newCodes
}

func (p *Playground) scoreResult(
	id int, result *BrainResult, inputs []SignalType) SpeciesScore {
	// An accuracy score of 0 means this is the ideal output.
	// Prioritize getting a high accuracy first before paring down.
	// if accScore != 0 {
	// 	score = accScore * 1000
	// 	score -= result.steps
	// } else {
	// 	score += 10 * result.steps
	// 	score += dnaComplexity(p.codes[id])
	// }

	return SpeciesScore{
		id:    id,
		score: p.config.AccuracyFn(inputs, result.moves),
	}
}

func dnaComplexity(dna *DNA) int {
	complexity := len(dna.visionIDs) + len(dna.motorIDs)
	for _, snip := range dna.snippets {
		complexity += 1 + len(snip.Synapses)
	}
	return complexity
}
*/

func (p *Playground) initRandDNA() *DNA {
	dna := NewDNA()
	for i := 0; i < p.config.DnaSeedSnippets; i++ {
		dna.AddSnippet(p.rnd.Intn(NumOps))
	}

	// Just winging this.
	dna.visionIDs.InsertID(0)
	dna.visionIDs.InsertID(1)
	dna.motorIDs.InsertID(p.config.DnaSeedSnippets - 1)

	for i := 0; i < p.config.DnaSeedMutations; i++ {
		p.mutateDNA(dna)
	}
	return dna
}

func (p *Playground) mutateDNA(dna *DNA) {
	// if p.rnd.Float32() < 0.3 {
	// 	dna.AddSnippet(p.rnd.Intn(NUM_OPS))
	// }

	for snipIndex, snip := range dna.snippets {
		// if p.rnd.Float32() < 0.01 {
		// 	dna.DeleteSnippet(snipIndex)
		// 	continue
		// }

		// if p.rnd.Float32() < 0.02 {
		// 	dna.AddVisionId(snipIndex)
		// }
		// if p.rnd.Float32() < 0.02 {
		// 	dna.AddMotorId(snipIndex)
		// }

		// if _, exists := dna.visionIDs[snipIndex]; exists {
		// 	if len(dna.visionIDs) >= 2 && p.rnd.Float32() < 0.03 {
		// 		delete(dna.visionIDs, snipIndex)
		// 	}
		// }
		// if _, exists := dna.motorIDs[snipIndex]; exists {
		// 	if len(dna.motorIDs) >= 2 && p.rnd.Float32() < 0.03 {
		// 		delete(dna.motorIDs, snipIndex)
		// 	}
		// }

		if p.rnd.Float32() < 0.10 {
			snip.SetOp(p.rnd.Intn(NumOps))
		}
		// if p.rnd.Float32() < 0.10 {
		// 	snip.AddSynapse(p.randSnippetId(dna))
		// }
		if p.rnd.Float32() < 0.30 {
			toAdd := p.randSnippetId(dna)
			if toAdd != snipIndex {
				snip.AddSynapse(toAdd)
			}
		}
		for synapseIndex := range snip.Synapses {
			if p.rnd.Float32() < 0.10 {
				toAdd := p.randSnippetId(dna)
				if toAdd != snipIndex {
					snip.AddSynapse(toAdd)
				}
			}
			if p.rnd.Float32() < 0.10 {
				snip.RemoveSynapse(synapseIndex)
			}
		}
	}
}

func (p *Playground) randSnippetId(dna *DNA) IDType {
	randIndex := p.rnd.Intn(len(dna.snippets))
	for snipID := range dna.snippets {
		if randIndex == 0 {
			return snipID
		}
		randIndex--
	}
	log.Fatal("How did you get here?")
	return -1
}
