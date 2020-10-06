package neuron

import (
	"log"
	"math"
	"sync"
)

// IDType standardizes the type of IDs used everywhere, so they can
// be distingushed from normal int values.
type IDType = int

type void struct{}

// IDSet is a set of ids. An insert looks like `set[id] = member`
type IDSet = map[IDType]void

var member void

// OperatorType enum for different commutative operations that neurons
// can perform on their inputs.
type OperatorType int

const (
	// ADD
	// MULTIPLY
	AND OperatorType = iota
	NAND
	OR
	NOR
	XOR
	IFF
	TRUTH
	FALSIFY
)

const NumOps = 8

func (op OperatorType) operate(a, b SignalType) SignalType {
	switch op {
	// case ADD:
	// 	return a + b
	// case MULTIPLY:
	// 	return a * b
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
		return MaxSignal()
	case FALSIFY:
		return 0
	default:
		log.Fatalf("Unhandled operator: %d", op)
		return 0
	}
}

func interpretOp(x int) OperatorType {
	ops := [...]OperatorType{AND, NAND, OR, NOR, XOR, IFF, TRUTH, FALSIFY}
	return ops[x]
}

// Snippet is a piece of DNA holding everything needed to create a Neuron.
type Snippet struct {
	ID       IDType
	Op       OperatorType
	Synapses IDSet
}

func (s *Snippet) SetOp(opVal int) {
	s.Op = interpretOp(opVal)
}

func (s *Snippet) AddSynapse(id IDType) {
	if id != s.ID {
		s.Synapses[id] = member
	}
}

func (s *Snippet) RemoveSynapse(id IDType) {
	delete(s.Synapses, id)
}

func MakeSnippet(id IDType, opVal int, synapses ...IDType) *Snippet {
	return MakeSnippetOp(id, interpretOp(opVal), synapses...)
}

func MakeSnippetOp(id IDType, op OperatorType, synapses ...IDType) *Snippet {
	s := Snippet{
		ID:       id,
		Op:       op,
		Synapses: make(IDSet),
	}
	for _, synapse := range synapses {
		s.AddSynapse(synapse)
	}
	return &s
}

// SignalType is the value held in a neuron
type SignalType = uint8

// MaxSignal returns the highest number for the signal type, to
// avoid having to change math.Max___ everywhere in the code.
func MaxSignal() SignalType {
	return math.MaxUint8
}

// Signal holds a value that this neuron is firing off.
type Signal struct {
	isActive bool
	signal   SignalType
	source   *Snippet
}

type Neuron struct {
	snip     *Snippet
	sigChan  chan Signal
	isVision bool

	mu             sync.Mutex
	pendingSignals []SignalType
}

func (n *Neuron) ReceiveSignal(input SignalType) {
	n.mu.Lock()
	n.pendingSignals = append(n.pendingSignals, input)
	n.mu.Unlock()
}

func (n *Neuron) Fire() {
	n.mu.Lock()
	// Vision neurons only need 1 input signal, others need 2.
	if len(n.pendingSignals) == 0 || (len(n.pendingSignals) == 1 && !n.isVision) {
		// Send an empty struct on the channel to alert the caller
		// that there is nothing to do.
		n.mu.Unlock()
		n.sigChan <- Signal{}
		return
	}

	output := n.pendingSignals[0]
	for i := 1; i < len(n.pendingSignals); i++ {
		output = n.snip.Op.operate(output, n.pendingSignals[i])
	}

	n.pendingSignals = make([]SignalType, 0)
	n.mu.Unlock()

	n.sigChan <- Signal{
		isActive: true,
		signal:   output,
		source:   n.snip,
	}
}
