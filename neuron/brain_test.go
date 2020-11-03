package neuron

import (
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	d := NewDNA()
	d.AddSnippet(2).AddSynapse(2)
	d.AddVisionID(0)

	d.AddSnippet(2).AddSynapse(2)
	d.AddVisionID(1)

	d.AddSnippet(2)
	d.AddMotorID(2)

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
	dna.AddSnippet(5) // id=3
	dna.DeleteSnippet(1)

	if dna.VisionIDs.HasID(1) {
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
	orig.SetSeed(0, 5)
	copy := orig.DeepCopy()
	if !reflect.DeepEqual(orig, copy) {
		t.Errorf("Want equal, orig: %v, copy: %v", orig, copy)
	}

	orig.AddSnippet(4)
	orig.AddSynapse(0, 1)
	if reflect.DeepEqual(orig, copy) {
		t.Errorf("Want not equal, orig: %v, copy: %v", orig, copy)
	}
}

func TestDNAPrettyPrint(t *testing.T) {
	want := "(V0)=0:2[2]  (V1)=1:2<9[2]  (M0)=2:2"
	d := SimpleTestDNA()
	d.SetSeed(1, 9)
	if got := d.PrettyPrint(); got != want {
		t.Errorf("Want %s, got %s", want, got)
	}
}

func TestBrainStep(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(2).AddSynapse(1)
	d.AddSnippet(7)
	b := Flourish(d)

	b.addPendingSignal(0, SignalType(1))
	b.addPendingSignal(0, SignalType(2))
	moves := b.StepFunction()

	if want, got := 0, len(moves); want != got {
		t.Errorf("Want %d, got %d", want, got)
	}

	wantMap := make(map[IDType][]SignalType, 2)
	wantMap[0] = make([]SignalType, 0)
	wantMap[1] = append(wantMap[1], 3)
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	// Ensure the pending signals aren't cleared without firing.
	b.StepFunction()
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}
}

func TestEyesight(t *testing.T) {
	b := Flourish(SimpleTestDNA())
	b.SeeInput([]SignalType{1, 2})

	// First step fires the vision neurons and pends for the motor neuron.
	if got, want := b.StepFunction(), make([]SignalType, 0); !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}

	// Second step fires the motor neuron.
	want := make([]SignalType, 1)
	want[0] = 3
	if got := b.StepFunction(); !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSignalSeeds(t *testing.T) {
	d := NewDNA()
	d.AddSnippet(2).AddSynapse(1)
	d.AddSnippet(2)
	d.AddMotorID(1)
	d.SetSeed(1, 8)

	b := Flourish(d)
	b.addPendingSignal(0, SignalType(1))
	b.addPendingSignal(0, SignalType(2))

	b.StepFunction()
	got := b.StepFunction()

	want := make([]SignalType, 1)
	want[0] = 11
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Want %v, got %v", want, got)
	}
	t.Errorf("double check")
}
