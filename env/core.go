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
		play := neuron.NewPlayground(econf.Pconf)

		if dna != nil {
			play.SeedKnownDNA(dna)
		} else {
			play.SeedRandDNA()
		}

		inputs := econf.GenInputsFn()
		dna = play.SimulatePlayground(inputs)

		// Just used to show off the final result, can be deleted at some point
		moves := neuron.FireBrainBlock(dna, inputs, econf.Pconf.MaxStepsPerGen)
		fmt.Printf("Play %d: move %v from %s\n", p, moves, dna.PrettyPrint())
	}
}
