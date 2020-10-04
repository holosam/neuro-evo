package env

import (
	"hackathon/sam/evolve/neuron"
	"math"
	"math/rand"
	"time"
)

func Adder(econf EnvironmentConfig) *neuron.DNA {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// econf.GenInputsFn = func() []neuron.SignalType {
	// 	numInputs := 2 + rng.Intn(3)
	// 	inputs := make([]neuron.SignalType, numInputs)
	// 	for i := 0; i < numInputs; i++ {
	// 		inputs[i] = neuron.SignalType(rng.Intn((math.MaxUint8 / numInputs) - 1))
	// 	}
	// 	return inputs
	// }
	econf.GenInputsFn = func() []neuron.SignalType {
		inputs := make([]neuron.SignalType, 2)
		for i := 0; i < 2; i++ {
			inputs[i] = neuron.SignalType(rng.Intn(math.MaxUint8))
		}
		return inputs
	}

	econf.Pconf.AccuracyFn = func(inputs []neuron.SignalType, outputs []neuron.SignalType) int {
		expectedResult := neuron.SignalType(0)
		for _, sig := range inputs {
			expectedResult += sig
		}
		return int(math.Abs(float64(expectedResult-outputs[0]))) + (10 * (len(outputs) - 1))
	}

	return RunEnvironment(econf)
}
