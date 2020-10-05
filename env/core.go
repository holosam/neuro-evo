package env

import (
	"fmt"
	"hackathon/sam/evolve/neuron"
)

type EnvironmentConfig struct {
	Pconf neuron.PlaygroundConfig

	NumPlaygrounds int

	GenInputsFn func() []neuron.SignalType
}

func DefaultEnvConfig() EnvironmentConfig {
	return EnvironmentConfig{
		Pconf: neuron.PlaygroundConfig{
			NumSpecies:            500,
			MaxGensPerPlay:        5,
			DnaSeedSnippets:       50,
			DnaSeedMutations:      0,
			MaxStepsPerGen:        50,
			ContinueAfterAccurate: false,
		},
		NumPlaygrounds: 1000,
	}
}

func RunEnvironment(econf EnvironmentConfig) *neuron.DNA {
	var dna *neuron.DNA

	// accurateSpecies := 0

	play := neuron.NewPlayground(econf.Pconf)
	play.SeedRandDNA()
	for p := 0; p < econf.NumPlaygrounds; p++ {

		// if dna != nil {
		// 	play.SeedKnownDNA(dna)
		// } else {
		// 	play.SeedRandDNA()
		// }

		inputs := econf.GenInputsFn()
		dna = play.SimulatePlayground(inputs)

		// Just used to show off the final result, can be deleted at some point
		moves := neuron.FireBrainBlock(dna, inputs, econf.Pconf.MaxStepsPerGen)
		// if econf.Pconf.AccuracyFn(econf.GenInputsFn(), moves) == 0 {
		// 	accurateSpecies++
		// 	if accurateSpecies == 5 {
		// 		econf.Pconf.ContinueAfterAccurate = true
		// 	}
		// }
		fmt.Printf("Play %d: move %v from %s\n", p, moves, dna.PrettyPrint())
	}

	return dna
}
