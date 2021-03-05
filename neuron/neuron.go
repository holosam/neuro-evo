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

// OperatorType enum for different operations that neurons can perform on
// signal values. The operations must be associative and commutative.
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

// NumOps is the total number of OperatorTypes, used to pick one randomly.
const NumOps = 13

// Operate performs the operation on a series of inputs.
func (op OperatorType) Operate(sigs []SignalType) SignalType {
	switch op {
	case AND:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a & b })
	case NAND:
		return ^AND.Operate(sigs)
	case OR:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a | b })
	case NOR:
		return ^OR.Operate(sigs)
	case XOR:
		return applyToAll(sigs, func(a, b SignalType) SignalType { return a ^ b })
	case IFF:
		return ^XOR.Operate(sigs)
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

func applyToAll(sigs []SignalType, operateFn func(a, b SignalType) SignalType) SignalType {
	x := sigs[0]
	for i := 1; i < len(sigs); i++ {
		x = operateFn(x, sigs[i])
	}
	return x
}

// InterpretOp converts an int to its corresponding OperatorType.
func InterpretOp(x int) OperatorType {
	return [...]OperatorType{AND, NAND, OR, NOR, XOR, IFF, ADD, MULTIPLY, GCF, MAX, MIN, TRUTH, FALSIFY}[x]
}

// IDType standardizes the type of IDs used everywhere, so they can
// be distingushed from normal int values.
type IDType = int

// Void is an empty struct to help turn a map into a set.
type Void struct{}

// IDSet is a set of ids. An insert looks like `set[id] = member`
type IDSet = map[IDType]Void

var member Void

// NeuronType is an enum for neuron specializations.
type NeuronType int

const (
	// SENSE neurons represent the visual cortex, which accept external inputs.
	SENSE NeuronType = iota
	// INTER neurons make up the bulk of the processing power within the brain.
	INTER
	// MOTOR neurons give results as output.
	MOTOR
)

// NeuronTypes holds all possible enum values for looping.
var NeuronTypes = []NeuronType{SENSE, INTER, MOTOR}

// Neuron is the base struct of this entire project. It performs a simple
// operation on its inputs and gives one output.
type Neuron struct {
	Op OperatorType

	HasSeed bool
	Seed    SignalType
}

// NewNeuron inits a neuron from an operation.
func NewNeuron(op OperatorType) *Neuron {
	return &Neuron{
		Op:      op,
		HasSeed: false,
		Seed:    0,
	}
}

// SetSeed accepts a seed value that will be used as an input to every Fire().
func (n *Neuron) SetSeed(seed SignalType) {
	n.Seed = seed
	n.HasSeed = true
}

// RemoveSeed doesn't actually change the Seed variable since the value will
// only be respected if HasSeed is true.
func (n *Neuron) RemoveSeed() {
	n.HasSeed = false
}

// Copy returns a copy of this Neuron's fields in a different pointer.
func (n *Neuron) Copy() *Neuron {
	return &Neuron{
		Op:      n.Op,
		HasSeed: n.HasSeed,
		Seed:    n.Seed,
	}
}

// IsEquiv returns if all the Neuron fields are equivalent, even if the two
// pointers differ. If they both don't have seeds set, then it doesn't matter
// what value in the Seed field.
func (n *Neuron) IsEquiv(other *Neuron) bool {
	return n.Op == other.Op && n.HasSeed == other.HasSeed &&
		(!n.HasSeed || (n.HasSeed && n.Seed == other.Seed))
}

// Fire runs the Neuron's operation on all the inputs, including the Seed.
func (n *Neuron) Fire(inputs []SignalType) SignalType {
	// Seed inputs are "sticky" so they come back for every trigger even when the
	// rest of the inputs gets cleared.
	if n.HasSeed {
		inputs = append(inputs, n.Seed)
	}
	return n.Op.Operate(inputs)
}

// Synapse is a simple representation of a neuron -> neuron connection.
type Synapse struct {
	src IDType
	dst IDType
}
