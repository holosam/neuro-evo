package neuron

import (
	"log"
	"math"
)

// SignalType is the value held in a Neuron
type SignalType = uint8

type void struct{}

type IntSet = map[int]void

var member void

type Signal struct {
	val       SignalType
	nIndicies IntSet
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

type Snippet struct {
	Op       OperatorType
	Synapses IntSet
}

func (s *Snippet) SetOp(op int) {
	s.Op = interpretOp(op)
}

func (s *Snippet) AddSynapse(id int) {
	s.Synapses[id] = member
}

func (s *Snippet) RemoveSynapse(id int) {
	delete(s.Synapses, id)
}

func MakeSnippet(op OperatorType, synapes ...int) Snippet {
	s := Snippet{
		Op:       op,
		Synapses: make(IntSet),
	}
	for _, synapse := range synapes {
		s.AddSynapse(synapse)
	}
	return s
}

type Neuron struct {
	snip Snippet
}

// Fire fires the neuron if there are at least 2 inputs.
func (n Neuron) Fire(sigChan chan Signal, sigs []SignalType) {
	// Will need to question this assumption for vision neurons
	if len(sigs) < 2 {
		return
	}

	signal := sigs[0]
	for i := 1; i < len(sigs); i++ {
		signal = n.snip.Op.operate(signal, sigs[i])
	}

	sigChan <- Signal{
		val:       signal,
		nIndicies: n.snip.Synapses,
	}
}
