package neuron

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

const (
	dnaSeedSnippets  = 10
	dnaSeedMutations = 20

	maxStepsPerGen = 500

	winnerRatio = 2
)

type Playground struct {
	codes map[int]*DNA
	rnd   *rand.Rand
}

type SpeciesScore struct {
	id    int
	score int
}

func NewPlayground() *Playground {
	p := Playground{
		codes: make(map[int]*DNA),
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &p
}

func (p *Playground) SeedRandDNA(numSpecies int) {
	for id := 0; id < numSpecies; id++ {
		dna := p.initRandDNA()
		p.codes[id] = dna
	}
}

// Run an evolution and return the best DNA after n generations.
func (p *Playground) SimulatePlayground(n int, envInputs []SignalType, accuracy AccuracyFunc) *DNA {
	fmt.Printf("Beginning evolution with input %v\n", envInputs)
	for i := 0; i < n; i++ {
		results := RunGeneration(p.codes, envInputs)

		scores := make([]SpeciesScore, len(results))
		for id, result := range results {
			scores[id] = p.scoreResult(id, result, accuracy)
			fmt.Printf("Gen %d: species #%d, score %d\n", i, id, scores[id])
		}

		p.setNextGenCodes(scores)
	}

	return p.codes[0]
}

func (p *Playground) setNextGenCodes(scores []SpeciesScore) {
	// Sorts low to high (lower scores are better).
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	numSpecies := len(scores)
	newCodes := make(map[int]*DNA, numSpecies)

	// Copy the winning DNA the most times, the 2nd place DNA the second most, etc.
	numBrainsPerSpecies := numSpecies / winnerRatio
	// The highest score is at the first index of scores.
	winnerIndex := 0
	for i := numSpecies; i > 0; i-- {
		if i == numBrainsPerSpecies {
			// numSpecies = 100, winnerRatio = 2. So when it hits 50, need to go to 25.
			// 50 /= 2 -> 25 /= 2 -> 12
			numBrainsPerSpecies /= winnerRatio
			// Now grab the next best DNA from scores.
			winnerIndex++
		}
		// Insert so that best DNA starts at index 0 even though the loop counts down.
		newCodes[numSpecies-i] = p.codes[scores[winnerIndex].id]
	}
	p.codes = newCodes
}

func (p *Playground) scoreResult(id int, result *BrainResult, accuracy AccuracyFunc) SpeciesScore {
	score := 1000000 * accuracy(result.moves)
	score += 1000 * result.steps
	score += dnaComplexity(p.codes[id])
	return SpeciesScore{
		id:    id,
		score: score,
	}
}

func dnaComplexity(dna *DNA) int {
	complexity := 0
	for _, snip := range dna.snips {
		complexity += 1 + len(snip.Synapses)
	}
	return complexity
}

func (p *Playground) initRandDNA() *DNA {
	dna := NewDNA()
	for i := 0; i < dnaSeedSnippets; i++ {
		dna.AddSnippet(p.rnd.Intn(10))
	}
	dna.AddVisionId(0)
	dna.AddMotorId(dnaSeedSnippets - 1)

	for i := 0; i < dnaSeedMutations; i++ {
		p.mutateDNA(dna)
	}
	return dna
}

func (p *Playground) mutateDNA(dna *DNA) {
	if p.rnd.Float32() < 0.3 {
		dna.AddSnippet(p.rnd.Intn(10))
	}

	for snipIndex, snip := range dna.snips {
		if p.rnd.Float32() < 0.03 {
			dna.DeleteSnippet(snipIndex)
			continue
		}
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

		if p.rnd.Float32() < 0.05 {
			snip.SetOp(p.rnd.Intn(10))
		}
		if p.rnd.Float32() < 0.15 {
			snip.AddSynapse(p.randSnippetId(dna))
		}
		for synapseIndex := range snip.Synapses {
			if p.rnd.Float32() < 0.05 {
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
