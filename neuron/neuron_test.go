package neuron

import (
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

func TestSignalingPathway(t *testing.T) {
	sigChan := make(chan Signal)
	n := Neuron{
		snip:    MakeSnippet(0, 2, 1, 2),
		sigChan: sigChan,
	}

	n.ReceiveSignal(1)
	n.ReceiveSignal(2)
	n.ReceiveSignal(3)
	n.ReceiveSignal(4)
	n.ReceiveSignal(5)
	go n.Fire()
	sig := <-sigChan

	if !reflect.DeepEqual(sig.source.Synapses, n.snip.Synapses) {
		t.Errorf("Want %v, got %v", sig.source.Synapses, n.snip.Synapses)
	}
	want := SignalType(7)
	if sig.signal != want {
		t.Errorf("Want %d, got %d", want, sig.signal)
	}
}

func TestVision(t *testing.T) {
	sigChan := make(chan Signal)
	n := Neuron{
		snip:     MakeSnippet(0, 0, 1, 2),
		sigChan:  sigChan,
		isVision: true,
	}

	go n.Fire()
	sig := <-sigChan
	if got := sig.isActive; got {
		t.Errorf("Want inactive signal, got active=%v", got)
	}

	n.ReceiveSignal(1)
	go n.Fire()
	sig = <-sigChan
	if got := sig.signal; got != 1 {
		t.Errorf("Want 1, got %d", got)
	}

	n.isVision = false
	n.ReceiveSignal(1)
	go n.Fire()
	sig = <-sigChan
	if got := sig.isActive; got {
		t.Errorf("Want inactive signal, got active=%v", got)
	}
}
