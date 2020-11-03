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
			DnaSeedSnippets:  30,
			DnaSeedMutations: 10,

			NumSpecies:   1000,
			Generations:  500,
			RoundsPerGen: 10,

			NumSpeciesReproduce: 20,

			Gconf: neuron.GenerationConfig{
				MaxSteps: 100,
			},
		},
	}
}

func Adder(econf EnvironmentConfig) *neuron.DNA {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	econf.Pconf.GenInputsFn = func(round int) []neuron.SignalType {
		inputs := make([]neuron.SignalType, 2)
		for i := 0; i < 2; i++ {
			// inputs[i] = neuron.SignalType(rng.Intn(math.MaxUint8))
			inputs[i] = neuron.SignalType(rng.Intn(8))
		}
		return inputs
	}

	econf.Pconf.FitnessFn = func(inputs []neuron.SignalType, outputs []neuron.SignalType) neuron.ScoreType {
		if len(outputs) != 1 {
			return math.MaxUint32
		}
		expectedResult := neuron.SignalType(0)
		for _, sig := range inputs {
			expectedResult += sig
		}
		return neuron.ScoreType(math.Abs(float64(expectedResult - outputs[0])))
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
