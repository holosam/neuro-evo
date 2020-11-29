package env

import (
	"hackathon/sam/evolve/neuron"
	"math"
	"math/rand"
	"time"
)

type EnvironmentConfig struct {
	Pconf neuron.PlaygroundConfig
}

func DefaultEnvConfig() EnvironmentConfig {
	return EnvironmentConfig{
		Pconf: neuron.PlaygroundConfig{
			DnaSeedSnippets:  20,
			DnaSeedMutations: 5,

			NumSpecies:   1000,
			Generations:  500,
			RoundsPerGen: 8,

			NumParents: 3,

			Gconf: neuron.GenerationConfig{
				// Could try to return early when there are no pending signals.
				MaxSteps: 50,
			},
		},
	}
}

func Adder(econf EnvironmentConfig) *neuron.DNA {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	econf.Pconf.GenInputsFn = func(round int) []neuron.SignalType {
		inputs := make([]neuron.SignalType, 2)
		for i := 0; i < 2; i++ {
			// inputs[i] = neuron.SignalType(rng.Intn(int(neuron.MaxSignal())))
			inputs[i] = neuron.SignalType(rng.Intn(8))
			// inputs[i] = neuron.SignalType(rng.Intn(4))
		}
		return inputs
	}

	econf.Pconf.FitnessFn = func(inputs []neuron.SignalType, outputs []neuron.SignalType) neuron.ScoreType {
		// if len(outputs) != 1 {
		// 	return math.MaxUint32
		// }
		if len(outputs) == 0 {
			return math.MaxUint32
		}
		expectedResult := neuron.SignalType(0)
		for _, sig := range inputs {
			expectedResult += sig
		}

		return neuron.ScoreType(math.Pow(float64(expectedResult-outputs[0]), 2))
		// return int(math.Abs(float64(expectedResult-outputs[0]))) + (10 * (len(outputs) - 1))

		// if outputs[0] == expectedResult {
		// 	return 0
		// } else {
		// 	return math.MaxInt16
		// }
	}

	play := neuron.NewPlayground(econf.Pconf)
	play.SeedRandDNA()
	play.SimulatePlayground()

	return play.GetWinner()
}
