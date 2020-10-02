package env

import (
	"fmt"
	"hackathon/sam/evolve/neuron"
	"math"
)

func main() {
	play := neuron.NewPlayground()
	play.SeedRandDNA(100)

	inputs := []neuron.SignalType{5, 12}
	expectedResult := inputs[0] + inputs[1]

	acc := func(moves []neuron.SignalType) int {
		return int(math.Abs(float64(expectedResult-moves[0]))) + (10 * (len(moves) - 1))
	}

	bestDna := play.SimulatePlayground(10, inputs, acc)

	moves := neuron.FireBrainBlock(bestDna, inputs)

	fmt.Printf("Results: %v", moves)
}
