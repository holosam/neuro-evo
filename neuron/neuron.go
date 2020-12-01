package neuron

import (
	"log"
	"math"
)

// SignalType is the value held in a neuron. "byte" is an alias for uint8.
type SignalType = uint8

// MaxSignal returns the highest number for the signal type, to
// avoid having to change math.Max___ everywhere in the code.
func MaxSignal() SignalType {
	return math.MaxUint8
}

// OperatorType enum for different commutative operations that neurons
// can perform on their inputs.
type OperatorType int

const (
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

// IDType standardizes the type of IDs used everywhere, so they can
// be distingushed from normal int values.
type IDType = int

type void struct{}

// IDSet is a set of ids. An insert looks like `set[id] = member`
type IDSet = map[IDType]void

var member void

type Neuron struct {
	id       IDType
	op       OperatorType
	synapses IDSet
}

func NewNeuron(id IDType, op OperatorType) *Neuron {
	return &Neuron{
		id:       id,
		op:       op,
		synapses: make(IDSet),
	}
}

func (n *Neuron) AddSynapse(id IDType) {
	if id != n.id {
		n.synapses[id] = member
	}
}

func (n *Neuron) RemoveSynapse(id IDType) {
	delete(n.synapses, id)
}

func (n *Neuron) Fire(inputs []SignalType) SignalType {
	output := inputs[0]
	for i := 1; i < len(inputs); i++ {
		output = n.op.operate(output, inputs[i])
	}
	return output
}
