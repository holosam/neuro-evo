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

type DNA struct {
	Snippets map[IDType]*Neuron
	NextID   IDType

	Seeds map[IDType]SignalType

	VisionIDs *IndexedIDs
	MotorIDs  *IndexedIDs
}

func NewDNA() *DNA {
	return &DNA{
		Snippets:  make(map[IDType]*Neuron, 0),
		NextID:    0,
		Seeds:     make(map[IDType]SignalType, 0),
		VisionIDs: NewIndexedIDs(),
		MotorIDs:  NewIndexedIDs(),
	}
}

func (src *DNA) DeepCopy() *DNA {
	dst := &DNA{
		Snippets:  make(map[IDType]*Neuron, len(src.Snippets)),
		NextID:    src.NextID,
		Seeds:     make(map[IDType]SignalType, len(src.Seeds)),
		VisionIDs: src.VisionIDs.Copy(),
		MotorIDs:  src.MotorIDs.Copy(),
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

func (d *DNA) AddSnippet(opVal int) *Neuron {
	s := NewNeuron(d.NextID, interpretOp(opVal))
	d.Snippets[d.NextID] = s
	d.NextID++
	return s
}

func (d *DNA) DeleteSnippet(id IDType) {
	delete(d.Snippets, id)

	d.VisionIDs.RemoveID(id)
	d.MotorIDs.RemoveID(id)

	for _, snip := range d.Snippets {
		snip.RemoveSynapse(id)
	}
}

func (d *DNA) AddSynapse(snipID, synID IDType) {
	if d.MotorIDs.HasID(snipID) {
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

func (d *DNA) AddVisionID(id IDType) {
	if !d.VisionIDs.HasID(id) {
		d.VisionIDs.InsertID(id)
	}
}

func (d *DNA) AddMotorID(id IDType) {
	if !d.MotorIDs.HasID(id) {
		d.MotorIDs.InsertID(id)
	}
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

		if d.VisionIDs.HasID(id) {
			s += fmt.Sprintf("(V%d)=", d.VisionIDs.GetIndex(id))
		}

		if d.MotorIDs.HasID(id) {
			s += fmt.Sprintf("(M%d)=", d.MotorIDs.GetIndex(id))
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
		// Send the signal to the vision ID at the signal's index.
		b.addPendingInput(b.dna.VisionIDs.GetId(i), sig)
	}
}

func (b *Brain) StepFunction() []Signal {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	sigChan := make(chan *Signal)
	for neuronID, signals := range b.pendingSignals {
		fmt.Printf("Firing neuron %d\n", neuronID)
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
		fmt.Printf("Got signal %v\n", signal)

		// May receive an inactive signal if the firing threshold isn't met.
		if !signal.isActive {
			continue
		}

		// The neuron fired, so clear any pending signals.
		delete(b.pendingSignals, signal.neuronID)

		if b.dna.MotorIDs.HasID(signal.neuronID) {
			// Resize the outputs to handle the number of motor neurons.
			if len(outputs) == 0 {
				outputs = make([]Signal, b.dna.MotorIDs.Length())
			}
			// Insert the signal at the index of the motor neuron ID.
			outputs[b.dna.MotorIDs.GetIndex(signal.neuronID)] = *signal
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
			fmt.Printf("Adding pending signal for %d: %v\n", neuronID, signal)
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

	fmt.Printf("Adding pending input for %d: %v\n", neuronID, signal)
	b.addPendingSignal(b.pendingSignals, neuronID, signal)
}

func (b *Brain) addPendingSignal(pendingSignals PendingSignalsMap,
	neuronID IDType, signal *Signal) {
	if _, exists := pendingSignals[neuronID]; !exists {
		pendingSignals[neuronID] = make(map[IDType]*Signal)
	}
	pendingSignals[neuronID][signal.neuronID] = signal
}
