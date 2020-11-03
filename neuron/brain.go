package neuron

import (
	"fmt"
	"strings"
	"sync"
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
	Snippets map[IDType]*Snippet
	NextID   IDType

	Seeds map[IDType]SignalType

	VisionIDs *IndexedIDs
	MotorIDs  *IndexedIDs
}

func NewDNA() *DNA {
	return &DNA{
		Snippets:  make(map[IDType]*Snippet, 0),
		NextID:    0,
		Seeds:     make(map[IDType]SignalType, 0),
		VisionIDs: NewIndexedIDs(),
		MotorIDs:  NewIndexedIDs(),
	}
}

func (src *DNA) DeepCopy() *DNA {
	dst := &DNA{
		Snippets:  make(map[IDType]*Snippet, len(src.Snippets)),
		NextID:    src.NextID,
		Seeds:     make(map[IDType]SignalType, len(src.Seeds)),
		VisionIDs: src.VisionIDs.Copy(),
		MotorIDs:  src.MotorIDs.Copy(),
	}

	for id, snippet := range src.Snippets {
		s := MakeSnippetOp(id, snippet.Op)
		for syn := range snippet.Synapses {
			s.AddSynapse(syn)
		}
		dst.Snippets[id] = s
	}

	for id, seed := range src.Seeds {
		dst.Seeds[id] = seed
	}

	return dst
}

func (d *DNA) AddSnippet(opVal int) *Snippet {
	s := MakeSnippet(d.NextID, opVal)
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
	sortedSnips := make([]*Snippet, d.NextID)
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

		s += fmt.Sprintf("%d:%d", id, snip.Op)

		if seed, exists := d.Seeds[id]; exists {
			s += fmt.Sprintf("<%d", seed)
		}

		if len(snip.Synapses) > 0 {
			s += "["
			sortedSyns := make([]bool, d.NextID)
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
	s = strings.TrimRight(s, " ")
	return s
}

type pathNode struct {
	id          IDType
	downstream  map[IDType]*pathNode
	motorOutput bool
}

type SignalPathway struct {
	root *pathNode

	mu           sync.Mutex
	currentLevel []pathNode
}

func NewSignalPathway() *SignalPathway {
	s := &SignalPathway{
		mu:           sync.Mutex{},
		currentLevel: make([]pathNode, 1),
	}
	s.currentLevel[0] = pathNode{
		id:          -1,
		downstream:  make(map[IDType]*pathNode, 0),
		motorOutput: true,
	}
	s.root = &s.currentLevel[0]
	return s
}

func (s *SignalPathway) addNode(srcID, dstID IDType) {
	for _, node := range s.currentLevel {
		if node.id == srcID {

			node.downstream[dstID]
		}
	}
}

// Brain docs
type Brain struct {
	dna     *DNA
	neurons map[IDType]*Neuron

	pendingSignals map[IDType][]SignalType
	sigChan        chan Signal
	motorChan      chan Signal

	signalPathway *SignalPathway
}

func Flourish(dna *DNA) *Brain {
	b := &Brain{
		dna:     dna,
		neurons: make(map[IDType]*Neuron, len(dna.Snippets)),

		pendingSignals: make(map[IDType][]SignalType, len(dna.Snippets)),
		sigChan:        make(chan Signal),
		motorChan:      make(chan Signal),
	}

	for id, snip := range dna.Snippets {
		// Select which signal channel should be injected. Motor neurons fire on a
		// different channel.
		selectedChan := b.sigChan
		if dna.MotorIDs.HasID(id) {
			selectedChan = b.motorChan
		}

		b.neurons[id] = &Neuron{
			snip:     snip,
			sigChan:  selectedChan,
			isVision: dna.VisionIDs.HasID(id),
		}
	}

	for id, seed := range dna.Seeds {
		b.addPendingSignal(id, seed)
	}

	return b
}

func (b *Brain) SeeInput(sigs []SignalType) {
	for i, sig := range sigs {
		// Send the signal to the vision ID at the signal's index.
		b.addPendingSignal(b.dna.VisionIDs.GetId(i), sig)
	}
}

func (b *Brain) StepFunction() []SignalType {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	for neuronID, sigs := range b.pendingSignals {
		go b.neurons[neuronID].Fire(sigs)
	}

	// Create a separate map that will be merged with pendingSignals after all
	// firing is done. This avoids a race condition where a synapse would add
	// a pending signal to the map and then be cleared later if that neuron was
	// going to fire anyway.
	nextSignals := make(map[IDType][]SignalType, len(b.neurons))
	outputs := make([]SignalType, 0)

	for i := 0; i < expectedSignals; i++ {
		select {
		case signal := <-b.sigChan:
			// May receive an inactive signal if the firing threshold isn't met.
			if signal.isActive {
				b.clearPendingSignals(signal.source.ID)

				for neuronID := range signal.source.Synapses {
					// Queue up signals to be added to pendingSignals.
					nextSignals[neuronID] = append(nextSignals[neuronID], signal.signal)
				}
			}
		case signal := <-b.motorChan:
			if signal.isActive {
				b.clearPendingSignals(signal.source.ID)

				// Resize the outputs to handle the number of motor neurons.
				if len(outputs) == 0 {
					outputs = make([]SignalType, b.dna.MotorIDs.Length())
				}
				// Insert the signal at the index of the motor neuron ID.
				outputs[b.dna.MotorIDs.GetIndex(signal.source.ID)] = signal.signal
			}
		}
	}

	// Merge nextSignals into pendingSignals now that the step is over.
	for neuronID, signals := range nextSignals {
		for _, signal := range signals {
			b.addPendingSignal(neuronID, signal)
		}
	}

	return outputs
}

func (b Brain) startSignalingPathway(neuronID IDType, sig SignalType) {

	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}

func (b Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}

func (b Brain) clearPendingSignals(neuronID IDType) {
	b.pendingSignals[neuronID] = make([]SignalType, 0)
}
