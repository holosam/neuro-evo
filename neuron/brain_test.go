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

func TestEyesight(t *testing.T) {
	// Two vision neurons pointing at a motor neuron.
	d := NewDNA()
	d.AddSnippet(ADD).AddSynapse(2)
	d.AddVisionId(0)

	d.AddSnippet(ADD).AddSynapse(2)
	d.AddVisionId(1)

	d.AddSnippet(ADD)
	d.AddMotorId(2)

	b := Flourish(d)

	b.SeeInput(1)

	// First step fires the vision neurons and pends for the motor neuron.
	moves := b.StepFunction()
	if got := len(moves); got != 0 {
		t.Errorf("Want 0, got %d", got)
	}

	// Second step fires the motor neuron.
	moves = b.StepFunction()
	if got := len(moves); got != 1 {
		t.Errorf("Want 1, got %d", got)
	}
	if got := moves[0]; got != 2 {
		t.Errorf("Want 2, got %d", got)
	}
}
