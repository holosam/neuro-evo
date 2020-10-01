package neuron

import (
	"log"
	"math"
)

// SignalType is the value held in a Neuron
type SignalType = uint8

type void struct{}

type IntSet = map[int]void

type Signal struct {
	val       SignalType
	nIndicies IntSet
}

// Neuron docs
type Neuron interface {
	// Trigger downstream
	Fire(sigChan chan Signal, sigs []SignalType)

	// Operate on inputs
	Operate(sigs []SignalType) SignalType
}

// AbstractNeuron docs
type AbstractNeuron struct {
	Neuron

	downstream IntSet
}

// Fire fires the neuron if there are at least 2 inputs.
func (a *AbstractNeuron) Fire(sigChan chan Signal, sigs []SignalType) {
	sigChan <- Signal{
		val:       a.Operate(sigs),
		nIndicies: a.downstream,
	}
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

// Operate on a list of signals
func (g *Genetic) Operate(sigs []SignalType) SignalType {
	finalSignal := g.op.operate(sigs[0], sigs[1])
	for i := 2; i < len(sigs); i++ {
		finalSignal = g.op.operate(finalSignal, sigs[i])
	}
	return finalSignal
}

var member void

// Grow creates a GeneticNeuron from a snippet of DNA
func Grow(snippet, numNeurons int) *Genetic {
	op := interpretOp(snippet % numOps)

	g := Genetic{
		op:             op,
		AbstractNeuron: AbstractNeuron{downstream: make(IntSet)},
	}
	g.AbstractNeuron.Neuron = &g

	// Not sure if this should be -numOps or -op
	snipLeft := snippet - numOps
	for snipLeft > 0 {
		nIndex := snipLeft % numNeurons
		g.downstream[nIndex] = member
		// Gets caught in a vortex when nIndex == numNeurons - 1
		snipLeft -= nIndex + 1
	}

	return &g
}
