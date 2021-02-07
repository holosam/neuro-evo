package env

import (
	"hackathon/sam/evolve/neuron"
	"testing"
)

/*
func TestUpdateBoughtSold(t *testing.T) {
	d := &DayTrader{
		minute:      1,
		stockValues: make([]neuron.SignalType, 10),
		money:       100,
		sharesOwned: 3,
	}
	d.stockValues[1] = 30

	// Buy 2, sell 1.
	d.Update([]neuron.SignalType{2, 1})
	if got, want := d.minute, 2; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := d.money, 70; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := d.sharesOwned, 4; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	d.minute = 1
	d.Update([]neuron.SignalType{10, 1})
	// Only have money to buy 2, before selling 1.
	if got, want := d.money, 40; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := d.sharesOwned, 5; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	d.minute = 1
	d.Update([]neuron.SignalType{1, 10})
	// Only have 6 shares to sell.
	if got, want := d.money, 190; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := d.sharesOwned, 0; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestDayTrader(t *testing.T) {
	StockSimulation()
	t.Errorf("always error to read logs")
}
*/

func TestAdderFitness(t *testing.T) {
	a := &Adder{
		inputs: [][]neuron.SignalType{{3, 4}, {5, 6}}, // sum: 18
	}
	a.CurrentState()

	a.Update([][]neuron.SignalType{})
	if got, want := a.output, neuron.SignalType(0); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := a.Fitness(), neuron.ScoreType(0); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	a.Update([][]neuron.SignalType{{48, 2}})
	if got, want := a.output, neuron.SignalType(48); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := a.Fitness(), neuron.ScoreType(256*256-30*30); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	a.Update([][]neuron.SignalType{{18}})
	if got, want := a.Fitness(), neuron.ScoreType(256*256); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestAdder(t *testing.T) {
	RunAdder()
	t.Errorf("always error to read logs")
}

func TestRomanNumeralConversion(t *testing.T) {
	testcases := []struct {
		input  int
		output string
	}{
		{0, ""},
		{1, "I"},
		{4, "IV"},
		{39, "XXXIX"},
		{246, "CCXLVI"},
		{789, "DCCLXXXIX"},
		{1009, "MIX"},
		{2421, "MMCDXXI"},
		{3888, "MMMDCCCLXXXVIII"},
	}

	for _, testcase := range testcases {
		if got, want := convert(testcase.input), testcase.output; got != want {
			t.Errorf("Got %v, want %v", got, want)
		}
	}
}

func TestRomanNumeralFitness(t *testing.T) {
	r := &RomanNumeral{
		input:  246,
		output: []rune{'C', 'C', 'X', 'M', 'T'},
	}
	expected := neuron.ScoreType(-(1 + 4 + 65536))

	if got := r.Fitness(); got != expected {
		t.Errorf("Got %v, want %v", got, expected)
	}
}

func TestNumeralConversion(t *testing.T) {
	RomanNumeralConverter()
	t.Errorf("always error to read logs")
}
