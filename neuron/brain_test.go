package neuron

import (
	"reflect"
	"testing"
)

func TestCreateBrain(t *testing.T) {
	b := Flourish("10|27|13")
	if got := len(b.neurons); got != 3 {
		t.Errorf("Want 3, got %d", got)
	}

	// These neurons should be equivalent, so ensure they fire the same way.
	sigs := []SignalType{1, 2, 3, 4, 5}
	g3 := Grow(13, 3) // The third neuron in the DNA string above.
	go g3.Fire(b.sigChan, sigs)
	sig1 := <-b.sigChan

	go b.neurons[2].Fire(b.sigChan, sigs)
	sig2 := <-b.sigChan

	if sig1.val != sig2.val {
		t.Errorf("Want equal, got %d and %d", sig1.val, sig2.val)
	}
	if !reflect.DeepEqual(sig1.nIndicies, sig2.nIndicies) {
		t.Errorf("Want equal, got %v and %v", sig1.nIndicies, sig2.nIndicies)
	}
}

func TestBrainStep(t *testing.T) {
	b := Flourish("11|10")
	sigs := []SignalType{1, 2, 3, 4, 5}
	for _, sig := range sigs {
		b.pendingSignals[0] = append(b.pendingSignals[0], sig)
	}
	b.StepFunction()

	want := make(map[int][]SignalType, 2)
	want[1] = append(want[1], 120)

	if !reflect.DeepEqual(want, b.pendingSignals) {
		t.Errorf("Want equal, got %v and %v", want, b.pendingSignals)
	}
}
