package neuron

import (
	"testing"
)

func TestFireBrain(t *testing.T) {
	resChan := make(chan BrainResult)
	go FireBrain(0, SimpleTestDNA(), []SignalType{1, 2}, 10, resChan)
	result := <-resChan

	// Seeing input doesn't count as a step.
	if got := result.steps; got != 1 {
		t.Errorf("Want 2, got %d", got)
	}
	if got := len(result.moves); got != 1 {
		t.Errorf("Want 1, got %d", got)
	} else {
		if result.moves[0] != 3 {
			t.Errorf("Want 3, got %v", result.moves[0])
		}
	}
}

func TestRunGeneration(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	codes[1] = SimpleTestDNA()

	results := RunGeneration(codes, []SignalType{1, 2}, 10)
	if got := results[1].moves; got[0] != 3 {
		t.Errorf("Want 3, got %v", got)
	}
}
