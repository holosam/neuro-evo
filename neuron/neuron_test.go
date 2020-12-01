package neuron

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func testOperator(t *testing.T, op OperatorType, inputs []SignalType, want SignalType) {
	if got := op.operate(inputs); got != want {
		t.Errorf("Got wrong value for op %d: inputs %v, got %d, want %d", op, inputs, got, want)
	}
}

func TestOperators(t *testing.T) {
	testOperator(t, AND, []SignalType{9, 14}, 8)
	testOperator(t, NAND, []SignalType{MaxSignal(), MaxSignal() - 4}, 4)
	testOperator(t, OR, []SignalType{9, 10}, 11)
	testOperator(t, NOR, []SignalType{MaxSignal() - 4, MaxSignal() - 4}, 4)
	testOperator(t, XOR, []SignalType{11, 12}, 7)
	testOperator(t, IFF, []SignalType{MaxSignal(), 4}, 4)
	testOperator(t, ADD, []SignalType{5, 6, 7}, 18)
	testOperator(t, MULTIPLY, []SignalType{5, 6, 2}, 60)
	testOperator(t, GCF, []SignalType{12, 9}, 3)
	testOperator(t, MAX, []SignalType{7, 9}, 9)
	testOperator(t, MIN, []SignalType{7, 9}, 7)
	testOperator(t, TRUTH, []SignalType{3, 7}, MaxSignal())
	testOperator(t, FALSIFY, []SignalType{3, 7}, 0)
}

func TestCommutative(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for opVal := 0; opVal < NumOps; opVal++ {
		op := interpretOp(opVal)
		for i := 0; i < 100; i++ {
			r1 := SignalType(rnd.Intn(int(MaxSignal())))
			r2 := SignalType(rnd.Intn(int(MaxSignal())))
			if op.operate([]SignalType{r1, r2}) != op.operate([]SignalType{r2, r1}) {
				t.Errorf("Op %d is not commutative for %d and %d", opVal, r1, r2)
			}
		}
	}
}

func TestAssociative(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for opVal := 0; opVal < NumOps; opVal++ {
		op := interpretOp(opVal)
		for i := 0; i < 50; i++ {
			r1 := SignalType(rnd.Intn(int(MaxSignal())))
			r2 := SignalType(rnd.Intn(int(MaxSignal())))
			r3 := SignalType(rnd.Intn(int(MaxSignal())))
			v1 := op.operate([]SignalType{r1, r2, r3})
			v2 := op.operate([]SignalType{r2, r3, r1})
			if v1 != v2 {
				t.Errorf("Op %v is not associative for [%d, %d, %d], got %d vs %d", op, r1, r2, r3, v1, v2)
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
