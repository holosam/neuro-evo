package env

import (
	"hackathon/sam/evolve/neuron"
	"testing"
)

func TestAdder(t *testing.T) {
	econf := DefaultEnvConfig()
	dna := Adder(econf)

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			got := neuron.FireBrainBlock(dna, []neuron.SignalType{neuron.SignalType(i), neuron.SignalType(j)}, econf.Pconf.MaxStepsPerGen)
			if len(got) != 1 {
				t.Errorf("Want 1 move, got %d", len(got))
			} else {
				want := neuron.SignalType(i + j)
				if want != got[0] {
					t.Errorf("%d + %d = %d, got %d", i, j, want, got[0])
				}
			}
		}
	}
}
