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
			NumSpecies:       250,
			NumGensPerPlay:   200,
			DnaSeedSnippets:  50,
			DnaSeedMutations: 20,
			WinnerRatio:      4,
			MaxStepsPerGen:   500,
		},
		NumPlaygrounds: 1,
	}
}

func RunEnvironment(econf EnvironmentConfig) *neuron.DNA {
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

	return dna
}
