package env

import (
	"hackathon/sam/evolve/neuron"
	"math/rand"
	"time"
)

type EnvironmentConfig struct {
	Pconf neuron.PlaygroundConfig
}

func DefaultEnvConfig() EnvironmentConfig {
	return EnvironmentConfig{
		Pconf: neuron.PlaygroundConfig{
			// NumInputs:  2,
			// NumOutputs: 1,

			NumVariants: 100,
			Generations: 10,

			Mconf: neuron.MutationConfig{
				NeuronExpansion:  0.20,
				SynapseExpansion: 0.30,

				AddNeuron:  0.2,
				AddSynapse: 0.3,

				ChangeOp:  0.5,
				SetSeed:   0.1,
				UnsetSeed: 0.1,
			},

			Econf: neuron.EvolutionConfig{
				Parents:           3,
				BottomTierPercent: 0.25,
				DistanceThreshold: 0.25,
			},

			Gconf: neuron.GenerationConfig{
				Rounds: 5,
				// Could try to return early when there are no pending signals.
				MaxSteps: 30,

				// InputsFn
				// FitnessFn
			},
		},
	}
}

/*
func SortList(econf EnvironmentConfig) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	econf.Pconf.NumInputs = 5
	econf.Pconf.NumOutputs = 5

	econf.Pconf.Gconf.InputsFn = func(action int) []neuron.SignalType {
		if action >= 1 {
			return make([]neuron.SignalType, 0)
		}
		return make([]neuron.SignalType)
	}

}
*/

func DayTrader(econf EnvironmentConfig) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	tradingWindowMinutes := 390
	stockValues := make([]neuron.SignalType, tradingWindowMinutes)
	stockValues[0] = neuron.MaxSignal() / 2

	minute := 1
	add := false
	for {
		add = rng.Float32() < 0.50
		length := rng.Intn(10) + 1

		for i := 0; i < length; i++ {
			if minute == tradingWindowMinutes {
				break
			}
			minute++

			changeVal := neuron.SignalType(rng.Intn(3) + 1)
			if add {
				stockValues[minute] = stockValues[minute-1] + changeVal
			} else {
				stockValues[minute] = stockValues[minute-1] - changeVal
			}
		}

		if minute >= tradingWindowMinutes {
			break
		}
	}

	econf.Pconf.Gconf.InputsFn = func(action int) []neuron.SignalType {
		inputs := make([]neuron.SignalType, 0)
		if action <= tradingWindowMinutes {
			inputs = append(inputs, stockValues[action])
		}
		return inputs
	}

	play := neuron.NewPlayground(econf.Pconf)
	play.InitDNA()
	play.SimulatePlayground()
}
