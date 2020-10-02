package env

import (
	"fmt"
	"hackathon/sam/evolve/neuron"
	"math"
)

func Adder() {
	play := neuron.NewPlayground()
	play.SeedRandDNA(50)

	inputs := []neuron.SignalType{5, 12}
	expectedResult := inputs[0] + inputs[1]

	acc := func(moves []neuron.SignalType) int {
		return int(math.Abs(float64(expectedResult-moves[0]))) + (10 * (len(moves) - 1))
	}

	bestDna := play.SimulatePlayground(100, inputs, acc)

	moves := neuron.FireBrainBlock(bestDna, inputs)

	fmt.Printf("Results: %v from %s\n", moves, bestDna.PrettyPrint())
}
