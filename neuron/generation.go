package neuron

import "fmt"

// GenerateInputsFunc returns a set of inputs, which is called when needed
// (after each output). The actions param indicate how many outputs have
// It's the responsibility of the owner of the function to
// track the progress of the session and input values accordingly.
// When this function returns an empty list, the simulation is ended.
type GenerateInputsFunc func(actions int) []SignalType

type ScoreType uint64

// FitnessFunc scores the output based on its effect on the session.
// It's only called when the brain has a full output, so the outputs param
// doesn't need to be checked for unset values.
type FitnessFunc func(outputs []SignalType) ScoreType

type BrainScore struct {
	// The ID is included in this struct since it's passed on channels
	// and sorted so the score needs to travel with the ID.
	id    IDType
	score ScoreType
}

type GenerationConfig struct {
	Rounds int
	// Just a sanity check to prevent brains from going on forever.
	MaxSteps int

	InputsFn  GenerateInputsFunc
	FitnessFn FitnessFunc
}

// Generation handles all of the brain simulations.
type Generation struct {
	config GenerationConfig
	codes  map[IDType]*DNA
}

func NewGeneration(gconf GenerationConfig, codes map[IDType]*DNA) *Generation {
	return &Generation{
		config: gconf,
		codes:  codes,
	}
}

func (g *Generation) FireBrains() map[IDType]BrainScore {
	scores := make(map[IDType]BrainScore, len(g.codes))

	for round := 0; round < g.config.Rounds; round++ {
		// Simulate all brains in separate goroutines.
		scoreChan := make(chan BrainScore)
		for id := range g.codes {
			go g.fireBrain(id, scoreChan)
		}

		// Wait for all the results to come in before ending the round.
		for i := 0; i < len(g.codes); i++ {
			result := <-scoreChan
			scores[result.id] = result
		}
	}

	return scores
}

func (g *Generation) fireBrain(id IDType, scoreChan chan BrainScore) {
	// Brains keep states between actions but reset each round.
	brain := Flourish(g.codes[id])
	score := ScoreType(0)

	actions := 0
	// Let's get this party started.
	brain.SeeInput(g.config.InputsFn(actions))

	for step := 1; step <= g.config.MaxSteps; step++ {
		fmt.Printf("running brain step %d\n", step)
		if !brain.StepFunction() {
			continue
		}
		score += g.config.FitnessFn(brain.Output())

		actions++
		inputs := g.config.InputsFn(actions)
		if len(inputs) == 0 {
			break
		}
		brain.SeeInput(inputs)
	}

	scoreChan <- BrainScore{
		id:    id,
		score: score,
	}
}
