package env

import (
	"hackathon/sam/evolve/neuron"
	"testing"
)

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
	t.Errorf("always error for now")
}
