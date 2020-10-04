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
	NumGensPerPlay   int
	DnaSeedSnippets  int
	DnaSeedMutations int
	WinnerRatio      int

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
	for i := 0; i < p.config.NumGensPerPlay; i++ {
		for _, code := range p.codes {
			p.mutateDNA(code)
		}

		results := RunGeneration(p.codes, envInputs, p.config.MaxStepsPerGen)

		scores := make([]SpeciesScore, len(results))
		minScore := math.MaxInt32
		for id, result := range results {
			scores[id] = p.scoreResult(id, result, envInputs)
			if scores[id].score < minScore {
				minScore = scores[id].score
			}
			// fmt.Printf("Gen %d: species #%d, result=%v, score=%d, and dna %s\n", i, id, result, scores[id].score, p.codes[id].PrettyPrint())
		}

		p.setNextGenCodes(scores)
		// fmt.Printf("Gen %d: best score %d from dna %s\n", i, minScore, p.codes[0].PrettyPrint())
		fmt.Printf("Gen %d: best score %d\n", i, minScore)
	}

	return p.codes[0].DeepCopy()
}

func (p *Playground) setNextGenCodes(scores []SpeciesScore) {
	// Sorts low to high (lower scores are better).
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	newCodes := make(map[int]*DNA, p.config.NumSpecies)

	// Copy the winning DNA the most times, the 2nd place DNA the second most, etc.
	numBrainsPerSpecies := p.config.NumSpecies / p.config.WinnerRatio
	// The highest score is at the first index of scores.
	winnerIndex := 0
	for i := p.config.NumSpecies; i > 0; i-- {
		if i == numBrainsPerSpecies {
			// numSpecies = 100, winnerRatio = 2. So when it hits 50, need to go to 25.
			// 50 /= 2 -> 25 /= 2 -> 12
			numBrainsPerSpecies /= p.config.WinnerRatio
			// Now grab the next best DNA from scores.
			winnerIndex++
		}

		// Create a copy of the underlying DNA struct to have different references
		// at each index even though the (pointer to the) source DNA is the same.
		// Tried this at first, but it didn't work because it copied over all the
		// Snippet pointers which were being modified by different references.
		// copiedDNA := *p.codes[scores[winnerIndex].id] // (didn't work)
		copiedDNA := p.codes[scores[winnerIndex].id].DeepCopy()

		// Insert so that best DNA starts at index 0 even though the loop counts down.
		newCodes[p.config.NumSpecies-i] = copiedDNA
	}
	p.codes = newCodes
}

func (p *Playground) scoreResult(id int, result *BrainResult, inputs []SignalType) SpeciesScore {
	score := 0
	if len(result.moves) > 0 {
		accScore := p.config.AccuracyFn(inputs, result.moves)
		// An accuracy score of 0 means this is the ideal output.
		// Prioritize getting a high accuracy first before paring down.
		score = accScore

		// if accScore != 0 {
		// 	score = accScore * 1000000
		// } else {
		// 	score += 1000 * result.steps
		// 	score += dnaComplexity(p.codes[id])
		// }

	} else {
		score = math.MaxInt32
	}
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
	dna.AddVisionId(0)
	dna.AddMotorId(p.config.DnaSeedSnippets - 1)

	for i := 0; i < p.config.DnaSeedMutations; i++ {
		p.mutateDNA(dna)
	}
	return dna
}

func (p *Playground) mutateDNA(dna *DNA) {
	if p.rnd.Float32() < 0.3 {
		dna.AddSnippet(p.rnd.Intn(NUM_OPS))
	}

	for snipIndex, snip := range dna.snips {
		// if p.rnd.Float32() < 0.01 {
		// 	dna.DeleteSnippet(snipIndex)
		// 	continue
		// }

		if p.rnd.Float32() < 0.02 {
			dna.AddVisionId(snipIndex)
		}
		if p.rnd.Float32() < 0.02 {
			dna.AddMotorId(snipIndex)
		}

		if _, exists := dna.visionIDs[snipIndex]; exists {
			if len(dna.visionIDs) >= 2 && p.rnd.Float32() < 0.03 {
				delete(dna.visionIDs, snipIndex)
			}
		}
		if _, exists := dna.motorIDs[snipIndex]; exists {
			if len(dna.motorIDs) >= 2 && p.rnd.Float32() < 0.03 {
				delete(dna.motorIDs, snipIndex)
			}
		}

		if p.rnd.Float32() < 0.10 {
			snip.SetOp(p.rnd.Intn(NUM_OPS))
		}
		if p.rnd.Float32() < 0.10 {
			snip.AddSynapse(p.randSnippetId(dna))
		}
		// for synapseIndex := range snip.Synapses {
		// 	if p.rnd.Float32() < 0.05 {
		// 		snip.RemoveSynapse(synapseIndex)
		// 	}
		// }
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
