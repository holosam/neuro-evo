package neuron

import (
	"log"
	"math/rand"
	"time"
)

const (
	dnaSeedSnippets  = 10
	dnaSeedMutations = 20

	numSpeciesPerGen = 100
)

type Playground struct {
	codes map[int]*DNA

	gen *Generation

	rnd *rand.Rand
}

func NewPlayground() *Playground {
	p := Playground{
		codes: make(map[int]*DNA),
		gen:   NewGeneration(),
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for id := 0; id < numSpeciesPerGen; id++ {
		dna := p.initRandDNA()
		p.codes[id] = dna
		p.gen.AddSpecies(id, dna)
	}

	return &p
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
