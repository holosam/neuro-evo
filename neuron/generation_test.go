package neuron

import (
	"reflect"
	"testing"
)

func testGConf() GenerationConfig {
	return GenerationConfig{
		Rounds:   2,
		MaxSteps: 10,
		InputsFn: func(actions int) []SignalType {
			if actions >= 2 {
				return []SignalType{}
			}
			return []SignalType{SignalType(actions + 1), SignalType(actions + 2)}
		},
		FitnessFn: func(outputs []SignalType) ScoreType {
			return ScoreType(outputs[0])
		},
	}
}

func TestFireBrain(t *testing.T) {
	codes := make(map[IDType]*DNA, 1)
	codes[0] = SimpleTestDNA()
	g := NewGeneration(testGConf(), codes)

	resChan := make(chan BrainScore)
	go g.fireBrain(0, resChan)
	got := <-resChan

	want := BrainScore{
		id:    0,
		score: 6,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestRunGeneration(t *testing.T) {
	codes := make(map[int]*DNA, 2)
	codes[0] = SimpleTestDNA()
	codes[1] = SimpleTestDNA()

	g := NewGeneration(testGConf(), codes)
	results := g.FireBrains()

	want := make(map[IDType]BrainScore)
	want[0] = BrainScore{
		id:    0,
		score: 6,
	}
	want[1] = BrainScore{
		id:    1,
		score: 6,
	}

	if !reflect.DeepEqual(want, results) {
		t.Errorf("Want %v, got %v", want, results)
	}
}
