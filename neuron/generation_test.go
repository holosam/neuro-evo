package neuron

import (
	"reflect"
	"testing"
)

func TestFireBrain(t *testing.T) {
	codes := make(map[IDType]*DNA, 1)
	codes[0] = SimpleTestDNA()
	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)

	resChan := make(chan BrainResult)
	go g.fireBrain(0, []SignalType{1, 2}, resChan)
	gotRes := <-resChan

	wantRes := BrainResult{
		id:      0,
		inputs:  []SignalType{1, 2},
		Outputs: make([]Signal, 0),
		steps:   2,
	}

	if got, want := len(gotRes.Outputs), 1; got != want {
		t.Errorf("Want %v, got %v", want, got)
	} else {
		if got, want := len(gotRes.Outputs[0].sources), 2; got != want {
			t.Errorf("Want %v, got %v", want, got)
		}
		if got, want := gotRes.Outputs[0].output, SignalType(3); got != want {
			t.Errorf("Want %v, got %v", want, got)
		}
	}

	gotRes.Outputs = make([]Signal, 0)
	if !reflect.DeepEqual(wantRes, gotRes) {
		t.Errorf("Want %v, got %v", wantRes, gotRes)
	}
}

func TestRunGeneration(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	codes[1] = SimpleTestDNA()

	g := NewGeneration(GenerationConfig{MaxSteps: 10}, codes)
	results := g.FireBrains([]SignalType{1, 2})

	if got, want := results[1].Outputs, SignalType(3); got[0].output != want {
		t.Errorf("Want %d, got %v", want, got)
	}
}
