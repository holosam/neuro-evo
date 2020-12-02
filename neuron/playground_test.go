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
	if got := dna.NeuronIDs[SENSE].Length(); got != 1 {
		t.Errorf("Want 1, got %d", got)
	}
	if got := dna.NeuronIDs[MOTOR].Length(); got != 1 {
		t.Errorf("Want 1, got %d", got)
	}
}

func TestPathwayTraversal(t *testing.T) {
	parent := SimpleTestDNA()
	child := NewDNA()
	child.NeuronIDs[MOTOR].InsertID(10) // Necessary setup for this function.

	motorSignal := CreateTestSignal(3)
	motorSignal.neuronID = parent.NeuronIDs[MOTOR].GetId(0)

	visionID0 := parent.NeuronIDs[SENSE].GetId(0)
	visionSignal0 := CreateTestSignal(1)
	visionSignal0.neuronID = visionID0
	motorSignal.sources[visionID0] = visionSignal0

	visionSignal0.sources[-1] = CreateTestSignal(1)
	visionSignal0.sources[-2] = CreateTestSignal(0)

	visionID1 := parent.NeuronIDs[SENSE].GetId(1)
	visionSignal1 := CreateTestSignal(2)
	visionSignal1.neuronID = visionID1
	motorSignal.sources[visionID1] = visionSignal1

	visionSignal1.sources[-1] = CreateTestSignal(1)
	visionSignal1.sources[-2] = CreateTestSignal(0)

	p := NewPlayground(PlaygroundConfig{
		NumParents: 2,
	})

	p.randomTraversePathway(parent, child, motorSignal, -1)

	if dnaComplexity(child) >= dnaComplexity(parent) {
		t.Errorf("Parent %s should be more complex than child %s", parent.PrettyPrint(), child.PrettyPrint())
	}
	// Should have chosen 1 of the vision neurons to traverse.
	if got, want := child.NeuronIDs[SENSE].Length(), 1; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	// Could have 0 or 1 seeds depending on the traversal.
	if got, want := child.NeuronIDs[SENSE].Length(), 1; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSimulatePlayground(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		DnaSeedSnippets:  10,
		DnaSeedMutations: 10,

		NumSpecies:   10,
		Generations:  5,
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
		NumParents: 3,

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
