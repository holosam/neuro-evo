package neuron

import (
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	c := NewConglomerate()
	c.AddVisionAndMotor(2, 1)

	d := NewDNA(c)
	d.SetNeuron(0, OR)
	d.SetSeed(0, 0)
	d.SetNeuron(1, OR)
	d.SetSeed(1, 0)

	d.SetNeuron(2, OR)
	d.AddSynapse(0)
	d.AddSynapse(1)
	return d
}

func EqualIndexedIDs(got, want *IndexedIDs) bool {
	return reflect.DeepEqual(want.IDToIndex, got.IDToIndex) && reflect.DeepEqual(want.IndexToID, got.IndexToID)
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

func TestSynapseTracking(t *testing.T) {
	s := NewSynapseTracker()
	s.AddNewSynapse(0, 1)
	s.TrackSynapse(5, 0, 2)
	s.AddNewSynapse(1, 2)

	s.AddNewSynapse(3, 4)
	s.RemoveSynapse(7)

	expectedIDMap := make(map[IDType]Synapse)
	expectedIDMap[0] = Synapse{src: 0, dst: 1}
	expectedIDMap[5] = Synapse{src: 0, dst: 2}
	expectedIDMap[6] = Synapse{src: 1, dst: 2}
	if got := s.idMap; !reflect.DeepEqual(expectedIDMap, got) {
		t.Errorf("Want %v, got %v", expectedIDMap, got)
	}

	expectedDstMap := make(map[IDType]IDSet)
	expectedDstMap[0] = make(IDSet)
	expectedDstMap[1] = make(IDSet)
	expectedDstMap[0][1] = member
	expectedDstMap[0][2] = member
	expectedDstMap[1][2] = member
	if got := s.dstMap; !reflect.DeepEqual(expectedDstMap, got) {
		t.Errorf("Want %v, got %v", expectedDstMap, got)
	}

	if want, got := 8, s.nextID; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}
}

func TestSnippetEditing(t *testing.T) {
	c := NewConglomerate()
	c.AddVisionAndMotor(2, 2)

	expectedVisionIDs := NewIndexedIDs()
	expectedVisionIDs.InsertID(0)
	expectedVisionIDs.InsertID(1)
	expectedMotorIDs := NewIndexedIDs()
	expectedMotorIDs.InsertID(2)
	expectedMotorIDs.InsertID(3)
	if got, want := c.NeuronIDs[SENSE], expectedVisionIDs; !EqualIndexedIDs(got, want) {
		t.Errorf("Expected equal vision ids, got %v, want %v", got, want)
	}
	if got, want := c.NeuronIDs[MOTOR], expectedMotorIDs; !EqualIndexedIDs(got, want) {
		t.Errorf("Expected equal motor ids, got %v, want %v", got, want)
	}

	expectedSyns := make(map[IDType]Synapse)
	expectedSyns[0] = Synapse{src: 0, dst: 2}
	expectedSyns[1] = Synapse{src: 1, dst: 2}
	expectedSyns[2] = Synapse{src: 0, dst: 3}
	expectedSyns[3] = Synapse{src: 1, dst: 3}
	if got, want := c.Synapses.idMap, expectedSyns; !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected equal synapses, got %v, want %v", got, want)
	}

	c.AddInterNeuron(3)

	expectedInterIDs := NewIndexedIDs()
	expectedInterIDs.InsertID(4)
	if got, want := c.NeuronIDs[INTER], expectedInterIDs; !EqualIndexedIDs(got, want) {
		t.Errorf("Expected equal inter ids, got %v, want %v", got, want)
	}

	expectedSyns[4] = Synapse{src: 1, dst: 4}
	expectedSyns[5] = Synapse{src: 4, dst: 3}
	if got, want := c.Synapses.idMap, expectedSyns; !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected equal synapses, got %v, want %v", got, want)
	}
}

func TestDNAPrettyPrint(t *testing.T) {
	want := "0 (V0) = op2 <0> [2]\n1 (V1) = op2 <0> [2]\n2 (M0) = op2\n"
	if got := SimpleTestDNA().PrettyPrint(); got != want {
		t.Errorf("Want %s, got %s", want, got)
	}
}

func TestBrainStep(t *testing.T) {
	c := NewConglomerate()
	c.NeuronIDs[INTER].InsertID(0)
	c.NeuronIDs[INTER].InsertID(1)
	c.Synapses.AddNewSynapse(0, 1)

	d := NewDNA(c)
	d.SetNeuron(0, OR)
	d.SetSeed(0, 1)
	d.SetNeuron(1, FALSIFY)
	d.AddSynapse(0)
	b := Flourish(d)

	b.addPendingSignal(0, SignalType(2))

	wantMap := make(map[IDType][]SignalType, 2)
	wantMap[0] = []SignalType{2}
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	isDone := b.StepFunction()
	if want, got := false, isDone; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	delete(wantMap, 0)
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
