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
	b := Brain{
		sigChan:        make(chan Signal),
		DNA:            dna,
		pendingSignals: make(map[int][]SignalType),
	}

	snippets := strings.Split(dna, dnaSep)
	for _, snippet := range snippets {
		if snippint, err := strconv.Atoi(snippet); err != nil {
			b.neurons = append(b.neurons, Grow(snippint, len(snippets)))
		} else {
			log.Fatalf("Unrecognized DNA snippet (%s) got err: %v", snippet, err)
		}
	}

	return &b
}

func (b *Brain) StepFunction() {
	firingNeurons := 0
	for nIndex, sigs := range b.pendingSignals {
		if len(sigs) >= 2 {
			go b.neurons[nIndex].Fire(b.sigChan, sigs)
			firingNeurons++
		}
	}
	b.pendingSignals = make(map[int][]SignalType)

	for i := 0; i < firingNeurons; i++ {
		signal := <-b.sigChan
		for nIndex := range signal.nIndicies {
			b.pendingSignals[nIndex] = append(b.pendingSignals[nIndex], signal.val)
		}
	}

}
