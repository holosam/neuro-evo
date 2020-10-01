package neuron

import (
	"log"
	"strconv"
	"strings"
)

const dnaSep = "|"

// Brain docs
type Brain struct {
	DNA string

	neurons []Neuron

	signals chan Signal
}

func Flourish(dna string) *Brain {
	b := Brain{
		signals: make(chan Signal),
		DNA:     dna,
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

}

// Or, when the Brain does a step firing, all of the neurons use a channel to
// pass a signal and an index, so they don't need to keep references to other
// neurons, just holding numbers
