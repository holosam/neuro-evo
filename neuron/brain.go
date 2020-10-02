package neuron

const dnaSep = "|"

type DNA struct {
	snips map[int]*Snippet

	visionIDs IntSet
	motorIDs  IntSet

	nextID int
}

func NewDNA() *DNA {
	d := DNA{
		snips:  make(map[int]*Snippet),
		nextID: 0,
	}
	return &d
}

func (d *DNA) AddSnippet(op OperatorType) *Snippet {
	s := MakeSnippet(op)
	d.snips[d.nextID] = s
	d.nextID++
	return s
}

func (d *DNA) DeleteSnippet(id int) {
	delete(d.snips, id)
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
	numNeurons := len(dna.snips) // + 2

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
		go b.neurons[neuronID].Fire(b.sigChan, sigs)
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(map[int][]SignalType, len(b.neurons))
	movements := make([]SignalType, 0)

	for i := 0; i < expectedSignals; i++ {
		select {
		case signal := <-b.sigChan:
			// May send an empty signal if the action potential threshold isn't met.
			if len(signal.synapses) > 0 {
				for neuronID := range signal.synapses {
					b.addPendingSignal(neuronID, signal.val)
				}
			}
		case signal := <-b.motorChan:
			// Don't have to check if synapses are empty, just matters if it fired.
			movements = append(movements, signal.val)
		}
	}
	return movements
}

func (b Brain) addPendingSignal(neuronID int, sig SignalType) {
	// Possible to have a hanging synapse. Not ideal
	if _, exists := b.neurons[neuronID]; !exists {
		return
	}
	b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], sig)
}
