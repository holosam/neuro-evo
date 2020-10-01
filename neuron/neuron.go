package neuron

import (
	"log"
	"math"
	"sync"
)

// SignalType is the value held in a Neuron
type SignalType = uint8

type Signal struct {
	val    SignalType
	nIndex int
}

// Neuron docs
type Neuron interface {
	// Trigger downstream
	MaybeFire()

	// Encode relevant data into a string to be put into DNA
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

	downstream2 []int

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
	a.mu.RUnlock()

	for _, n := range a.downstream {
		// Remember the pattern should be the outside calling it with a context,
		// not launching a goroutine in the function
		go func(n Neuron, s SignalType) {
			n.Signal(s)
		}(n, finalSignal)
	}

	// Wait for all the signaling to be done here?
	// Would need to use channels probably
}

// OperatorType for genetic neurons
type OperatorType int

const (
	ADD OperatorType = iota
	MULTIPLY
	AND
	NAND
	OR
	NOR
	XOR
	IFF
	TRUTH
	FALSIFY
)

const numOps = 10

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

func interpretOp(x int) OperatorType {
	ops := [...]OperatorType{ADD, MULTIPLY, AND, NAND, OR, NOR, XOR, IFF, TRUTH, FALSIFY}
	return ops[x]
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

// Grow creates a GeneticNeuron from a snippet of DNA
func Grow(snippet, numNeurons int) *Genetic {
	op := interpretOp(snippet % numOps)

	genetic := Genetic{
		op:             op,
		AbstractNeuron: AbstractNeuron{},
	}
	genetic.AbstractNeuron.Neuron = &genetic

	for nIndex := snippet - numOps; nIndex > 0; nIndex %= numNeurons {
		genetic.downstream2 = append(genetic.downstream2, nIndex)
	}

	return &genetic
}
