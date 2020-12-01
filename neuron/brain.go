package neuron

import (
	"fmt"
	"strings"
)

type IndexedIDs struct {
	IDToIndex map[IDType]int
	IndexToID map[int]IDType
}

func NewIndexedIDs() *IndexedIDs {
	return &IndexedIDs{
		IDToIndex: make(map[IDType]int),
		IndexToID: make(map[int]IDType),
	}
}

func (x *IndexedIDs) InsertID(id IDType) {
	if !x.HasID(id) {
		nextIndex := x.Length()
		x.IDToIndex[id] = nextIndex
		x.IndexToID[nextIndex] = id
	}
}

func (x *IndexedIDs) RemoveID(id IDType) {
	if index, exists := x.IDToIndex[id]; exists {
		length := x.Length()
		for i := index; i < length-1; i++ {
			idToMove := x.IndexToID[i+1]
			x.IndexToID[i] = idToMove
			x.IDToIndex[idToMove] = i
		}

		delete(x.IDToIndex, id)
		delete(x.IndexToID, length-1)
	}
}

func (x *IndexedIDs) HasID(id IDType) bool {
	_, ok := x.IDToIndex[id]
	return ok
}

func (x *IndexedIDs) HasIndex(index int) bool {
	_, ok := x.IndexToID[index]
	return ok
}

func (x *IndexedIDs) GetIndex(id IDType) int {
	return x.IDToIndex[id]
}

func (x *IndexedIDs) GetId(index int) IDType {
	return x.IndexToID[index]
}

func (x *IndexedIDs) Length() int {
	return len(x.IDToIndex)
}

func (x *IndexedIDs) Copy() *IndexedIDs {
	c := &IndexedIDs{
		IDToIndex: make(map[IDType]int, x.Length()),
		IndexToID: make(map[int]IDType, x.Length()),
	}
	for id, index := range x.IDToIndex {
		c.IDToIndex[id] = index
		c.IndexToID[index] = id
	}
	return c
}

type NeuronType int

const (
	SENSE NeuronType = iota
	INTER
	MOTOR
)

var neuronTypes = []NeuronType{SENSE, INTER, MOTOR}

type DNA struct {
	NeuronIDs map[NeuronType]*IndexedIDs
	Snippets  map[IDType]*Neuron
	NextID    IDType

	Seeds map[IDType]SignalType
}

func NewDNA() *DNA {
	d := &DNA{
		NeuronIDs: make(map[NeuronType]*IndexedIDs, len(neuronTypes)),
		Snippets:  make(map[IDType]*Neuron, 0),
		NextID:    0,
		Seeds:     make(map[IDType]SignalType, 0),
	}

	for _, nType := range neuronTypes {
		d.NeuronIDs[nType] = NewIndexedIDs()
	}
	return d
}

func (src *DNA) DeepCopy() *DNA {
	dst := &DNA{
		NeuronIDs: make(map[NeuronType]*IndexedIDs, len(neuronTypes)),
		Snippets:  make(map[IDType]*Neuron, len(src.Snippets)),
		NextID:    src.NextID,
		Seeds:     make(map[IDType]SignalType, len(src.Seeds)),
	}

	for nType, indexedIDs := range src.NeuronIDs {
		dst.NeuronIDs[nType] = indexedIDs.Copy()
	}

	for id, neuron := range src.Snippets {
		n := NewNeuron(id, neuron.op)
		for syn := range neuron.synapses {
			n.AddSynapse(syn)
		}
		dst.Snippets[id] = n
	}

	for id, seed := range src.Seeds {
		dst.Seeds[id] = seed
	}

	return dst
}

func (d *DNA) AddSnippet(nType NeuronType, op OperatorType) IDType {
	addedID := d.NextID
	d.Snippets[addedID] = NewNeuron(addedID, op)
	d.NeuronIDs[nType].InsertID(addedID)
	d.NextID++
	return addedID
}

func (d *DNA) DeleteSnippet(id IDType) {
	delete(d.Snippets, id)

	for _, nType := range neuronTypes {
		d.NeuronIDs[nType].RemoveID(id)
	}

	delete(d.Seeds, id)

	for _, snip := range d.Snippets {
		snip.RemoveSynapse(id)
	}
}

func (d *DNA) AddSynapse(snipID, synID IDType) {
	if d.NeuronIDs[MOTOR].HasID(snipID) {
		return
	}

	if snip, snipExists := d.Snippets[snipID]; snipExists {
		if _, synExists := d.Snippets[synID]; synExists {
			snip.AddSynapse(synID)
		}
	}
}

func (d *DNA) RemoveSynapse(snipID, synID IDType) {
	if snip, snipExists := d.Snippets[snipID]; snipExists {
		snip.RemoveSynapse(synID)
	}
}

func (d *DNA) SetSeed(id IDType, seed SignalType) {
	d.Seeds[id] = seed
}

func (d *DNA) RemoveSeed(id IDType) {
	delete(d.Seeds, id)
}

func (d *DNA) GetNeuronType(id IDType) NeuronType {
	for _, nType := range neuronTypes {
		if d.NeuronIDs[nType].HasID(id) {
			return nType
		}
	}
	return -1
}

func (d *DNA) PrettyPrint() string {
	var sb strings.Builder
	sortedSnips := make([]*Neuron, d.NextID)
	for id, snip := range d.Snippets {
		sortedSnips[id] = snip
	}

	for id, snip := range sortedSnips {
		if snip == nil {
			continue
		}

		if d.NeuronIDs[SENSE].HasID(id) {
			sb.WriteString(fmt.Sprintf("(V%d)=", d.NeuronIDs[SENSE].GetIndex(id)))
		}

		if d.NeuronIDs[MOTOR].HasID(id) {
			sb.WriteString(fmt.Sprintf("(M%d)=", d.NeuronIDs[MOTOR].GetIndex(id)))
		}

		sb.WriteString(fmt.Sprintf("%d:%d", id, snip.op))

		if seed, exists := d.Seeds[id]; exists {
			sb.WriteString(fmt.Sprintf("<%d", seed))
		}

		if len(snip.synapses) > 0 {
			sb.WriteString("[")
			sortedSyns := make([]bool, d.NextID)
			for synapse := range snip.synapses {
				sortedSyns[synapse] = true
			}
			for synID, exists := range sortedSyns {
				if exists {
					sb.WriteString(fmt.Sprintf("%d,", synID))
				}
			}
			sb.WriteString("]")
		}
		sb.WriteString("  ")
	}
	return sb.String()
}

// Brain docs
type Brain struct {
	dna            *DNA
	pendingSignals map[IDType][]SignalType
	outputSignals  map[IDType]SignalType
}

func Flourish(dna *DNA) *Brain {
	b := &Brain{
		dna:            dna,
		pendingSignals: make(map[IDType][]SignalType, len(dna.Snippets)),
		outputSignals:  make(map[IDType]SignalType, dna.NeuronIDs[MOTOR].Length()),
	}

	for id, seed := range dna.Seeds {
		b.addPendingSignal(id, seed)
	}

	return b
}

func (b *Brain) SeeInput(sigs []SignalType) {
	for i, sig := range sigs {
		fmt.Printf("input for signal %d is %d\n", i, sig)
		// Send the signal to the vision ID at the signal's index.
		b.addPendingSignal(b.dna.NeuronIDs[SENSE].GetId(i), sig)
	}
}

func (b *Brain) Output() []SignalType {
	output := make([]SignalType, len(b.outputSignals))
	for id, sig := range b.outputSignals {
		output[b.dna.NeuronIDs[MOTOR].GetIndex(id)] = sig
	}
	return output
}

func (b *Brain) StepFunction() bool {
	// Create a separate map that will be merged with pendingSignals after all
	// firing is done. This avoids a race condition where a synapse would add
	// a pending signal to the map and then be cleared later if that neuron fires
	// too.
	nextPending := make(map[IDType][]SignalType, len(b.dna.Snippets))

	done := false
	for neuronID, inputs := range b.pendingSignals {
		fmt.Printf("checking neuron %d with pending signals: %v\n", neuronID, inputs)
		// Neurons fire when they have at least 2 signals.
		if len(inputs) < 2 {
			continue
		}

		neuron := b.dna.Snippets[neuronID]
		output := neuron.Fire(inputs)
		fmt.Printf("firing neuron %v and got output: %d\n", neuron, output)

		// Clear this neuron's pending signals now that it has fired.
		// It's okay to edit the underlying map while iterating.
		delete(b.pendingSignals, neuronID)

		if b.dna.NeuronIDs[MOTOR].HasID(neuronID) {
			// Add the signal to the outputSignals only if it's the first time this
			// motor neuron has fired.
			if _, ok := b.outputSignals[neuronID]; !ok {
				b.outputSignals[neuronID] = output
			}
			fmt.Printf("output signals: %v\n", b.outputSignals)
			// Done is true if all motor neurons have an output.
			done = len(b.outputSignals) == b.dna.NeuronIDs[MOTOR].Length()
		}

		// Queue up signal for all downstream neurons.
		for synID := range neuron.synapses {
			nextPending[synID] = append(nextPending[synID], output)
		}
	}

	// Merge in nextPending now that the step is over.
	for neuronID, signals := range nextPending {
		delete(b.pendingSignals, neuronID)

		for _, sig := range signals {
			b.addPendingSignal(neuronID, sig)
		}

		// Seed inputs are "sticky" so they come back after every trigger.
		if seed, ok := b.dna.Seeds[neuronID]; ok {
			b.addPendingSignal(neuronID, seed)
		}
	}

	return done
}

func (b *Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}
