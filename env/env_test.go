package env

import (
	"fmt"
	"hackathon/sam/evolve/neuron"
	"testing"
)

func TestAdder(t *testing.T) {
	econf := DefaultEnvConfig()
	dna := Adder(econf)
	fmt.Printf("Final DNA: %s\n", dna.PrettyPrint())

	codes := make(map[neuron.IDType]*neuron.DNA, 1)
	codes[0] = dna

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			gen := neuron.NewGeneration(neuron.GenerationConfig{
				MaxSteps: econf.Pconf.Gconf.MaxSteps,
			}, codes)

			results := gen.FireBrains([]neuron.SignalType{neuron.SignalType(i), neuron.SignalType(j)})
			got := results[0].Outputs

			fmt.Printf("%d + %d = %d|%d\n", i, j, neuron.SignalType(i+j), got[0].Output)
		}
	}

	t.Errorf("always error for now")
}
