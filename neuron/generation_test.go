package neuron

import (
	"testing"
)

func TestFireBrain(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)

	resChan := make(chan BrainResult)
	g.fireBrain(0, []SignalType{1, 2}, resChan)
	result := <-resChan

	// Seeing input doesn't count as a step.
	if got := result.steps; got != 1 {
		t.Errorf("Want 2, got %d", got)
	}
	if got := len(result.outputs); got != 1 {
		t.Errorf("Want 1, got %d", got)
	} else {
		if result.outputs[0] != 3 {
			t.Errorf("Want 3, got %v", result.outputs[0])
		}
	}
}

func TestRunGeneration(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	codes[1] = SimpleTestDNA()

	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)
	results := g.FireBrains([]SignalType{1, 2})

	if got := results[1].outputs; got[0] != 3 {
		t.Errorf("Want 3, got %v", got)
	}
}
