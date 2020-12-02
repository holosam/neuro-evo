package neuron

import (
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	d := NewDNA()
	v0 := d.AddSnippet(SENSE, OR)
	d.SetSeed(v0, 0)

	v1 := d.AddSnippet(SENSE, OR)
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

	if dna.NeuronIDs[SENSE].HasID(1) {
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
	d.SetSeed(0, 1)
	d.AddSnippet(INTER, FALSIFY)
	d.AddSynapse(0, 1)
	b := Flourish(d)

	b.addPendingSignal(0, SignalType(2))

	wantMap := make(map[IDType][]SignalType, 2)
	wantMap[0] = []SignalType{1, 2}
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	isDone := b.StepFunction()
	if want, got := false, isDone; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	// The seed should be sticky for ID 0.
	wantMap[0] = []SignalType{1}
	wantMap[1] = []SignalType{3}
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	// Ensure the pending signals aren't cleared without firing.
	b.StepFunction()
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}
}

func TestEyesightAndMuscles(t *testing.T) {
	b := Flourish(SimpleTestDNA())
	b.SeeInput([]SignalType{1, 2})

	// First step fires the vision neurons and pends for the motor neuron.
	if want, got := false, b.StepFunction(); want != got {
		t.Fatalf("Want %v, got %v", want, got)
	}

	// Second step fires the motor neuron.
	if want, got := true, b.StepFunction(); want != got {
		t.Fatalf("Want %v, got %v", want, got)
	}
	if want, got := SignalType(3), b.Output()[0]; want != got {
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
	b.addPendingSignal(0, SignalType(1))
	b.addPendingSignal(0, SignalType(2))

	b.StepFunction()
	b.StepFunction()

	if want, got := SignalType(11), b.Output()[0]; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
}
