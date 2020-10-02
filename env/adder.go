package env

import (
	"fmt"
	"hackathon/sam/evolve/neuron"
	"math"
	"math/rand"
	"time"
)

type EnvironmentConfig struct {
	pconf neuron.PlaygroundConfig

	NumPlaygrounds int

	genInputsFn func() []neuron.SignalType
}

func DefaultEnvConfig() EnvironmentConfig {
	return EnvironmentConfig{
		pconf: neuron.PlaygroundConfig{
			NumSpecies:       100,
			NumGensPerPlay:   10,
			DnaSeedSnippets:  10,
			DnaSeedMutations: 20,
			WinnerRatio:      2,
			MaxStepsPerGen:   200,
		},
		NumPlaygrounds: 20,
	}
}

func RunEnvironment(econf EnvironmentConfig) {
	var dna *neuron.DNA

	for p := 0; p < econf.NumPlaygrounds; p++ {
		play := neuron.NewPlayground(econf.pconf)

		if dna != nil {
			play.SeedKnownDNA(dna)
		} else {
			play.SeedRandDNA()
		}

		inputs := econf.genInputsFn()
		dna = play.SimulatePlayground(inputs)

		// Just used to show off the final result, can be deleted at some point
		moves := neuron.FireBrainBlock(dna, inputs, econf.pconf.MaxStepsPerGen)
		fmt.Printf("Play %d: move %v from %s\n", p, moves, dna.PrettyPrint())
	}
}

func Adder2() {
	econf := DefaultEnvConfig()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	econf.genInputsFn = func() []neuron.SignalType {
		numInputs := 2 + rng.Intn(3)
		inputs := make([]neuron.SignalType, numInputs)
		for i := 0; i < numInputs; i++ {
			inputs[i] = uint8(rng.Intn(math.MaxUint8))
		}
		return inputs
	}

	econf.pconf.AccuracyFn = func(inputs []neuron.SignalType, outputs []neuron.SignalType) int {
		expectedResult := uint8(0)
		for _, sig := range inputs {
			expectedResult += sig
		}
		return int(math.Abs(float64(expectedResult-outputs[0]))) + (10 * (len(outputs) - 1))
	}

	RunEnvironment(econf)
}

// func Adder(playgrounds int) {
// 	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

// 	var dna *neuron.DNA

// 	for p := 0; p < playgrounds; p++ {
// 		play := neuron.NewPlayground()

// 		if dna != nil {
// 			play.SeedKnownDNA(dna)
// 		} else {
// 			play.SeedRandDNA()
// 		}

// 		numInputs := 2 + rng.Intn(3)
// 		inputs := make([]neuron.SignalType, numInputs)
// 		expectedResult := uint8(0)
// 		for i := 0; i < numInputs; i++ {
// 			sig := uint8(rng.Intn(math.MaxUint8))
// 			inputs[i] = sig
// 			expectedResult += sig
// 		}

// 		acc := func(moves []neuron.SignalType) int {
// 			return int(math.Abs(float64(expectedResult-moves[0]))) + (10 * (len(moves) - 1))
// 		}

// 		dna = play.SimulatePlayground(numGensPerPlay, inputs, acc)

// 		moves := neuron.FireBrainBlock(dna, inputs)
// 		fmt.Printf("Play %d: move %v (expecting %d) from %s\n", p, moves, expectedResult, dna.PrettyPrint())
// 	}
// }
