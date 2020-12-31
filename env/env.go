package env

import (
	"hackathon/sam/evolve/neuron"
	"math/rand"
	"time"
)

func DefaultRunnerConfig() neuron.RunnerConfig {
	return neuron.RunnerConfig{
		Generations: 100,
		Rounds:      3,

		PConf: neuron.PlaygroundConfig{
			NumInputs:  1,
			NumOutputs: 2,

			NumVariants: 500,

			Mconf: neuron.MutationConfig{
				NeuronExpansion:  0.2,
				SynapseExpansion: 0.3,

				AddNeuron:  0.2,
				AddSynapse: 0.3,

				ChangeOp:  0.1,
				SetSeed:   0.1,
				UnsetSeed: 0.1,
			},

			Econf: neuron.EvolutionConfig{
				Parents:                 3,
				BottomTierPercent:       0.25,
				DistanceThreshold:       0.37,
				DistanceEdgeFactor:      0.8,
				DistanceOperationFactor: 0.2,
			},
		},
	}
}

type DayTrader struct {
	minute      int
	stockValues []neuron.SignalType
	rng         *rand.Rand

	money       int
	sharesOwned int
}

func (d *DayTrader) CurrentState() []neuron.SignalType {
	newVal := int(d.stockValues[d.minute-1]) + (5 - d.rng.Intn(11))
	if newVal <= 0 {
		newVal = 1
	}
	if newVal >= int(neuron.MaxSignal()) {
		newVal = int(neuron.MaxSignal())
	}
	d.stockValues[d.minute] = neuron.SignalType(newVal)
	return []neuron.SignalType{d.stockValues[d.minute]}
}

func (d *DayTrader) Update(signals []neuron.SignalType) {
	currentStockPrice := int(d.stockValues[d.minute])
	d.minute++

	if len(signals) != 2 {
		// Invalid output, make sure this isn't selected for.
		d.money = 0
		d.sharesOwned = 0
		return
	}

	sharesToBuy := int(signals[0])
	moneyToSpend := 0
	for i := 0; i < sharesToBuy; i++ {
		if d.money-currentStockPrice < moneyToSpend {
			break
		}
		moneyToSpend += currentStockPrice
	}
	sharesBought := moneyToSpend / currentStockPrice
	d.money -= moneyToSpend
	d.sharesOwned += sharesBought

	sharesToSell := int(signals[1])
	sharesSold := sharesToSell
	if sharesSold >= d.sharesOwned {
		sharesSold = d.sharesOwned
	}
	d.money += sharesSold * currentStockPrice
	d.sharesOwned -= sharesSold
}

func (d *DayTrader) IsOver() bool {
	return d.minute >= 250
}

func (d *DayTrader) Fitness() neuron.ScoreType {
	return neuron.ScoreType(d.money)
}

func StockSimulation() {
	config := DefaultRunnerConfig()
	config.NewGameFn = func() neuron.Game {
		d := &DayTrader{
			minute:      1,
			stockValues: make([]neuron.SignalType, 250),
			rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
			money:       1000,
			sharesOwned: 0,
		}
		d.stockValues[0] = neuron.MaxSignal() / 2
		return d
	}
	runner := neuron.NewRunner(config)
	runner.Run()
}
