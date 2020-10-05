package neuron

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
)

type PlaygroundConfig struct {
	// Playground
	NumSpecies       int
	MaxGensPerPlay   int
	DnaSeedSnippets  int
	DnaSeedMutations int

	ContinueAfterAccurate bool

	// Generation
	MaxStepsPerGen int

	AccuracyFn AccuracyFunc
}

type Playground struct {
	codes  map[int]*DNA
	config PlaygroundConfig
	rnd    *rand.Rand
}

type SpeciesScore struct {
	id    int
	score int
}

func NewPlayground(config PlaygroundConfig) *Playground {
	p := Playground{
		codes:  make(map[int]*DNA),
		config: config,
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &p
}

func (p *Playground) SeedRandDNA() {
	for id := 0; id < p.config.NumSpecies; id++ {
		dna := p.initRandDNA()
		p.codes[id] = dna
	}
}

func (p *Playground) SeedKnownDNA(dna *DNA) {
	for id := 0; id < p.config.NumSpecies; id++ {
		p.codes[id] = dna.DeepCopy()
	}
}

// Run an evolution and return the best DNA after n generations.
func (p *Playground) SimulatePlayground(envInputs []SignalType) *DNA {
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

	return p.codes[0].DeepCopy()
}

func (p *Playground) setNextGenCodes(scores []SpeciesScore) {
	// Sorts low to high (lower scores are better).
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	fmt.Printf("Min=%d 25th=%d 50th=%d 75th=%d Max=%d\n", scores[0].score, scores[len(scores)/4].score, scores[2*len(scores)/4].score, scores[3*len(scores)/4].score, scores[len(scores)-1].score)

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

func (p *Playground) scoreResult(id int, result *BrainResult, inputs []SignalType) SpeciesScore {
	accScore := p.config.AccuracyFn(inputs, result.moves)
	// An accuracy score of 0 means this is the ideal output.
	// Prioritize getting a high accuracy first before paring down.
	score := 0

	score = accScore
	// if accScore != 0 {
	// 	score = accScore * 1000
	// 	score -= result.steps
	// } else {
	// 	score += 10 * result.steps
	// 	score += dnaComplexity(p.codes[id])
	// }

	return SpeciesScore{
		id:    id,
		score: score,
	}
}

func dnaComplexity(dna *DNA) int {
	complexity := len(dna.visionIDs) + len(dna.motorIDs)
	for _, snip := range dna.snips {
		complexity += 1 + len(snip.Synapses)
	}
	return complexity
}

func (p *Playground) initRandDNA() *DNA {
	dna := NewDNA()
	for i := 0; i < p.config.DnaSeedSnippets; i++ {
		dna.AddSnippet(p.rnd.Intn(NUM_OPS))
	}
	// Just winging this.
	dna.AddVisionId(0)
	dna.AddVisionId(1)
	dna.AddMotorId(p.config.DnaSeedSnippets - 1)

	for i := 0; i < p.config.DnaSeedMutations; i++ {
		p.mutateDNA(dna)
	}
	return dna
}

func (p *Playground) mutateDNA(dna *DNA) {
	// if p.rnd.Float32() < 0.3 {
	// 	dna.AddSnippet(p.rnd.Intn(NUM_OPS))
	// }

	for snipIndex, snip := range dna.snips {
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
			snip.SetOp(p.rnd.Intn(NUM_OPS))
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

func (p *Playground) randSnippetId(dna *DNA) int {
	randIndex := p.rnd.Intn(len(dna.snips))
	for snipID := range dna.snips {
		if randIndex == 0 {
			return snipID
		}
		randIndex--
	}
	log.Fatal("How did you get here?")
	return -1
}
