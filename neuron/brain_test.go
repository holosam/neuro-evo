package neuron

import (
	"reflect"
	"testing"
)

func TestBrainStep(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(ADD).AddSynapse(1)
	d.AddSnippet(FALSIFY).AddSynapse(0)
	b := Flourish(d)
	sigs := []SignalType{1, 2, 3, 4, 5}
	for _, sig := range sigs {
		b.pendingSignals[0] = append(b.pendingSignals[0], sig)
	}
	moves := b.StepFunction()

	if got := len(moves); got != 0 {
		t.Errorf("Want 0, got %d", got)
	}

	want := make(map[int][]SignalType, 1)
	want[1] = append(want[1], 15)

	if !reflect.DeepEqual(want, b.pendingSignals) {
		t.Errorf("Want %v, got %v", want, b.pendingSignals)
	}
}
