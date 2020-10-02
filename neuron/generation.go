package neuron

import (
	"log"
	"math/rand"
)

type Generation struct {
	codes []DNA

	// Not supposed to use slices of pointers
	// brains []Brain
	brains map[int]*Brain

	rnd *rand.Rand // rand.NewSource(time.Now().UnixNano())
}

func (g *Generation) Generate(codes []DNA) {
	for codeID, code := range codes {
		g.brains[codeID] = Flourish(&code)
	}
}

func (g *Generation) FireBrains() {
	for _, brain := range g.brains {
		brain.StepFunction()
	}
}

func (g *Generation) MutateDNA(dnaIndex int) {
	if g.rnd.Float32() < 0.3 {
		g.codes[dnaIndex].AddSnippet(interpretOp(g.rnd.Intn(10)))
	}

	for snipIndex, snip := range g.codes[dnaIndex].snips {
		if g.rnd.Float32() < 0.03 {
			g.codes[dnaIndex].DeleteSnippet(snipIndex)
			continue
		}

		if g.rnd.Float32() < 0.05 {
			snip.SetOp(g.rnd.Intn(10))
		}
		if g.rnd.Float32() < 0.15 {
			snip.AddSynapse(g.RandSnippetId(dnaIndex))
		}
		for synapseIndex := range snip.Synapses {
			if g.rnd.Float32() < 0.05 {
				snip.RemoveSynapse(synapseIndex)
			}
		}
	}
}

func (g *Generation) RandSnippetId(dnaIndex int) int {
	randIndex := g.rnd.Intn(len(g.codes[dnaIndex].snips))
	for snipID := range g.codes[dnaIndex].snips {
		if randIndex == 0 {
			return snipID
		}
		randIndex--
	}
	log.Fatal("How did you get here?")
	return -1
}
