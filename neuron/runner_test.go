package neuron

import (
	"reflect"
	"testing"
)

type testGame struct {
	turn  int
	score ScoreType
}

func (t *testGame) CurrentState() [][]SignalType {
	return [][]SignalType{{SignalType(t.turn)}, {SignalType(t.turn + 1)}}
}

func (t *testGame) Update(signals [][]SignalType) {
	t.turn++
	if len(signals) > 0 && len(signals[0]) > 0 {
		t.score += ScoreType(signals[0][0])
	}
}

func (t *testGame) IsOver() bool {
	return t.turn >= 5
}

func (t *testGame) Fitness() ScoreType {
	return t.score
}

func createTestRunner() *Runner {
	return NewRunner(RunnerConfig{
		Generations: 3,
		Rounds:      2,
		NewGameFn: func() Game {
			return &testGame{
				turn:  1,
				score: 0,
			}
		},
		PConf: createTestPlayConfig(),
	})
}

func TestRunGeneration(t *testing.T) {
	runner := createTestRunner()
	runner.play.InitDNA()
	runner.runGeneration(0)
}

func TestGameSim(t *testing.T) {
	runner := createTestRunner()
	runner.play.InitDNA()
	runner.play.codes[0] = SimpleTestDNA()

	resChan := make(chan BrainScore)
	go runner.gameSimulation(0, resChan)

	result := <-resChan
	expected := BrainScore{
		id:    0,
		score: 18,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Got %+v, want %+v", result, expected)
	}
}
