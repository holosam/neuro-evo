package neuron

import (
	"fmt"
	"hackathon/sam/evolve/dynamo"
	"math"
	"time"
)

type ScoreType int64

// Game defines the methods needed to simulate a game.
type Game interface {
	// CurrentState is the state of the game represented by a series of signals.
	// In the future this should return a proto.Message
	CurrentState() [][]SignalType

	// Update changes the game state based on a series of moves.
	Update(signals [][]SignalType)

	IsOver() bool

	// Fitness scores the result of the game once it's over.
	Fitness() ScoreType
}

// NewGameFunc does any setup necessary to begin playing. Essentially a
// factory method/constructor to generate new games.
type NewGameFunc func() Game

type RunnerConfig struct {
	Generations int
	Rounds      int
	NewGameFn   NewGameFunc

	PConf PlaygroundConfig
}

type Runner struct {
	config RunnerConfig
	play   *Playground
}

func NewRunner(config RunnerConfig) *Runner {
	return &Runner{
		config: config,
		play:   NewPlayground(config.PConf),
	}
}

func (r *Runner) Run() {
	fmt.Printf("Beginning run with config: %+v\n", r.config)
	r.play.InitDNA()
	for gen := 0; gen < r.config.Generations; gen++ {
		fmt.Printf("\nGeneration #%d, starting at %v\n", gen, time.Now())
		if r.runGeneration(gen) {
			break
		}
	}

	dynScore := r.config.Generations * r.config.Rounds * r.play.config.NumVariants
	dynamo.Record("evolve", dynScore)
	fmt.Printf("Never found a winner :/\nDynamo result: %d\n", dynScore)
}

func (r *Runner) runGeneration(gen int) bool {
	results := make([]BrainScore, r.play.config.NumVariants)

	resChan := make(chan BrainScore)
	for round := 0; round < r.config.Rounds; round++ {
		// Simulate all games in separate goroutines.
		for id := 0; id < r.play.config.NumVariants; id++ {
			go r.gameSimulation(id, resChan)
		}

		// Wait for all the results to come in.
		for i := 0; i < r.play.config.NumVariants; i++ {
			result := <-resChan
			results[result.id].id = result.id
			results[result.id].score += result.score
		}
	}

	// If the max possible score has been reached, the simulation can end.
	maxResult := BrainScore{
		id:    -1,
		score: -math.MaxInt32,
	}
	for _, result := range results {
		if result.score > maxResult.score {
			maxResult = result
		}
	}
	bestDNA := r.play.codes[maxResult.id]
	fmt.Printf("Winner of generation:\n%sEnded with %d score\n\n", bestDNA.PrettyPrint(), maxResult.score)

	// For roman numerals: r.config.Rounds*256*256*7
	// For the healthchecker: r.config.Rounds*86400
	if maxResult.score == ScoreType(r.config.Rounds*256*256) { // For the adder.
		dynScore := gen * r.config.Rounds * r.play.config.NumVariants
		dynamo.Record("evolve", dynScore)
		fmt.Printf("We have a winner!\nDynamo result: %d\n", dynScore)
		return true
	}

	r.play.Evolve(results)
	return false
}

func (r *Runner) gameSimulation(id IDType, resChan chan BrainScore) {
	game := r.config.NewGameFn()
	brain := r.play.GetBrain(id)

	for !game.IsOver() {
		game.Update(brain.Fire(game.CurrentState()))
	}

	resChan <- BrainScore{
		id:    id,
		score: game.Fitness(),
	}
}
