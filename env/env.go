package env

import (
	"hackathon/sam/evolve/neuron"
	"math/rand"
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

			Econf: neuron.EvolutionConfig{
				Parents:           3,
				BottomTierPercent: 0.2,

				DistanceThreshold:       0.5,
				DistanceEdgeFactor:      0.7,
				DistanceOperationFactor: 0.3,
			},

			Mconf: neuron.MutationConfig{
				AddNeuron:  0.02,
				AddSynapse: 0.05,

				ChangeOp:  0.5,
				SetSeed:   0.4,
				UnsetSeed: 0.4,
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
	} else {
		a.output = signals[0][0]
	}
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
	// runes := []rune(strconv.Itoa(r.input))
	// input := make([]neuron.SignalType, len(runes))
	// for i, r := range runes {
	// 	input[i] = neuron.SignalType(r)
	// }
	// return [][]neuron.SignalType{input}
	return [][]neuron.SignalType{{neuron.SignalType(r.input)}}
}

func (r *RomanNumeral) Update(signals [][]neuron.SignalType) {
	if len(signals) != 1 {
		return
	}

	for _, sig := range signals[0] {
		r.output = append(r.output, rune(sig))
	}
}

func (r *RomanNumeral) IsOver() bool {
	return r.isOver
}

func (r *RomanNumeral) Fitness() neuron.ScoreType {
	answer := []rune(convert(r.input))
	score := 256 * 256 * 7

	// This scoring function punishes longer answers more: for XXXVIII, the
	// score gets added to 7 times.
	for i, char := range answer {
		if i < len(r.output) {
			diff := int(char - r.output[i])
			score -= diff * diff
		} else {
			score -= 256 * 256
		}
	}

	return neuron.ScoreType(score)
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
			input:  rng.Intn(40), //3999),
			output: make([]rune, 0),
		}
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}

/*
type VisualCortexAdder struct {
	values []int
	answer int
	output []rune
	isOver bool
}

func (a *VisualCortexAdder) CurrentState() [][]neuron.SignalType {
	a.isOver = true

	inputs := make([][]neuron.SignalType, 16)
	for i, value := range a.values {
		a.answer += value
		inputs[i] = make([]neuron.SignalType, len(a.values))
		for ii, hashVal := range md5.Sum([]byte(strconv.Itoa(value))) {
			// I'd rather have this be more bloom-filtery, where only a subset of
			// the vision neurons receive input, but this is just an experiment
			// anyway.
			inputs[i][ii] = hashVal
		}
	}

	return inputs
}

func (a *VisualCortexAdder) Update(signals [][]neuron.SignalType) {
	if len(signals) != 5 {
		a.output = make([]rune, 0)
		return
	}

	a.output = make([]rune, 5)
	for i, sig := range signals {
		a.output[i] = rune(sig[0])
	}
}

func (a *VisualCortexAdder) IsOver() bool {
	return a.isOver
}

func (a *VisualCortexAdder) Fitness() neuron.ScoreType {
	if len(a.output) == 0 {
		return 1
	}

	expected := []rune(strconv.Itoa(a.answer))
	if len(expected) < 5 {
		zeroPadding := make([]rune, 5-len(expected))

		// pick up here
		expected = append(zeroPadding, expected...)
	}
	for i := 0; i < len(a.output); i++ {

	}

	return neuron.ScoreType((256 * 256) - (diff * diff))
}

func RunVisualCortexAdder() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	config := DefaultStockConfig()
	config.PConf.NumInputs = 16
	config.PConf.NumOutputs = 5
	config.NewGameFn = func() neuron.Game {
		a := &VisualCortexAdder{
			values: make([]int, rng.Intn(8)+2),
		}
		for i := 0; i < len(a.values); i++ {
			a.values[i] = rng.Intn(1000)
		}
		return a
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}
*/

type HealthChecker struct {
	second int
	isUp   bool
	uptime int

	countdown int
	recovery  bool

	checks int
}

func (h *HealthChecker) CurrentState() [][]neuron.SignalType {
	state := neuron.MaxSignal()
	if !h.isUp {
		// Can't put in 0 because that's the NullRune.
		state = neuron.MaxSignal() / 2
	}

	return [][]neuron.SignalType{{state}}
}

func (h *HealthChecker) Update(signals [][]neuron.SignalType) {
	// This is an error state, where there are either not enough or too many
	// motor neurons.
	if len(signals) != 1 || len(signals[0]) == 0 {
		h.uptime = 1
		h.second = 86400
		return
	}

	for _, signal := range signals[0] {
		sleepTime := int(signal)
		if sleepTime == 0 {
			sleepTime = 1
		}

		for i := 0; i < sleepTime; i++ {
			// If the server is in recovery mode, wait the countdown then restart.
			if h.recovery {
				h.countdown--
				if h.countdown == 0 {
					h.isUp = true
					h.recovery = false
				}
			}

			if h.isUp {
				h.uptime++
			}

			h.second++
			if h.IsOver() {
				return
			}

			// Server doesn't have 100% uptime.
			if h.second%313 == 0 {
				h.isUp = false
				h.countdown = 5
			}
		}

		// After the sleep time, a health check is executed.
		h.checks++

		// Uptime isn't accumulated until the first health check sees it's up.
		if h.checks == 1 {
			h.isUp = true
		}

		// If the server went down, it can't come back up until 5 seconds after the
		// first health check to notice it's down.
		if !h.isUp {
			h.recovery = true
		}

		// Chance of server going down if sleep cadence is too short.
		// Can't recover until the health checks back off.
		if h.checks >= 20 && h.second/h.checks < 3 {
			h.isUp = false
			h.recovery = false
		}
	}
}

func (h *HealthChecker) IsOver() bool {
	return h.second >= 86400
}

func (h *HealthChecker) Fitness() neuron.ScoreType {
	return neuron.ScoreType(h.uptime)
}

func RunHealthChecker() {
	config := DefaultStockConfig()

	config.Rounds = 1
	config.PConf.NumInputs = 1
	config.PConf.NumOutputs = 1

	config.NewGameFn = func() neuron.Game {
		return &HealthChecker{}
	}

	runner := neuron.NewRunner(config)
	runner.Run()
}
