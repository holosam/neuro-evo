package env

import (
	"hackathon/sam/evolve/neuron"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func DefaultStockConfig() neuron.RunnerConfig {
	return neuron.RunnerConfig{
		Generations: 500,
		Rounds:      20,

		PConf: neuron.PlaygroundConfig{
			NumInputs:  1,
			NumOutputs: 2,

			NumVariants: 1000,

			Mconf: neuron.MutationConfig{
				NeuronExpansion:  0.1,
				SynapseExpansion: 0.1,

				AddNeuron:  0.1,
				AddSynapse: 0.2,

				ChangeOp:  0.3,
				SetSeed:   0.3,
				UnsetSeed: 0.3,
			},

			Econf: neuron.EvolutionConfig{
				Parents:                 3,
				BottomTierPercent:       0.25,
				DistanceThreshold:       0.5,
				DistanceEdgeFactor:      0.8,
				DistanceOperationFactor: 0.2,
			},
		},
	}
}

/*
type DayTrader struct {
	minute      int
	stockValues []neuron.SignalType
	rng         *rand.Rand

	money       int
	sharesOwned int
}

func (d *DayTrader) CurrentState() [][]neuron.SignalType {
	newVal := int(d.stockValues[d.minute-1]) + (5 - d.rng.Intn(11))
	if newVal <= 0 {
		newVal = 1
	}
	if newVal >= int(neuron.MaxSignal()) {
		newVal = int(neuron.MaxSignal())
	}
	d.stockValues[d.minute] = neuron.SignalType(newVal)
	return [][]neuron.SignalType{{d.stockValues[d.minute]}}
}

func (d *DayTrader) Update(signals [][]neuron.SignalType) {
	currentStockPrice := int(d.stockValues[d.minute])
	d.minute++

	if len(signals) != 2 {
		// Invalid output, make sure this isn't selected for.
		// If the entire generation's fitness sums to 0, then the species offspring
		// map will get a NaN and break. So min fitness should be >=10 (because
		// it's divided by species.Size())
		d.money = 100
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
	config := DefaultStockConfig()
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
*/

type RomanNumeral struct {
	input  int
	output []rune
	isOver bool
}

func (r *RomanNumeral) CurrentState() [][]neuron.SignalType {
	r.isOver = true
	runes := []rune(strconv.Itoa(r.input))
	input := make([]neuron.SignalType, len(runes))
	for i, r := range runes {
		input[i] = neuron.SignalType(r)
	}
	return [][]neuron.SignalType{input}
}

func (r *RomanNumeral) Update(signals [][]neuron.SignalType) {
	if len(signals) != 1 {
		return
	}

	for _, sig := range signals[0] {
		if sig > 0 {
			r.output = append(r.output, rune(sig))
		}
	}
}

func (r *RomanNumeral) IsOver() bool {
	return r.isOver
}

func (r *RomanNumeral) Fitness() neuron.ScoreType {
	answer := []rune(convert(r.input))
	score := 0

	for i, char := range answer {
		if len(r.output) > i {
			diff := int(char - r.output[i])
			score += diff * diff
		} else {
			score += 256 * 256
		}
	}

	return -neuron.ScoreType(score)
}

func convert(input int) string {
	conversions := []struct {
		val  int
		char string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var sb strings.Builder
	for _, conversion := range conversions {
		for input >= conversion.val {
			sb.WriteString(conversion.char)
			input -= conversion.val
		}
	}
	return sb.String()
}

func RomanNumeralConverter() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	config := DefaultStockConfig()
	config.PConf.NumInputs = 1
	config.PConf.NumOutputs = 1
	config.NewGameFn = func() neuron.Game {
		r := &RomanNumeral{
			input:  rng.Intn(3999),
			output: make([]rune, 0),
		}
		return r
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}
