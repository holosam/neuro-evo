package neuron

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestOperators(t *testing.T) {
	if got := AND.operate(9, 14); got != 8 {
		t.Errorf("Got wrong AND value: %v", got)
	}
	if got := NAND.operate(MaxSignal(), MaxSignal()-4); got != 4 {
		t.Errorf("Got wrong NAND value: %v", got)
	}
	if got := OR.operate(9, 10); got != 11 {
		t.Errorf("Got wrong OR value: %v", got)
	}
	if got := NOR.operate(MaxSignal()-4, MaxSignal()-4); got != 4 {
		t.Errorf("Got wrong NOR value: %v", got)
	}
	if got := XOR.operate(11, 12); got != 7 {
		t.Errorf("Got wrong XOR value: %v", got)
	}
	if got := IFF.operate(MaxSignal(), 4); got != 4 {
		t.Errorf("Got wrong IFF value: %v", got)
	}
	if got := TRUTH.operate(3, 7); got != MaxSignal() {
		t.Errorf("Got wrong TRUTH value: %v", got)
	}
	if got := FALSIFY.operate(3, 7); got != 0 {
		t.Errorf("Got wrong FALSIFY value: %v", got)
	}
}

func TestCommutative(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for opVal := 0; opVal < NumOps; opVal++ {
		op := interpretOp(opVal)
		for i := 0; i < 100; i++ {
			r1 := SignalType(rnd.Intn(int(MaxSignal())))
			r2 := SignalType(rnd.Intn(int(MaxSignal())))
			if op.operate(r1, r2) != op.operate(r2, r1) {
				t.Errorf("Op %d is not commutative for %d and %d", opVal, r1, r2)
			}
		}
	}
}

func CreateTestSignal(sig SignalType) *Signal {
	return &Signal{
		sources:  make(map[IDType]*Signal),
		neuronID: 0,
		isActive: true,
		Output:   sig,
	}
}

func CreateTestSignalInput(sigs ...SignalType) map[IDType]*Signal {
	m := make(map[IDType]*Signal)
	for i, sig := range sigs {
		m[i] = CreateTestSignal(sig)
	}
	return m
}

func TestSignalingPathway(t *testing.T) {
	n := NewNeuron(0, 2)
	n.AddSynapse(1)
	n.AddSynapse(2)

	sigChan := make(chan *Signal)

	sources := CreateTestSignalInput(1, 2, 3, 4, 5)

	go n.Fire(sources, sigChan)
	got := <-sigChan

	want := &Signal{
		sources:  sources,
		neuronID: 0,
		isActive: true,
		Output:   7,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}
}
