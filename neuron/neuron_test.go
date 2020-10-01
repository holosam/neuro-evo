package neuron

import (
	"math"
	"reflect"
	"testing"
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

func TestOverflow(t *testing.T) {
	if got := MULTIPLY.operate(20, 20); got != (20*20)-256 {
		t.Errorf("Got wrong MULTIPLY value: %v", got)
	}
}

func TestGeneticOperate(t *testing.T) {
	if got := Grow(0, 1).Operate([]uint8{1, 2}); got != 3 {
		t.Errorf("Want 3, got %d", got)
	}
	if got := Grow(0, 1).Operate([]uint8{1, 2, 3, 4, 5}); got != 15 {
		t.Errorf("Want 15, got %d", got)
	}
}

func TestGrowingGenetic(t *testing.T) {
	g := Grow(27, 7)
	if g.op != IFF {
		t.Errorf("Want IFF, got %v", g.op)
	}

	// 17 % 7 = *3*, 17 - (3+1) = 13, 13 % 7 = *6*, 13 - (6+1) = 6
	want := make(IntSet)
	want[3] = member
	want[6] = member

	if !reflect.DeepEqual(want, g.downstream) {
		t.Errorf("Want %v, got %v", want, g.downstream)
	}
}

func TestSignalingPathway(t *testing.T) {
	g := Grow(20, 3)
	sigChan := make(chan Signal)

	go g.Fire(sigChan, []uint8{1, 2, 3, 4, 5})
	sig := <-sigChan

	if !reflect.DeepEqual(sig.nIndicies, g.downstream) {
		t.Errorf("Want %v, got %v", sig.nIndicies, g.downstream)
	}
	if sig.val != 15 {
		t.Errorf("Want 15, got %d", sig.val)
	}
}
