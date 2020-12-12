package neuron

import (
	"fmt"
	"log"
	"sort"
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

func (x *IndexedIDs) InsertID(id IDType) int {
	if x.HasID(id) {
		log.Fatalf("Indexed IDs already has %d", id)
	}

	nextIndex := x.Length()
	x.IDToIndex[id] = nextIndex
	x.IndexToID[nextIndex] = id
	return nextIndex
}

func (x *IndexedIDs) RemoveID(id IDType) {
	index, exists := x.IDToIndex[id]
	if !exists {
		log.Fatalf("Indexed IDs doesn't have id %d", id)
	}

	length := x.Length()
	for i := index; i < length-1; i++ {
		idToMove := x.IndexToID[i+1]
		x.IndexToID[i] = idToMove
		x.IDToIndex[idToMove] = i
	}

	delete(x.IDToIndex, id)
	delete(x.IndexToID, length-1)
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

type SynapseTracker struct {
	// map[synID] -> Synapse
	idMap map[IDType]Synapse
	// map[src] -> set[dst]
	dstMap map[IDType]IDSet
	nextID IDType
}

func NewSynapseTracker() *SynapseTracker {
	return &SynapseTracker{
		idMap:  make(map[IDType]Synapse),
		dstMap: make(map[IDType]IDSet),
		nextID: 0,
	}
}

func (s *SynapseTracker) AddNewSynapse(src, dst IDType) IDType {
	return s.TrackSynapse(s.nextID, src, dst)
}

func (s *SynapseTracker) TrackSynapse(synID, src, dst IDType) IDType {
	syn := Synapse{
		src: src,
		dst: dst,
	}
	s.idMap[synID] = syn

	if _, ok := s.dstMap[src]; !ok {
		s.dstMap[src] = make(IDSet)
	}
	s.dstMap[src][dst] = member

	if synID >= s.nextID {
		s.nextID = synID + 1
	}
	return synID
}

func (s *SynapseTracker) RemoveSynapse(id IDType) {
	syn := s.idMap[id]
	delete(s.dstMap[syn.src], syn.dst)
	if len(s.dstMap[syn.src]) == 0 {
		delete(s.dstMap, syn.src)
	}
	delete(s.idMap, id)
}

func (s *SynapseTracker) FindID(src, dst IDType) IDType {
	for synID, syn := range s.idMap {
		if src == syn.src && dst == syn.dst {
			return synID
		}
	}

	log.Fatalf("Non-existent synapse src=%d,dst=%d passed to SynapseTracker.FindID", src, dst)
	return -1
}

type Conglomerate struct {
	NeuronIDs map[NeuronType]*IndexedIDs
	Synapses  *SynapseTracker
}

func NewConglomerate() *Conglomerate {
	c := &Conglomerate{
		NeuronIDs: make(map[NeuronType]*IndexedIDs, len(neuronTypes)),
		Synapses:  NewSynapseTracker(),
	}
	for _, nType := range neuronTypes {
		c.NeuronIDs[nType] = NewIndexedIDs()
	}
	return c
}

func (c *Conglomerate) AddVisionAndMotor(numInputs int, numOutputs int) {
	id := 0
	for ; id < numInputs; id++ {
		c.NeuronIDs[SENSE].InsertID(id)
	}

	for ; id < numInputs+numOutputs; id++ {
		c.NeuronIDs[MOTOR].InsertID(id)
		for v := 0; v < numInputs; v++ {
			c.Synapses.AddNewSynapse(v, id)
		}
	}
}

func (c *Conglomerate) AddInterNeuron(synID IDType) IDType {
	syn := c.Synapses.idMap[synID]
	newID := c.NeuronIDs[SENSE].Length() + c.NeuronIDs[MOTOR].Length() + c.NeuronIDs[INTER].Length()
	c.NeuronIDs[INTER].InsertID(newID)
	c.Synapses.AddNewSynapse(syn.src, newID)
	c.Synapses.AddNewSynapse(newID, syn.dst)
	return newID
}

func (c *Conglomerate) GetNeuronType(id IDType) NeuronType {
	for _, nType := range neuronTypes {
		if c.NeuronIDs[nType].HasID(id) {
			return nType
		}
	}
	log.Fatalf("Non-existent neuron %d passed to GetNeuronType", id)
	return -1
}

type DNA struct {
	Source   *Conglomerate
	Neurons  map[IDType]*Neuron
	Synpases *SynapseTracker
}

func NewDNA(source *Conglomerate) *DNA {
	return &DNA{
		Source:   source,
		Neurons:  make(map[IDType]*Neuron),
		Synpases: NewSynapseTracker(),
	}
}

func (d *DNA) AddNeuron(id IDType, op OperatorType) {
	d.Neurons[id] = NewNeuron(op)
}

func (d *DNA) SetNeuron(id IDType, neuron *Neuron) {
	d.Neurons[id] = neuron.Copy()
}

func (d *DNA) AddSynapse(id IDType) {
	syn := d.Source.Synapses.idMap[id]
	d.Synpases.TrackSynapse(id, syn.src, syn.dst)
}

func (d *DNA) RemoveSynapse(id IDType) {
	d.Synpases.RemoveSynapse(id)
}

func (d *DNA) SetSeed(id IDType, seed SignalType) {
	d.Neurons[id].SetSeed(seed)
}

func (d *DNA) RemoveSeed(id IDType) {
	d.Neurons[id].RemoveSeed()
}

func (d *DNA) PrettyPrint() string {
	var sb strings.Builder

	for _, nType := range neuronTypes {
		nTypeChar := "I"
		if nType == SENSE {
			nTypeChar = "V"
		} else if nType == MOTOR {
			nTypeChar = "M"
		}
		for index := 0; index < d.Source.NeuronIDs[nType].Length(); index++ {
			neuronID := d.Source.NeuronIDs[nType].GetId(index)
			neuron := d.Neurons[neuronID]

			sb.WriteString(fmt.Sprintf("%d (%s%d) = op%d", neuronID, nTypeChar, index, neuron.op))
			if neuron.hasSeed {
				sb.WriteString(fmt.Sprintf(" <%d>", neuron.seed))
			}

			sortedDstIDs := make([]IDType, 0)
			for dst := range d.Synpases.dstMap[neuronID] {
				sortedDstIDs = append(sortedDstIDs, dst)
			}
			if len(sortedDstIDs) > 0 {
				sort.Slice(sortedDstIDs, func(i, j int) bool {
					return sortedDstIDs[i] < sortedDstIDs[j]
				})
				sb.WriteString(fmt.Sprintf(" %v", sortedDstIDs))
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Brain docs
type Brain struct {
	dna            *DNA
	pendingSignals map[IDType][]SignalType
	// outputSignals is a map instead of slice to tell which motor neurons have
	// received and set an output.
	outputSignals map[IDType]SignalType
}

func Flourish(dna *DNA) *Brain {
	return &Brain{
		dna:            dna,
		pendingSignals: make(map[IDType][]SignalType, len(dna.Neurons)),
		outputSignals:  make(map[IDType]SignalType, dna.Source.NeuronIDs[MOTOR].Length()),
	}
}

func (b *Brain) SeeInput(sigs []SignalType) {
	for i, sig := range sigs {
		fmt.Printf("input for signal %d is %d\n", i, sig)
		// Send the signal to the vision ID at the signal's index.
		b.addPendingSignal(b.dna.Source.NeuronIDs[SENSE].GetId(i), sig)
	}
	fmt.Printf("full pending signals %v\n", b.pendingSignals)
}

func (b *Brain) Output() []SignalType {
	output := make([]SignalType, len(b.outputSignals))
	for id, sig := range b.outputSignals {
		output[b.dna.Source.NeuronIDs[MOTOR].GetIndex(id)] = sig
	}
	// Clear the output after it's used to make way for a new action.
	b.outputSignals = make(map[IDType]SignalType, b.dna.Source.NeuronIDs[MOTOR].Length())
	return output
}

func (b *Brain) StepFunction() bool {
	// Create a separate map that will be merged with pendingSignals after all
	// firing is done. This avoids a race condition where a synapse would add
	// a pending signal to the map and then be cleared later if that neuron fires
	// too.
	nextPending := make(map[IDType][]SignalType, len(b.dna.Neurons))

	done := false
	for neuronID, inputs := range b.pendingSignals {
		numInputs := len(inputs)
		if b.dna.Neurons[neuronID].hasSeed {
			numInputs++
		}

		// Neurons fire when they have at least 2 signals.
		if numInputs < 2 {
			continue
		}

		neuron := b.dna.Neurons[neuronID]
		output := neuron.Fire(inputs)
		fmt.Printf("firing neuron %v with inputs %v and got output: %d\n", neuron, inputs, output)

		// Clear this neuron's pending signals now that it has fired.
		// It's okay to edit the underlying map while iterating.
		delete(b.pendingSignals, neuronID)

		if b.dna.Source.NeuronIDs[MOTOR].HasID(neuronID) {
			// Add the signal to the outputSignals only if it's the first time this
			// motor neuron has fired.
			if _, ok := b.outputSignals[neuronID]; !ok {
				b.outputSignals[neuronID] = output
			}
			fmt.Printf("output signals: %v\n", b.outputSignals)
			// Done is true if all motor neurons have an output.
			done = len(b.outputSignals) == b.dna.Source.NeuronIDs[MOTOR].Length()
		}

		// Queue up signal for all downstream neurons.
		for dst := range b.dna.Synpases.dstMap[neuronID] {
			nextPending[dst] = append(nextPending[dst], output)
		}
	}

	// Merge in nextPending now that the step is over.
	for neuronID, signals := range nextPending {
		for _, sig := range signals {
			b.addPendingSignal(neuronID, sig)
		}
	}

	return done
}

func (b *Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}
