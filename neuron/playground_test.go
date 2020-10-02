package neuron

import (
	"reflect"
	"testing"
)

// Test some invariants after mutating DNA.
func TestRandDNA(t *testing.T) {
	p := NewPlayground()
	p.SeedRandDNA(1)
	dna := p.codes[0]

	if got := len(dna.snips); got < 1 {
		t.Errorf("Want at least 1, got %d", got)
	}
	if got := len(dna.visionIDs); got < 1 {
		t.Errorf("Want at least 1, got %d", got)
	}
	if got := len(dna.motorIDs); got < 1 {
		t.Errorf("Want at least 1, got %d", got)
	}
}

func TestResultScoring(t *testing.T) {
	p := NewPlayground()

	targetID := 0
	p.codes[targetID] = SimpleTestDNA()

	score := p.scoreResult(targetID, &BrainResult{
		id:    targetID,
		moves: []SignalType{10, 6},
		steps: 20,
	}, func(moves []SignalType) int {
		return int(moves[0])
	})

	if got := score.score; got != 10020005 {
		t.Errorf("Want 10020005, got %d", got)
	}
}

func TestNextGenCodes(t *testing.T) {
	numSpecies := 10
	p := NewPlayground()
	p.SeedRandDNA(numSpecies)

	scores := make([]SpeciesScore, numSpecies)
	for i := 0; i < numSpecies; i++ {
		scores[i] = SpeciesScore{i, i * 10}
	}

	want := make(map[int]*DNA, numSpecies)
	want[0] = p.codes[0]
	want[1] = p.codes[0]
	want[2] = p.codes[0]
	want[3] = p.codes[0]
	want[4] = p.codes[0]
	want[5] = p.codes[1]
	want[6] = p.codes[1]
	want[7] = p.codes[1]
	want[8] = p.codes[2]
	want[9] = p.codes[3]

	p.setNextGenCodes(scores)

	if !reflect.DeepEqual(want, p.codes) {
		t.Errorf("Want %v, got %v", want, p.codes)
	}
}

func TestSimulatePlayground(t *testing.T) {
	p := NewPlayground()
	p.SeedRandDNA(10)
	arbitaryDNA := p.codes[0]

	p.SimulatePlayground(10, []SignalType{1, 2}, func(moves []SignalType) int {
		return int(moves[0])
	})

	if arbitaryDNA == p.codes[0] {
		t.Errorf("Expected evolution, got nothing")
	}
}
