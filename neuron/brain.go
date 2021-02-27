package neuron

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

// IndexedIDs is a way of tracking IDs in numerical order. Neurons IDs are
// unique regardless of the type, so it's useful to be able to reference, for
// example, the "first" vision neuron even when its IDs don't start at zero.
// The indices are always 0 to N, so when an ID is removed, all the following
// indices shift down by one.
type IndexedIDs struct {
	IDToIndex map[IDType]int
	IndexToID map[int]IDType
}

// NewIndexedIDs initializes a new struct.
func NewIndexedIDs() *IndexedIDs {
	return &IndexedIDs{
		IDToIndex: make(map[IDType]int),
		IndexToID: make(map[int]IDType),
	}
}

// InsertID adds a new ID at the end of the ordered set, and returns the index
// of the newly added ID.
func (x *IndexedIDs) InsertID(id IDType) int {
	if x.HasID(id) {
		log.Fatalf("Indexed IDs already has %d", id)
	}

	nextIndex := x.Length()
	x.IDToIndex[id] = nextIndex
	x.IndexToID[nextIndex] = id
	return nextIndex
}

// RemoveID finds the ID and removes it out of the order, and shifts the
// remaining IDs down one to maintain indices of 0 to N.
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

// HasID returns true if the input ID is being tracked.
func (x *IndexedIDs) HasID(id IDType) bool {
	_, ok := x.IDToIndex[id]
	return ok
}

// GetIndex returns the index of the input ID.
func (x *IndexedIDs) GetIndex(id IDType) int {
	return x.IDToIndex[id]
}

// GetID returns the ID at the input index.
func (x *IndexedIDs) GetID(index int) IDType {
	return x.IndexToID[index]
}

// Length is the number of IDs that are being tracked.
func (x *IndexedIDs) Length() int {
	return len(x.IDToIndex)
}

// Copy makes a deep copy of the maps and returns a pointer to a new struct.
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

// SynapseTracker holds the all the synapses (connections) in a brain.
// The fields are conveniences for referencing the synapses either by the
// source neuron, or by the ID of the synapse itself.
type SynapseTracker struct {
	// map[synID] -> Synapse
	idMap map[IDType]Synapse
	// map[src] -> set[synID]
	srcMap map[IDType]IDSet
	nextID IDType
}

// NewSynapseTracker inits a new struct that can be added to.
func NewSynapseTracker() *SynapseTracker {
	return &SynapseTracker{
		idMap:  make(map[IDType]Synapse),
		srcMap: make(map[IDType]IDSet),
		nextID: 0,
	}
}

// TrackSynapse adds a new synapse to the map fields so it can easily be
// referenced later.
func (s *SynapseTracker) TrackSynapse(synID, src, dst IDType) IDType {
	syn := Synapse{
		src: src,
		dst: dst,
	}
	s.idMap[synID] = syn

	if _, ok := s.srcMap[src]; !ok {
		s.srcMap[src] = make(IDSet)
	}
	s.srcMap[src][synID] = member

	// Update the nextID to always be an unused number.
	if synID >= s.nextID {
		s.nextID = synID + 1
	}
	return synID
}

// AddNewSynapse tracks a synapse at the next available ID number.
func (s *SynapseTracker) AddNewSynapse(src, dst IDType) IDType {
	return s.TrackSynapse(s.nextID, src, dst)
}

// RemoveSynapse takes an ID un-tracks the associated synapse.
func (s *SynapseTracker) RemoveSynapse(id IDType) {
	syn := s.idMap[id]
	delete(s.srcMap[syn.src], id)
	if len(s.srcMap[syn.src]) == 0 {
		delete(s.srcMap, syn.src)
	}
	delete(s.idMap, id)
}

// AllDsts returns a set of IDs containing each neuron ID that the input neuron
// has connections to.
func (s *SynapseTracker) AllDsts(src IDType) IDSet {
	synIDs, ok := s.srcMap[src]
	if !ok {
		return make(IDSet, 0)
	}
	dsts := make(IDSet, len(synIDs))
	for synID := range synIDs {
		dsts[s.idMap[synID].dst] = member
	}
	return dsts
}

// FindID returns the synapse ID of the connection between the two input
// neurons.
func (s *SynapseTracker) FindID(src, dst IDType) (IDType, error) {
	for synID := range s.srcMap[src] {
		if s.idMap[synID].dst == dst {
			return synID, nil
		}
	}

	return 0, fmt.Errorf("non-existent synapse src=%d,dst=%d", src, dst)
}

// DeepCopy returns a copy of all the fields in a new struct.
func (s *SynapseTracker) DeepCopy() *SynapseTracker {
	dst := NewSynapseTracker()

	for synID, syn := range s.idMap {
		dst.idMap[synID] = Synapse{src: syn.src, dst: syn.dst}
	}

	for neuronID, synIDs := range s.srcMap {
		dst.srcMap[neuronID] = make(IDSet, len(synIDs))
		for synID := range synIDs {
			dst.srcMap[neuronID][synID] = member
		}
	}

	dst.nextID = s.nextID
	return dst
}

// Conglomerate represents the superset of many different brains.
// The conglomerate is essentialy a brain itself (many neurons and synapses),
// but it only holds the structure of the neurons without the ops and seeds.
// Aligning brains (graphs) with each other is a very difficult problem,
// especially when the structures get larger and continue diverging. So, the
// purpose of having one Conglomerate for evolution is that during the
// crossover step when brains "breed" children, their neurons and synapses are
// superimposed on the Conglomerate so that it's always clear how their neurons
// align with each other.
type Conglomerate struct {
	NeuronIDs map[NeuronType]*IndexedIDs
	Synapses  *SynapseTracker
}

// NewConglomerate inits a new conglomerate struct that's ready to be added to.
func NewConglomerate() *Conglomerate {
	c := &Conglomerate{
		NeuronIDs: make(map[NeuronType]*IndexedIDs, len(NeuronTypes)),
		Synapses:  NewSynapseTracker(),
	}
	for _, nType := range NeuronTypes {
		c.NeuronIDs[nType] = NewIndexedIDs()
	}
	return c
}

// AddVisionAndMotor sets up the Congolmerate to with a certain number of
// vision and motor neurons to match the expected number of inputs and outputs.
// The number of inputs and outputs shouldn't change during the evolution
// process.
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

// AddInterNeuron adds a neuron along a synapse. This means that if neuron #1
// has a synapse to neuron #2, then neuron #3 will be added in between 1 and 2,
// creating synapses from 1->3 and 3->2, in addition to leaving the original
// synapse from 1->2.
func (c *Conglomerate) AddInterNeuron(synID IDType) IDType {
	syn := c.Synapses.idMap[synID]
	newID := c.NeuronIDs[SENSE].Length() + c.NeuronIDs[MOTOR].Length() + c.NeuronIDs[INTER].Length()
	c.NeuronIDs[INTER].InsertID(newID)
	c.Synapses.AddNewSynapse(syn.src, newID)
	c.Synapses.AddNewSynapse(newID, syn.dst)
	return newID
}

// GetNeuronType returns the NeuronType of the input neuron ID.
func (c *Conglomerate) GetNeuronType(id IDType) NeuronType {
	for _, nType := range NeuronTypes {
		if c.NeuronIDs[nType].HasID(id) {
			return nType
		}
	}
	log.Fatalf("Non-existent neuron %d passed to GetNeuronType", id)
	return -1
}

// DNA references neurons and synapses which make up the genetic code of a
// brain. The DNA is a subset of the Conglomerate, so it can never have neurons
// or synapses that aren't in the Conglomerate. It also contains actual Neuron
// references, which have ops and seeds.
type DNA struct {
	Source   *Conglomerate
	Neurons  map[IDType]*Neuron
	Synpases *SynapseTracker
}

// NewDNA initializes a new DNA struct, pointing to its source of IDs which is
// a Conglomerate.
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

func (src *DNA) DeepCopy() *DNA {
	dst := NewDNA(src.Source)
	for neuronID, neuron := range src.Neurons {
		dst.Neurons[neuronID] = neuron.Copy()
	}
	dst.Synpases = src.Synpases.DeepCopy()
	return dst
}

// PrettyPrint returns a formatted string of all neurons and synapses in the
// DNA, which is useful for debugging.
func (d *DNA) PrettyPrint() string {
	var sb strings.Builder

	for _, nType := range NeuronTypes {
		nTypeChar := "I"
		if nType == SENSE {
			nTypeChar = "V"
		} else if nType == MOTOR {
			nTypeChar = "M"
		}
		for index := 0; index < d.Source.NeuronIDs[nType].Length(); index++ {
			neuronID := d.Source.NeuronIDs[nType].GetID(index)
			neuron, ok := d.Neurons[neuronID]
			if !ok {
				continue
			}

			sb.WriteString(fmt.Sprintf("%d (%s%d) = op%d", neuronID, nTypeChar, index, neuron.Op))
			if neuron.HasSeed {
				sb.WriteString(fmt.Sprintf(" <%d>", neuron.Seed))
			}

			sortedDstIDs := make([]IDType, 0)
			for dst := range d.Synpases.AllDsts(neuronID) {
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

const NullRune = 0

type brainOutput struct {
	signalString []SignalType
	isTerminated bool
}

// Brain docs
type Brain struct {
	dna            *DNA
	pendingSignals map[IDType][]SignalType
	// outputSignals is a map instead of slice to tell which motor neurons have
	// received and set an output.
	outputSignals []brainOutput
}

func Flourish(dna *DNA) *Brain {
	return &Brain{
		dna:            dna,
		pendingSignals: make(map[IDType][]SignalType, len(dna.Neurons)),
		outputSignals:  make([]brainOutput, dna.Source.NeuronIDs[MOTOR].Length()),
	}
}

// [][]SignalType can come from a single proto message in the future.
func (b *Brain) Fire(inputs [][]SignalType) [][]SignalType {
	inputStringIndex := 0

	// Cut off firing once it's very likely the output won't be generated.
	for step := 0; step < 100; step++ {
		// If there are any input signals left, add them to pendingSignals.
		for visionIndex, inputString := range inputs {
			if inputStringIndex > len(inputString) {
				continue
			}

			var inputSignal SignalType
			if inputStringIndex < len(inputString) {
				inputSignal = inputString[inputStringIndex]
			} else if inputStringIndex == len(inputString) {
				// Send a null termination to this vision neuron, signalling that no
				// more input will be coming on this action.
				inputSignal = NullRune
			}
			b.addPendingSignal(b.dna.Source.NeuronIDs[SENSE].GetId(visionIndex), inputSignal)
		}
		inputStringIndex++

		b.stepFunction()

		// Check if all of the output is ready to be returned.
		allTerminated := true
		for _, brainOutput := range b.outputSignals {
			if !brainOutput.isTerminated {
				allTerminated = false
			}
		}
		if allTerminated {
			break
		}
	}

	outputs := make([][]SignalType, len(b.outputSignals))
	for motorIndex, brainOutput := range b.outputSignals {
		// Only terminated outputs are returned.
		if brainOutput.isTerminated {
			outputs[motorIndex] = make([]SignalType, len(brainOutput.signalString))
			copy(outputs[motorIndex], brainOutput.signalString)
		} else {
			outputs[motorIndex] = make([]SignalType, 0)
		}
	}

	// Clear the output after it's used to make way for a new action.
	b.outputSignals = make([]brainOutput, b.dna.Source.NeuronIDs[MOTOR].Length())

	return outputs
}

func (b *Brain) stepFunction() {
	// Create a separate map that will be merged with pendingSignals after all
	// firing is done. This avoids a race condition where a synapse would add
	// a pending signal to the map and then be cleared later if that neuron fires
	// too.
	nextPending := make(map[IDType][]SignalType, len(b.dna.Neurons))

	for neuronID, inputs := range b.pendingSignals {
		numInputs := len(inputs)
		if b.dna.Neurons[neuronID].HasSeed {
			numInputs++
		}

		// Neurons fire when they have at least 2 signals.
		if numInputs < 2 {
			continue
		}

		neuron := b.dna.Neurons[neuronID]
		output := neuron.Fire(inputs)
		// fmt.Printf("firing neuron %d %+v with inputs %v and got output: %d\n", neuronID, neuron, inputs, output)

		// Clear this neuron's pending signals now that it has fired.
		// It's okay to edit the underlying map while iterating.
		delete(b.pendingSignals, neuronID)

		if b.dna.Source.NeuronIDs[MOTOR].HasID(neuronID) {
			motorIndex := b.dna.Source.NeuronIDs[MOTOR].GetIndex(neuronID)
			if !b.outputSignals[motorIndex].isTerminated {
				if output == NullRune {
					// A value of 0 is the termination character to cease listening for
					// output on this neuron.
					b.outputSignals[motorIndex].isTerminated = true
				} else {
					b.outputSignals[motorIndex].signalString = append(b.outputSignals[motorIndex].signalString, output)
				}
			}
			// fmt.Printf("output signals: %v+\n", b.outputSignals)
		}

		// Queue up signal for all downstream neurons.
		for dst := range b.dna.Synpases.AllDsts(neuronID) {
			nextPending[dst] = append(nextPending[dst], output)
		}
	}

	// Merge in nextPending now that the step is over.
	for neuronID, signals := range nextPending {
		for _, sig := range signals {
			// fmt.Printf("new pending signal %d for %d\n", sig, neuronID)
			b.addPendingSignal(neuronID, sig)
		}
	}
}

func (b *Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}
