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
		Generations: 1000,
		Rounds:      10,

		PConf: neuron.PlaygroundConfig{
			NumInputs:  1,
			NumOutputs: 2,

			NumVariants: 2000,

			Mconf: neuron.MutationConfig{
				AddNeuron:  0.03,
				AddSynapse: 0.1,

				ChangeOp:  0.5,
				SetSeed:   0.4,
				UnsetSeed: 0.4,
			},

			Econf: neuron.EvolutionConfig{
				Parents:                 3,
				BottomTierPercent:       0.25,
				DistanceThreshold:       0.5,
				DistanceEdgeFactor:      0.7,
				DistanceOperationFactor: 0.3,
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

type Adder struct {
	inputs [][]neuron.SignalType
	answer neuron.SignalType
	output neuron.SignalType
	isOver bool
}

func (a *Adder) CurrentState() [][]neuron.SignalType {
	a.isOver = true

	for i := 0; i < len(a.inputs); i++ {
		for ii := 0; ii < len(a.inputs[i]); ii++ {
			a.answer += a.inputs[i][ii]
		}
	}

	return a.inputs
}

func (a *Adder) Update(signals [][]neuron.SignalType) {
	if len(signals) == 0 || len(signals[0]) == 0 {
		a.output = 0
		return
	}
	a.output = signals[0][0]
}

func (a *Adder) IsOver() bool {
	return a.isOver
}

func (a *Adder) Fitness() neuron.ScoreType {
	var diff int
	if a.output == 0 {
		diff = 256
	} else {
		diff = int(a.output) - int(a.answer)
	}
	return neuron.ScoreType((256 * 256) - (diff * diff))
}

func RunAdder() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	config := DefaultStockConfig()
	config.PConf.NumInputs = 2
	config.PConf.NumOutputs = 1
	config.NewGameFn = func() neuron.Game {
		a := &Adder{
			inputs: make([][]neuron.SignalType, 2),
		}

		for i := 0; i < 2; i++ {
			for ii := 0; ii < 2; ii++ {
				a.inputs[i] = append(a.inputs[i], neuron.SignalType(rng.Intn(63)+1))
			}
		}
		return a
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}

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

	// This scoring function punishes longer answers more, which isn't ideal.
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
		return &RomanNumeral{
			input:  rng.Intn(99), //3999),
			output: make([]rune, 0),
		}
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}
