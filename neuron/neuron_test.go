package neuron

import (
	"math"
	"reflect"
	"testing"
)

func TestOperators(t *testing.T) {
	// if got := ADD.operate(3, 7); got != 10 {
	// 	t.Errorf("Got wrong ADD value: %v", got)
	// }
	// if got := MULTIPLY.operate(3, 7); got != 21 {
	// 	t.Errorf("Got wrong MULTIPLY value: %v", got)
	// }
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

// func TestOverflow(t *testing.T) {
// 	if got := MULTIPLY.operate(20, 20); got != (20*20)-(math.MaxUint8+1) {
// 		t.Errorf("Got wrong MULTIPLY value: %v", got)
// 	}
// }

func TestSignalingPathway(t *testing.T) {
	sigChan := make(chan Signal)
	n := Neuron{
		snip:    MakeSnippet(0, 1, 2),
		sigChan: sigChan,
	}

	go n.Fire([]SignalType{1, 2, 3, 4, 5})
	sig := <-sigChan

	if !reflect.DeepEqual(sig.synapses, n.snip.Synapses) {
		t.Errorf("Want %v, got %v", sig.synapses, n.snip.Synapses)
	}
	if sig.val != 15 {
		t.Errorf("Want 15, got %d", sig.val)
	}
}

func TestVision(t *testing.T) {
	sigChan := make(chan Signal)
	n := Neuron{
		snip:     MakeSnippet(0, 1, 2),
		sigChan:  sigChan,
		isVision: true,
	}

	go n.Fire([]SignalType{})
	sig := <-sigChan
	if got := sig.active; got {
		t.Errorf("Want inactive signal, got active=%v", got)
	}

	go n.Fire([]SignalType{1})
	sig = <-sigChan
	if got := sig.val; got != 1 {
		t.Errorf("Want 1, got %d", got)
	}

	n.isVision = false
	go n.Fire([]SignalType{1})
	sig = <-sigChan
	if got := sig.active; got {
		t.Errorf("Want inactive signal, got active=%v", got)
	}
}
