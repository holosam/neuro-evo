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
	ADD
	MULTIPLY
	GCF
	MAX
	MIN
	TRUTH
	FALSIFY
)

const NumOps = 13

func applyToAll(sigs []SignalType, operateFn func(a, b SignalType) SignalType) SignalType {
	x := sigs[0]
	for i := 1; i < len(sigs); i++ {
		x = operateFn(x, sigs[i])
	}
	return x
}

func (op OperatorType) operate(sigs []SignalType) SignalType {
	switch op {
	case AND:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a & b })
	case NAND:
		return ^AND.operate(sigs)
	case OR:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a | b })
	case NOR:
		return ^OR.operate(sigs)
	case XOR:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a ^ b })
	case IFF:
		return ^XOR.operate(sigs)
	case ADD:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a + b })
	case MULTIPLY:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a * b })
	case GCF:
		return applyToAll(sigs, func(a, b SignalType) SignalType {
			for b != 0 {
				tmp := b
				b = a % b
				a = tmp
			}
			return a
		})
	case MIN:
		return applyToAll(sigs, func(a, b SignalType) SignalType {
			if a < b {
				return a
			}
			return b
		})
	case MAX:
		return applyToAll(sigs, func(a, b SignalType) SignalType {
			if a > b {
				return a
			}
			return b
		})
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
	return [...]OperatorType{AND, NAND, OR, NOR, XOR, IFF, ADD, MULTIPLY, GCF, MAX, MIN, TRUTH, FALSIFY}[x]
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
	return n.op.operate(inputs)
}
