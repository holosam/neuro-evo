package neuron

import (
	"log"
	"math/rand"
	"sort"
	"time"
)

const (
	dnaSeedSnippets  = 10
	dnaSeedMutations = 20

	numSpeciesPerGen = 100
	maxStepsPerGen   = 500

	winnerRatio = 2
)

type Playground struct {
	codes     map[int]*DNA
	envInputs []SignalType
	accuracy  AccuracyFunc
	rnd       *rand.Rand
}

type SpeciesScore struct {
	id    int
	score int64
}

func NewPlayground(accuracy AccuracyFunc, envInputs []SignalType) *Playground {
	p := Playground{
		codes:     make(map[int]*DNA, numSpeciesPerGen),
		envInputs: envInputs,
		accuracy:  accuracy,
		rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for id := 0; id < numSpeciesPerGen; id++ {
		dna := p.initRandDNA()
		p.codes[id] = dna
	}

	return &p
}

func (p *Playground) SimulatePlayground() {
	results := RunGeneration(p.codes, p.envInputs)

	scores := make([]SpeciesScore, len(results))
	for id, result := range results {
		scores[id] = p.scoreResult(id, result)
	}

	// Need to figure out if this sorts high to low or low to high
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	newCodes := make(map[int]*DNA, numSpeciesPerGen)
	numBrainsPerSpecies := numSpeciesPerGen / winnerRatio
	speciesIndex := numSpeciesPerGen - 1
	winnerIndex := 0
	for numBrainsPerSpecies > 0 {
		for i := speciesIndex; i > numBrainsPerSpecies; i-- {
			newCodes[i] = p.codes[winnerIndex]
		}
		speciesIndex = numBrainsPerSpecies
		numBrainsPerSpecies /= winnerRatio
		winnerIndex++
	}
	p.codes = newCodes
}

func (p *Playground) scoreResult(id int, result *BrainResult) SpeciesScore {
	score := 1000000 * p.accuracy(result.moves)
	score += 1000 * int64(result.steps)
	score += int64(dnaComplexity(p.codes[id]))
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
