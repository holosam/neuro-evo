package neuron

import (
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	d := NewDNA()
	v0 := d.AddSnippet(SENSORY, OR)
	d.SetSeed(v0, 0)

	v1 := d.AddSnippet(SENSORY, OR)
	d.SetSeed(v1, 0)

	m0 := d.AddSnippet(MOTOR, OR)
	d.AddSynapse(v0, m0)
	d.AddSynapse(v1, m0)

	return d
}

func TestIndexedIDs(t *testing.T) {
	x := NewIndexedIDs()
	if got, want := x.HasID(0), false; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}

	x.InsertID(5)
	if got, want := x.HasID(5), true; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}

	x.InsertID(10)
	if got, want := x.Length(), 2; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := x.GetId(1), 10; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := x.GetIndex(10), 1; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}

	x.InsertID(15)
	x.RemoveID(5)
	if got, want := x.HasID(5), false; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := x.GetIndex(10), 0; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := x.GetIndex(15), 1; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := x.Length(), 2; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSnippetEditing(t *testing.T) {
	dna := SimpleTestDNA()
	dna.AddSnippet(INTER, IFF) // id=3
	dna.DeleteSnippet(1)

	if dna.NeuronIDs[SENSORY].HasID(1) {
		t.Errorf("VisionIDs should not have id 1")
	}
	if got, want := len(dna.Snippets), 3; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
	if got, want := dna.NextID, 4; got != want {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestDNADeepCopy(t *testing.T) {
	orig := SimpleTestDNA()
	copy := orig.DeepCopy()
	if !reflect.DeepEqual(orig, copy) {
		t.Errorf("Want equal, orig: %v, copy: %v", orig, copy)
	}

	orig.AddSnippet(INTER, XOR)
	orig.AddSynapse(0, 1)
	if reflect.DeepEqual(orig, copy) {
		t.Errorf("Want not equal, orig: %v, copy: %v", orig, copy)
	}
}

func TestDNAPrettyPrint(t *testing.T) {
	want := "(V0)=0:2<0[2,]  (V1)=1:2<0[2,]  (M0)=2:2  "
	d := SimpleTestDNA()
	if got := d.PrettyPrint(); got != want {
		t.Errorf("Want %s, got %s", want, got)
	}
}

func TestBrainStep(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(INTER, OR)
	d.AddSnippet(INTER, FALSIFY)
	d.AddSynapse(0, 1)
	b := Flourish(d)

	b.addPendingInput(0, SignalType(1))
	b.addPendingInput(0, SignalType(2))

	if want, got := 1, len(b.pendingSignals); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}
	if want, got := 2, len(b.pendingSignals[0]); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}

	outputs := b.StepFunction()

	if want, got := 0, len(outputs); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}
	if want, got := 1, len(b.pendingSignals); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}

	if sources, exists := b.pendingSignals[1]; !exists {
		t.Errorf("Want %v, got %v", true, exists)
	} else {
		if want, got := 1, len(sources); want != got {
			t.Errorf("Want %d, got %d", want, got)
		}
		if source, exists2 := sources[0]; !exists2 {
			t.Errorf("Want %v, got %v", true, exists2)
		} else {
			if want, got := 2, len(source.sources); want != got {
				t.Errorf("Want %v, got %v", want, got)
			}
			if want, got := 0, source.neuronID; want != got {
				t.Errorf("Want %v, got %v", want, got)
			}
			if want, got := true, source.isActive; want != got {
				t.Errorf("Want %v, got %v", want, got)
			}
			if want, got := SignalType(3), source.Output; want != got {
				t.Errorf("Want %v, got %v", want, got)
			}
		}
	}

	// Ensure the pending signals aren't cleared without firing.
	b.StepFunction()
	if want, got := 1, len(b.pendingSignals); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}
}

func TestEyesight(t *testing.T) {
	b := Flourish(SimpleTestDNA())
	b.SeeInput([]SignalType{1, 2})

	// First step fires the vision neurons and pends for the motor neuron.
	if got, want := b.StepFunction(), make([]Signal, 0); !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}

	// Second step fires the motor neuron.
	if want, got := SignalType(3), b.StepFunction()[0].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSignalSeeds(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(INTER, OR)
	d.AddSnippet(MOTOR, OR)
	d.AddSynapse(0, 1)
	d.SetSeed(1, 8)

	b := Flourish(d)
	b.addPendingInput(0, SignalType(1))
	b.addPendingInput(0, SignalType(2))

	b.StepFunction()
	got := b.StepFunction()

	if want, got := SignalType(11), got[0].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSignalTraceBack(t *testing.T) {
	d := SimpleTestDNA()
	d.SetSeed(2, 8)

	b := Flourish(d)
	b.SeeInput([]SignalType{1, 2})

	b.StepFunction()
	output := b.StepFunction()

	if want, got := SignalType(11), output[0].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	s2 := output[0].sources
	if want, got := 3, len(s2); want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
	if want, got := SignalType(8), s2[-1].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
	if want, got := SignalType(1), s2[0].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
	if want, got := SignalType(2), s2[1].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	s0 := s2[0].sources
	if want, got := SignalType(0), s0[-1].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
	if want, got := SignalType(1), s0[-2].Output; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
}
