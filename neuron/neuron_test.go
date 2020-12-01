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

func TestAssociative(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for opVal := 0; opVal < NumOps; opVal++ {
		op := interpretOp(opVal)
		for i := 0; i < 100; i++ {
			r1 := SignalType(rnd.Intn(int(MaxSignal())))
			r2 := SignalType(rnd.Intn(int(MaxSignal())))
			r3 := SignalType(rnd.Intn(int(MaxSignal())))
			v1 := op.operate(op.operate(r1, r2), r3)
			v2 := op.operate(op.operate(r2, r3), r1)
			if v1 != v2 {
				t.Errorf("Op %d is not associative for %d, %d, and %d", opVal, r1, r2, r3)
			}
		}
	}
}

func TestSynapses(t *testing.T) {
	n := NewNeuron(0, IFF)
	n.AddSynapse(0) // Can't add own ID.
	n.AddSynapse(1)
	n.AddSynapse(2)
	n.RemoveSynapse(2)
	want := make(IDSet)
	want[1] = member
	if got := n.synapses; !reflect.DeepEqual(n.synapses, want) {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestFire(t *testing.T) {
	n := NewNeuron(0, OR)
	got := n.Fire([]SignalType{1, 2, 3, 4, 5})
	if want := SignalType(7); got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
}
