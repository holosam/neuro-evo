package env

import (
	"hackathon/sam/evolve/neuron"
	"testing"
)

func TestAdder(t *testing.T) {
	econf := DefaultEnvConfig()
	dna := Adder(econf)
	got := neuron.FireBrainBlock(dna, []neuron.SignalType{17, 13}, econf.Pconf.MaxStepsPerGen)
	if len(got) != 1 {
		t.Errorf("Want 1 move, got %d", len(got))
	} else {
		if got[0] != 30 {
			t.Errorf("Want 30, got %d", got[0])
		}
	}
}
