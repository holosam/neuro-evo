package neuron

import (
	"math"
	"testing"
	"time"
)

func TestOperators(t *testing.T) {
	if got := ADD.operate(3, 7); got != 10 {
		t.Errorf("Got wrong ADD value: %v", got)
	}
	if got := MULTIPLY.operate(3, 7); got != 21 {
		t.Errorf("Got wrong MULTIPLY value: %v", got)
	}
	if got := AND.operate(9, 14); got != 8 {
		t.Errorf("Got wrong AND value: %v", got)
	}
	if got := NAND.operate(math.MaxUint8, math.MaxUint8-4); got != 4 {
		t.Errorf("Got wrong NAND value: %v", got)
	}
	if got := OR.operate(9, 10); got != 11 {
		t.Errorf("Got wrong OR value: %v", got)
	}
	if got := NOR.operate(math.MaxUint8-4, math.MaxUint8-4); got != 4 {
		t.Errorf("Got wrong NOR value: %v", got)
	}
	if got := XOR.operate(11, 12); got != 7 {
		t.Errorf("Got wrong XOR value: %v", got)
	}
	if got := IFF.operate(math.MaxUint8, 4); got != 4 {
		t.Errorf("Got wrong IFF value: %v", got)
	}
	if got := TRUTH.operate(3, 7); got != math.MaxUint8 {
		t.Errorf("Got wrong TRUTH value: %v", got)
	}
	if got := FALSIFY.operate(3, 7); got != 0 {
		t.Errorf("Got wrong FALSIFY value: %v", got)
	}
}

func CreateGenetic(op OperatorType) *Genetic {
	genetic := Genetic{
		op:             op,
		AbstractNeuron: AbstractNeuron{},
	}
	genetic.AbstractNeuron.Neuron = &genetic
	return &genetic
}

func TestSignalingPathway(t *testing.T) {
	upstream := CreateGenetic(ADD)
	downstream := CreateGenetic(ADD)

	upstream.downstream = append(upstream.downstream, downstream)

	upstream.Signal(5)
	upstream.MaybeFire()
	time.Sleep(time.Second)
	if len(downstream.pendingSignals) > 0 {
		t.Errorf("Neuron shouldn't have fired with 1 signal.")
	}

	upstream.Signal(7)
	upstream.MaybeFire()
	time.Sleep(time.Second)
	if len(downstream.pendingSignals) != 1 {
		t.Errorf("Neuron should have fired.")
	}
	if downstream.pendingSignals[0] != 12 {
		t.Errorf("Got wrong val.")
	}
}
