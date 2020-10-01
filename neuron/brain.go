package neuron

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

const dnaSep = "|"

// Brain docs
type Brain struct {
	DNA string

	neurons []Neuron

	sigChan chan Signal

	mu sync.Mutex

	pendingSignals map[int][]SignalType
}

func Flourish(dna string) *Brain {
	snippets := strings.Split(dna, dnaSep)
	numNeurons := len(snippets)

	b := Brain{
		DNA:            dna,
		neurons:        make([]Neuron, numNeurons),
		sigChan:        make(chan Signal),
		mu:             sync.Mutex{},
		pendingSignals: make(map[int][]SignalType, numNeurons),
	}

	for nIndex, snippet := range snippets {
		if snippint, err := strconv.Atoi(snippet); err == nil {
			b.neurons[nIndex] = Grow(snippint, numNeurons)
		} else {
			log.Fatalf("Unrecognized DNA snippet \"%s\" got err: %v", snippet, err)
		}
	}

	return &b
}

func (b *Brain) StepFunction() {
	firingNeurons := 0
	for nIndex, sigs := range b.pendingSignals {
		// A neuron only fires when it receives at least 2 inputs
		// Will need to question this assumption for vision neurons
		if len(sigs) >= 2 {
			go b.neurons[nIndex].Fire(b.sigChan, sigs)
			firingNeurons++
		}
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(map[int][]SignalType, len(b.neurons))

	for i := 0; i < firingNeurons; i++ {
		signal := <-b.sigChan
		for nIndex := range signal.nIndicies {
			b.pendingSignals[nIndex] = append(b.pendingSignals[nIndex], signal.val)
		}
	}

}
