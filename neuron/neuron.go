package neuron

import (
	"log"
	"math"
)

// SignalType is the value held in a Neuron
type SignalType = int64

// Neuron docs
type Neuron interface {
	// Trigger downstream
	MaybeFire()

	// Copy relevant data into a string to be put into DNA
	Encode() string

	// Receive an input
	Signal(SignalType)
}

// AbstractNeuron docs
type AbstractNeuron struct {
	pendingInputs []SignalType

	downstream []Neuron
}

// Implement signal and maybe fire here

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
		return math.MaxInt64
	case FALSIFY:
		return 0
	default:
		log.Fatalf("Unhandled operator: %d", op)
		return 0
	}
}

// Genetic implements Neuron
type Genetic struct {
	op OperatorType
}
