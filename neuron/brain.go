package neuron

import "fmt"

const dnaSep = "|"

type DNA struct {
	snips map[int]*Snippet

	visionIDs IntSet
	motorIDs  IntSet

	nextID int
}

func NewDNA() *DNA {
	d := DNA{
		snips:     make(map[int]*Snippet),
		visionIDs: make(IntSet),
		motorIDs:  make(IntSet),
		nextID:    0,
	}
	return &d
}

func (d *DNA) AddSnippet(opVal int) *Snippet {
	s := MakeSnippet(opVal)
	d.snips[d.nextID] = s
	d.nextID++
	return s
}

func (d *DNA) DeleteSnippet(id int) {
	delete(d.snips, id)
}

func (d *DNA) AddVisionId(id int) {
	if _, exists := d.motorIDs[id]; exists {
		// Should probably fail here, but nbd for now.
		return
	}
	d.visionIDs[id] = member
}

func (d *DNA) AddMotorId(id int) {
	if _, exists := d.visionIDs[id]; exists {
		// Should probably fail here, but nbd for now.
		return
	}
	d.motorIDs[id] = member
}

// Brain docs
type Brain struct {
	neurons map[int]*Neuron

	visionIDs IntSet

	pendingSignals map[int][]SignalType

	sigChan   chan Signal
	motorChan chan Signal
}

func Flourish(dna *DNA) *Brain {
	numNeurons := len(dna.snips)

	b := Brain{
		neurons:        make(map[int]*Neuron, numNeurons),
		visionIDs:      dna.visionIDs,
		pendingSignals: make(map[int][]SignalType, numNeurons),
		sigChan:        make(chan Signal),
		motorChan:      make(chan Signal),
	}

	for snipID, snip := range dna.snips {
		selectedChan := b.sigChan
		if _, exists := dna.motorIDs[snipID]; exists {
			selectedChan = b.motorChan
		}

		isVision := false
		if _, exists := dna.visionIDs[snipID]; exists {
			isVision = true
		}

		b.neurons[snipID] = &Neuron{
			snip:     snip,
			sigChan:  selectedChan,
			isVision: isVision,
		}
	}

	return &b
}

func (b *Brain) SeeInput(sigs ...SignalType) {
	for _, sig := range sigs {
		for visionID := range b.visionIDs {
			b.addPendingSignal(visionID, sig)
		}
	}
}

func (b *Brain) StepFunction() []SignalType {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	for neuronID, sigs := range b.pendingSignals {
		go b.neurons[neuronID].Fire(sigs)
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(map[int][]SignalType, len(b.neurons))
	movements := make([]SignalType, 0)

	for i := 0; i < expectedSignals; i++ {
		select {
		case signal := <-b.sigChan:
			fmt.Printf("Channel normal signal %d\n", signal.val)
			// May send an empty signal if the action potential threshold isn't met.
			if signal.active {
				for neuronID := range signal.synapses {
					b.addPendingSignal(neuronID, signal.val)
				}
			}
		case signal := <-b.motorChan:
			fmt.Printf("Channel motor signal %d\n", signal.val)
			if signal.active {
				movements = append(movements, signal.val)
			}
		}
	}
	return movements
}

func (b Brain) addPendingSignal(neuronID int, sig SignalType) {
	// Possible to have a hanging synapse. Not ideal
	if _, exists := b.neurons[neuronID]; !exists {
		return
	}
	fmt.Printf("Adding pending signal id=%d sig=%d\n", neuronID, sig)
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}
