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
	b.StepFunction()

	want := make(map[int][]SignalType, 2)
	want[1] = append(want[1], 15)

	if !reflect.DeepEqual(want, b.pendingSignals) {
		t.Errorf("Want equal, got %v and %v", want, b.pendingSignals)
	}
}
