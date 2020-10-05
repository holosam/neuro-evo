package neuron

import (
	"fmt"
	"strings"

	"github.com/ulule/deepcopier"
)

type DNA struct {
	snippets map[IDType]*Snippet
	nextID   IDType

	vision *Snippet
	motor  *Snippet
}

func NewDNA() *DNA {
	d := DNA{
		snippets: make(map[IDType]*Snippet),
		nextID:   0,
	}
	return &d
}

func (src *DNA) DeepCopy() *DNA {
	dst := &DNA{}
	deepcopier.Copy(src).To(dst)
	return dst
	// for id, snip := range src.snippets {
	// 	synapses := make([]int, len(snip.Synapses))
	// 	synIndex := 0
	// 	for synapse := range snip.Synapses {
	// 		synapses[synIndex] = synapse
	// 		synIndex++
	// 	}
	// 	dst.snippets[id] = MakeSnippetOp(id, snip.Op, synapses...)
	// }
	// for id := range src.visionIDs {
	// 	dst.AddVisionId(id)
	// }
	// for id := range src.motorIDs {
	// 	dst.AddMotorId(id)
	// }
	// dst.nextID = src.nextID
	// return dst
}

// Add AddSynapse and RemoveSynapse here
// Then delete the addPendingSignal stuff
// Prevent hanging synapses!!

func (d *DNA) AddSnippet(opVal int) *Snippet {
	s := MakeSnippet(d.nextID, opVal)
	d.snippets[d.nextID] = s
	d.nextID++
	return s
}

func (d *DNA) DeleteSnippet(id int) {
	// Delete from vision and motor too
	delete(d.snippets, id)
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
		if _, exists := d.visionIDs[id]; exists {
			s += "(V)="
		}
		if _, exists := d.motorIDs[id]; exists {
			s += "(M)="
		}
		s += fmt.Sprintf("%d:%v", id, snip.Op)

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
	neurons map[IDType]*Neuron
	// vision  *Neuron
	// motor   *Neuron

	pendingSignals IDSet
	sigChan        chan Signal
	motorChan      chan Signal
}

func Flourish(dna *DNA) *Brain {
	b := Brain{
		neurons: make(map[IDType]*Neuron, len(dna.snippets)),

		pendingSignals: make(IDSet, len(dna.snippets)),
		sigChan:        make(chan Signal),
		motorChan:      make(chan Signal),
	}

	for snipID, snip := range dna.snippets {
		selectedChan := b.sigChan
		if _, exists := dna.motorIDs[snipID]; exists {
			selectedChan = b.motorChan
		}

		_, isVision := dna.visionIDs[snipID]

		b.neurons[snipID] = &Neuron{
			snip:           snip,
			sigChan:        selectedChan,
			isVision:       isVision,
			pendingSignals: make([]SignalType, 0),
		}
	}

	return &b
}

func (b *Brain) SeeInput(sigs []SignalType) {
	b.addPendingSignal(0, sigs[0])
	b.addPendingSignal(1, sigs[1])
}

func (b *Brain) StepFunction() []SignalType {
	// Track the number of expected signals to receive from channels.
	expectedSignals := len(b.pendingSignals)
	for neuronID := range b.pendingSignals {
		go b.neurons[neuronID].Fire()
	}
	// Clear pending signals before refilling.
	b.pendingSignals = make(IDSet, len(b.neurons))
	outputs := make([]SignalType, 0)

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
				outputs = append(outputs, signal.signal)
			}
		}
	}
	return outputs
}

func (b Brain) addPendingSignal(neuronID IDType, sig SignalType) {
	b.neurons[neuronID].ReceiveSignal(sig)
	b.pendingSignals[neuronID] = member
}
