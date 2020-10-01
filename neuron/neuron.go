package neuron

import (
	"log"
	"math"
	"sync"
)

// SignalType is the value held in a Neuron
type SignalType = uint8

// Neuron docs
type Neuron interface {
	// Trigger downstream
	MaybeFire()

	// Copy relevant data into a string to be put into DNA
	// Encode() string

	// Receive an input
	Signal(signal SignalType)

	// Operate on inputs
	Operate() SignalType
}

// AbstractNeuron docs
type AbstractNeuron struct {
	Neuron

	// Initialized on creaton based on DNA
	downstream []Neuron

	mu sync.RWMutex

	pendingSignals []SignalType
}

// Signal adds a pending input to the queue.
func (a *AbstractNeuron) Signal(signal SignalType) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pendingSignals = append(a.pendingSignals, signal)
}

// MaybeFire fires the neuron if there are at least 2 inputs.
func (a *AbstractNeuron) MaybeFire() {
	a.mu.RLock()
	if len(a.pendingSignals) < 2 {
		a.mu.RUnlock()
		return
	}
	finalSignal := a.Operate()

	for _, n := range a.downstream {
		go func(n Neuron, s SignalType) {
			n.Signal(s)
		}(n, finalSignal)
	}

	a.mu.RUnlock()
	// Wait for all the signalling to be done here?
	// Would need to use channels probably
}

// OperatorType for genetic neurons
type OperatorType int

const (
	ADD      OperatorType = 0
	MULTIPLY OperatorType = 1
	AND      OperatorType = 2
	NAND     OperatorType = 3
	OR       OperatorType = 4
	NOR      OperatorType = 5
	XOR      OperatorType = 6
	IFF      OperatorType = 7
	TRUTH    OperatorType = 8
	FALSIFY  OperatorType = 9
)

func (op OperatorType) operate(a, b SignalType) SignalType {
	switch op {
	case ADD:
		return a + b
	case MULTIPLY:
		return a * b
	case AND:
		return a & b
	case NAND:
		return ^(a & b)
	case OR:
		return a | b
	case NOR:
		return ^(a | b)
	case XOR:
		return a ^ b
	case IFF:
		return ^(a ^ b)
	case TRUTH:
		// Note this will need to be changed if SignalType does.
		return math.MaxUint8
	case FALSIFY:
		return 0
	default:
		log.Fatalf("Unhandled operator: %d", op)
		return 0
	}
}

// Genetic implements Neuron
type Genetic struct {
	AbstractNeuron

	op OperatorType
}

func (g *Genetic) Operate() SignalType {
	g.mu.RLock()
	defer g.mu.RUnlock()
	finalSignal := g.op.operate(g.pendingSignals[0], g.pendingSignals[1])
	for i := 2; i < len(g.pendingSignals); i++ {
		finalSignal = g.op.operate(finalSignal, g.pendingSignals[i])
	}
	return finalSignal
}

// func (g *Genetic) MaybeFire() {
// }
