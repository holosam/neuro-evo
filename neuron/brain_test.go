package neuron

import (
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	d := NewDNA()
	d.AddSnippet(2).AddSynapse(2)
	d.AddVisionId(0)

	d.AddSnippet(2).AddSynapse(2)
	d.AddVisionId(1)

	d.AddSnippet(2)
	d.AddMotorId(2)

	return d
}

func TestBrainStep(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(2).AddSynapse(1)
	d.AddSnippet(7).AddSynapse(0)
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
	want[1] = append(want[1], 7)
	if !reflect.DeepEqual(want, b.pendingSignals) {
		t.Errorf("Want %v, got %v", want, b.pendingSignals)
	}
}

func TestEyesight(t *testing.T) {
	b := Flourish(SimpleTestDNA())
	b.SeeInput(3)

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
	if got := moves[0]; got != 3 {
		t.Errorf("Want 3, got %d", got)
	}
}
