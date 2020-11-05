package neuron

import (
	"reflect"
	"testing"
)

func TestFireBrain(t *testing.T) {
	codes := make(map[int]*DNA, 1)
	codes[0] = SimpleTestDNA()
	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)

	resChan := make(chan BrainResult)
	go g.fireBrain(0, []SignalType{1, 2}, resChan)
	got := <-resChan

	want := BrainResult{
		id:      0,
		inputs:  []SignalType{1, 2},
		Outputs: make([]Signal, 1),
		steps:   2,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Want %v, got %v", want, got)
	}
}

/*
func TestRunGeneration(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	codes[1] = SimpleTestDNA()

	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)
	results := g.FireBrains([]SignalType{1, 2})

	if got, want := results[1].Outputs, SignalType(3); got[0] != want {
		t.Errorf("Want %d, got %d", want, got)
	}
}
*/
