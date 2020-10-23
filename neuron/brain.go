package neuron

import (
	"fmt"
	"strings"

	"github.com/ulule/deepcopier"
)

type IndexedIDs struct {
	idToIndex map[IDType]int
	indexToID map[int]IDType
}

func NewIndexedIDs() *IndexedIDs {
	return &IndexedIDs{
		idToIndex: make(map[IDType]int),
		indexToID: make(map[int]IDType),
	}
}

func (x *IndexedIDs) InsertID(id IDType) {
	nextIndex := x.Length()
	x.idToIndex[id] = nextIndex
	x.indexToID[nextIndex] = id
}

func (x *IndexedIDs) RemoveID(id IDType) {
	if index, exists := x.idToIndex[id]; exists {
		length := x.Length()

		delete(x.idToIndex, id)
		delete(x.indexToID, index)

		for i := index + 1; i < length; i++ {
			idToMove := x.indexToID[i+1]
			x.indexToID[i] = idToMove
			x.idToIndex[idToMove] = i
		}
	}
}

func (x *IndexedIDs) HasID(id IDType) bool {
	_, ok := x.idToIndex[id]
	return ok
}

func (x *IndexedIDs) GetIndex(id IDType) int {
	return x.idToIndex[id]
}

func (x *IndexedIDs) GetId(index int) IDType {
	return x.indexToID[index]
}

func (x *IndexedIDs) Length() int {
	return len(x.idToIndex)
}

type DNA struct {
	snippets map[IDType]*Snippet
	nextID   IDType

	visionIDs *IndexedIDs
	motorIDs  *IndexedIDs
}

func NewDNA() *DNA {
	return &DNA{
		snippets:  make(map[IDType]*Snippet),
		nextID:    0,
		visionIDs: NewIndexedIDs(),
		motorIDs:  NewIndexedIDs(),
	}
}

func (src *DNA) DeepCopy() *DNA {
	dst := &DNA{}
	deepcopier.Copy(src).To(dst)
	return dst
}

func (d *DNA) AddSnippet(opVal int) *Snippet {
	s := MakeSnippet(d.nextID, opVal)
	d.snippets[d.nextID] = s
	d.nextID++
	return s
}

func (d *DNA) DeleteSnippet(id IDType) {
	delete(d.snippets, id)

	d.visionIDs.RemoveID(id)
	d.motorIDs.RemoveID(id)

	for _, snip := range d.snippets {
		snip.RemoveSynapse(id)
	}
}

func (d *DNA) AddSynapse(snipID, synID IDType) {
	if snip, snipExists := d.snippets[snipID]; snipExists {
		if _, synExists := d.snippets[synID]; synExists {
			snip.AddSynapse(synID)
		}
	}
}

func (d *DNA) RemoveSynapse(snipID, synID IDType) {
	if snip, snipExists := d.snippets[snipID]; snipExists {
		snip.RemoveSynapse(synID)
	}
}

func (d *DNA) AddVisionID(id IDType) {
	if !d.visionIDs.HasID(id) {
		d.visionIDs.InsertID(id)
	}
}

func (d *DNA) AddMotorID(id IDType) {
	if !d.motorIDs.HasID(id) {
		d.motorIDs.InsertID(id)
	}
}

func (d *DNA) PrettyPrint() string {
	s := ""
	sortedSnips := make([]*Snippet, d.nextID)
	for id, snip := range d.snippets {
		sortedSnips[id] = snip
	}

	for id, snip := range sortedSnips {
		if snip == nil {
			continue
		}

		if d.visionIDs.HasID(id) {
			s += fmt.Sprintf("(V-%d)=", d.visionIDs.GetIndex(id))
		}

		if d.motorIDs.HasID(id) {
			s += fmt.Sprintf("(M-%d)=", d.motorIDs.GetIndex(id))
		}

		s += fmt.Sprintf("%d:%d", id, snip.Op)

		if len(snip.Synapses) > 0 {
			s += "["
			sortedSyns := make([]bool, d.nextID)
			for synapse := range snip.Synapses {
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
	return s
}

// Brain docs
type Brain struct {
	dna     *DNA
	neurons map[IDType]*Neuron

	pendingSignals IDSet
	sigChan        chan Signal
	motorChan      chan Signal
}

func Flourish(dna *DNA) *Brain {
	b := &Brain{
		dna:     dna,
		neurons: make(map[IDType]*Neuron, len(dna.snippets)),

		pendingSignals: make(IDSet, len(dna.snippets)),
		sigChan:        make(chan Signal),
		motorChan:      make(chan Signal),
	}

	for id, snip := range dna.snippets {
		// Select which signal channel should be injected. Motor neurons fire a
		// different channel.
		selectedChan := b.sigChan
		if dna.motorIDs.HasID(id) {
			selectedChan = b.motorChan
		}

		b.neurons[id] = &Neuron{
			snip:           snip,
			sigChan:        selectedChan,
			isVision:       dna.visionIDs.HasID(id),
			pendingSignals: make([]SignalType, 0),
		}
	}

	return b
}

func (b *Brain) SeeInput(sigs []SignalType) {
	for i, sig := range sigs {
		// Send the signal to the vision ID at the signal's index.
		b.neurons[b.dna.visionIDs.GetId(i)].ReceiveSignal(sig)
	}
}

func (b *Brain) StepFunction() []SignalType {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	for neuronID := range b.pendingSignals {
		go b.neurons[neuronID].Fire()
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(IDSet, len(b.neurons))
	outputs := make([]SignalType, b.dna.motorIDs.Length())

	for i := 0; i < expectedSignals; i++ {
		select {
		case signal := <-b.sigChan:
			// May send an empty signal if the action potential threshold isn't met.
			if signal.isActive {
				for neuronID := range signal.source.Synapses {
					b.addPendingSignal(neuronID, signal.signal)
				}
			}
		case signal := <-b.motorChan:
			if signal.isActive {
				// Insert the signal at the index of the motor neuron ID.
				outputs[b.dna.motorIDs.GetIndex(signal.source.ID)] = signal.signal
			}
		}
	}
	return outputs
}

func (b Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.neurons[neuronID].ReceiveSignal(sig)
	b.pendingSignals[neuronID] = member
}
