package neuron

const dnaSep = "|"

type DNA struct {
	snips map[int]Snippet

	// visionID int
	// motorID  int

	nextID int
}

func NewDNA() *DNA {
	d := DNA{
		snips:  make(map[int]Snippet),
		nextID: 0,
	}
	return &d
}

func (d *DNA) AddSnippet(op OperatorType) *Snippet {
	s := MakeSnippet(op)
	d.snips[d.nextID] = s
	d.nextID++
	return &s
}

func (d *DNA) DeleteSnippet(id int) {
	delete(d.snips, id)
}

// Brain docs
type Brain struct {
	neurons map[int]Neuron

	pendingSignals map[int][]SignalType

	sigChan chan Signal
}

func Flourish(dna *DNA) *Brain {
	numNeurons := len(dna.snips) // + 2

	b := Brain{
		neurons:        make(map[int]Neuron, numNeurons),
		pendingSignals: make(map[int][]SignalType, numNeurons),
		sigChan:        make(chan Signal),
	}

	for nIndex, snip := range dna.snips {
		b.neurons[nIndex] = Neuron{snip: snip}
	}

	return &b
}

func (b *Brain) StepFunction() {
	firingNeurons := 0
	for neuronID, sigs := range b.pendingSignals {
		// A neuron only fires when it receives at least 2 inputs
		if len(sigs) >= 2 {
			go b.neurons[neuronID].Fire(b.sigChan, sigs)
			firingNeurons++
		}
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(map[int][]SignalType, len(b.neurons))

	for i := 0; i < firingNeurons; i++ {
		signal := <-b.sigChan
		for neuronID := range signal.nIndicies {
			// Possible to have a hanging synapse. Not ideal
			if _, exists := b.neurons[neuronID]; !exists {
				continue
			}
			b.pendingSignals[neuronID] = append(b.pendingSignals[neuronID], signal.val)
		}
	}
}
