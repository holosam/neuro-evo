package neuron

import (
	"math"
	"testing"
)

// Test some invariants after mutating DNA.
func TestRandDNA(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		NumSpecies:       1,
		DnaSeedSnippets:  10,
		DnaSeedMutations: 5,
		GenInputsFn: func(round int) []SignalType {
			return []SignalType{1}
		},
	})
	p.SeedRandDNA()
	dna := p.codes[0]

	if got := len(dna.Snippets); got < 1 {
		t.Errorf("Want at least 1, got %d", got)
	}
	if got := dna.VisionIDs.Length(); got != 1 {
		t.Errorf("Want 1, got %d", got)
	}
	if got := dna.MotorIDs.Length(); got != 1 {
		t.Errorf("Want 1, got %d", got)
	}
}

func TestResultScoring(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		FitnessFn: func(inputs []SignalType, outputs []SignalType) ScoreType {
			return ScoreType(outputs[0])
		},
	})

	targetID := 0
	p.codes[targetID] = SimpleTestDNA()

	score := p.scoreResult(targetID, &BrainResult{
		id:      targetID,
		inputs:  []SignalType{5, 20},
		Outputs: []SignalType{10, 6},
		steps:   20,
	}, []SignalType{})

	if got, want := score, ScoreType(102005); got != want {
		t.Errorf("Want %d, got %d", want, got)
	}
}

func TestSimulatePlayground(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		DnaSeedSnippets:  10,
		DnaSeedMutations: 10,

		NumSpecies:   10,
		Generations:  10,
		RoundsPerGen: 3,

		GenInputsFn: func(round int) []SignalType {
			return []SignalType{1, 2}
		},

		FitnessFn: func(inputs []SignalType, outputs []SignalType) ScoreType {
			if len(outputs) == 0 {
				return math.MaxUint64
			}
			return ScoreType(outputs[0])
		},
		NumSpeciesReproduce: 2,

		Gconf: GenerationConfig{
			MaxSteps: 10,
		},
	})
	p.SeedRandDNA()
	arbitaryDNA := p.codes[0]

	p.SimulatePlayground()

	if arbitaryDNA == p.codes[0] {
		t.Errorf("Expected evolution, got nothing")
	}
}
