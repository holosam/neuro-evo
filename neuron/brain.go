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
	SENSORY NeuronType = iota
	INTER
	MOTOR
)

var neuronTypes = []NeuronType{SENSORY, INTER, MOTOR}

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

func (d *DNA) NumPathways() int {
	return d.NeuronIDs[SENSORY].Length() + len(d.Seeds)
}

func (d *DNA) PrettyPrint() string {
	s := ""
	sortedSnips := make([]*Neuron, d.NextID)
	for id, snip := range d.Snippets {
		sortedSnips[id] = snip
	}

	for id, snip := range sortedSnips {
		if snip == nil {
			continue
		}

		if d.NeuronIDs[SENSORY].HasID(id) {
			s += fmt.Sprintf("(V%d)=", d.NeuronIDs[SENSORY].GetIndex(id))
		}

		if d.NeuronIDs[MOTOR].HasID(id) {
			s += fmt.Sprintf("(M%d)=", d.NeuronIDs[MOTOR].GetIndex(id))
		}

		s += fmt.Sprintf("%d:%d", id, snip.op)

		if seed, exists := d.Seeds[id]; exists {
			s += fmt.Sprintf("<%d", seed)
		}

		if len(snip.synapses) > 0 {
			s += "["
			sortedSyns := make([]bool, d.NextID)
			for synapse := range snip.synapses {
				sortedSyns[synapse] = true
			}
			for synID, exists := range sortedSyns {
				if exists {
					s += fmt.Sprintf("%d,", synID)
				}
			}
			s = strings.TrimRight(s, ",") + "]"
		}
		s += "  "
	}
	s = strings.TrimRight(s, " ")
	return s
}

// PendingSignalsMap is a map of neurons that will receive a set of signals.
type PendingSignalsMap = map[IDType]map[IDType]*Signal

// Brain docs
type Brain struct {
	dna            *DNA
	pendingSignals PendingSignalsMap
}

func Flourish(dna *DNA) *Brain {
	b := &Brain{
		dna:            dna,
		pendingSignals: make(PendingSignalsMap, len(dna.Snippets)),
	}

	for id, seed := range dna.Seeds {
		b.addPendingInput(id, seed)
	}

	return b
}

func (b *Brain) SeeInput(sigs []SignalType) {
	for i, sig := range sigs {
		if b.dna.NeuronIDs[SENSORY].HasIndex(i) {
			// Send the signal to the vision ID at the signal's index.
			b.addPendingInput(b.dna.NeuronIDs[SENSORY].GetId(i), sig)
		}
	}
}

func (b *Brain) StepFunction() []Signal {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	sigChan := make(chan *Signal)
	for neuronID, signals := range b.pendingSignals {
		// fmt.Printf("Firing neuron %d\n", neuronID)
		go b.dna.Snippets[neuronID].Fire(signals, sigChan)
	}

	// Create a separate map that will be merged with pendingSignals after all
	// firing is done. This avoids a race condition where a synapse would add
	// a pending signal to the map and then be cleared later if that neuron was
	// going to fire anyway.
	nextInputs := make(PendingSignalsMap, len(b.dna.Snippets))
	outputs := make([]Signal, 0)

	for i := 0; i < expectedSignals; i++ {
		signal := <-sigChan
		// fmt.Printf("Got signal %v\n", signal)

		// May receive an inactive signal if the firing threshold isn't met.
		if !signal.isActive {
			continue
		}

		// The neuron fired, so clear any pending signals.
		delete(b.pendingSignals, signal.neuronID)

		if b.dna.NeuronIDs[MOTOR].HasID(signal.neuronID) {
			// Resize the outputs to handle the number of motor neurons.
			if len(outputs) == 0 {
				outputs = make([]Signal, b.dna.NeuronIDs[MOTOR].Length())
			}
			// Insert the signal at the index of the motor neuron ID.
			outputs[b.dna.NeuronIDs[MOTOR].GetIndex(signal.neuronID)] = *signal
		} else {
			for synID := range b.dna.Snippets[signal.neuronID].synapses {
				// Queue up signals to be added to pendingSignals for every synapse.
				b.addPendingSignal(nextInputs, synID, signal)
			}
		}
	}

	// Merge nextSignals into pendingSignals now that the step is over.
	for neuronID, sources := range nextInputs {
		for _, signal := range sources {
			// fmt.Printf("Adding pending signal for %d: %v\n", neuronID, signal)
			b.addPendingSignal(b.pendingSignals, neuronID, signal)
		}
	}

	return outputs
}

// addPendingInput is used to start a signal pathway from an input. The input
// is coming from the "environment" (either vision or DNA seed) so the sources
// map is empty and the neuronID is an arbitrary unique ID.
func (b *Brain) addPendingInput(neuronID IDType, sig SignalType) {
	signal := &Signal{
		sources:  make(map[IDType]*Signal),
		neuronID: -1 - len(b.pendingSignals[neuronID]),
		isActive: true,
		output:   sig,
	}

	// fmt.Printf("Adding pending input for %d: %v with DNA %s\n", neuronID, signal, b.dna.PrettyPrint())
	b.addPendingSignal(b.pendingSignals, neuronID, signal)
}

func (b *Brain) addPendingSignal(pendingSignals PendingSignalsMap,
	neuronID IDType, signal *Signal) {
	if _, exists := pendingSignals[neuronID]; !exists {
		pendingSignals[neuronID] = make(map[IDType]*Signal)
	}
	pendingSignals[neuronID][signal.neuronID] = signal
}
