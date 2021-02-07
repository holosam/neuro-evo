package neuron

import (
	"fmt"
	"reflect"
	"testing"
)

// Two vision neurons pointing at a motor neuron.
func SimpleTestDNA() *DNA {
	c := NewConglomerate()
	c.AddVisionAndMotor(2, 1)

	d := NewDNA(c)
	d.AddNeuron(0, OR)
	d.SetSeed(0, 0)
	d.AddNeuron(1, OR)
	d.SetSeed(1, 0)

	d.AddNeuron(2, OR)
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
	if got, want := s.AddNewSynapse(0, 1), 0; got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
	if got, want := s.TrackSynapse(5, 0, 2), 5; got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
	if got, want := s.AddNewSynapse(1, 2), 6; got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	if got, want := s.AddNewSynapse(3, 4), 7; got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
	s.RemoveSynapse(7)

	expectedIDMap := make(map[IDType]Synapse)
	expectedIDMap[0] = Synapse{src: 0, dst: 1}
	expectedIDMap[5] = Synapse{src: 0, dst: 2}
	expectedIDMap[6] = Synapse{src: 1, dst: 2}
	if got := s.idMap; !reflect.DeepEqual(expectedIDMap, got) {
		t.Errorf("Want %v, got %v", expectedIDMap, got)
	}

	expectedSrcMap := make(map[IDType]IDSet)
	expectedSrcMap[0] = make(IDSet)
	expectedSrcMap[1] = make(IDSet)
	expectedSrcMap[0][0] = member
	expectedSrcMap[0][5] = member
	expectedSrcMap[1][6] = member
	if got := s.srcMap; !reflect.DeepEqual(expectedSrcMap, got) {
		t.Errorf("Want %v, got %v", expectedSrcMap, got)
	}

	if want, got := 8, s.nextID; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	expectedDsts := make(IDSet)
	expectedDsts[1] = member
	expectedDsts[2] = member
	if got := s.AllDsts(0); !reflect.DeepEqual(expectedDsts, got) {
		t.Errorf("Want %v, got %v", expectedDsts, got)
	}

	if got, want := len(s.AllDsts(99)), 0; want != got {
		t.Errorf("Want %v, got %v", want, got)
	}

	foundID, err := s.FindID(1, 2)
	if want := 6; err != nil || foundID != want {
		t.Errorf("Want %v, got %v", want, foundID)
	}

	_, err = s.FindID(9, 10)
	if err == nil {
		t.Errorf("Want error, got none")
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

func TestDeepCopy(t *testing.T) {
	d := SimpleTestDNA()
	if got, want := d.DeepCopy().PrettyPrint(), d.PrettyPrint(); got != want {
		t.Errorf("Got %s, want %s", got, want)
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
	d.AddNeuron(0, OR)
	d.SetSeed(0, 1)
	d.AddNeuron(1, FALSIFY)
	d.AddSynapse(0)
	b := Flourish(d)

	b.addPendingSignal(0, SignalType(2))

	wantMap := make(map[IDType][]SignalType, 2)
	wantMap[0] = []SignalType{2}
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	b.stepFunction()

	delete(wantMap, 0)
	wantMap[1] = []SignalType{3}
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}

	// Ensure the pending signals aren't cleared without firing.
	b.stepFunction()
	if !reflect.DeepEqual(wantMap, b.pendingSignals) {
		t.Errorf("Want %v, got %v", wantMap, b.pendingSignals)
	}
}

func TestBrainFire(t *testing.T) {
	b := Flourish(SimpleTestDNA())
	if got, want := b.Fire([][]SignalType{{1}, {2}}), [][]SignalType{{3}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}

	fmt.Printf("pending signals: %+v\n", b.pendingSignals)

	t.Errorf("error for logs")
}

// Create a circular brain that won't ever output to test if Fire will
// stop.
func TestCircularBrainFiring(t *testing.T) {
	c := NewConglomerate()
	c.AddVisionAndMotor(2, 1)
	c.AddInterNeuron(0)
	c.AddInterNeuron(1)
	c.AddInterNeuron(3)
	syn43 := c.Synapses.AddNewSynapse(4, 3)
	syn54 := c.Synapses.AddNewSynapse(5, 4)

	d := NewDNA(c)
	for neuronID := 0; neuronID < 6; neuronID++ {
		d.AddNeuron(neuronID, OR)
		if neuronID == 2 {
			continue
		}
		d.SetSeed(neuronID, 0)
	}

	d.AddSynapse(1)
	d.AddSynapse(2)
	d.AddSynapse(4)
	d.AddSynapse(6)
	d.AddSynapse(syn43)
	d.AddSynapse(syn54)

	fmt.Printf("Circular brain: %s\n", d.PrettyPrint())

	b := Flourish(d)
	got := b.Fire([][]SignalType{{1}, {2}})
	if want := [][]SignalType{{}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, got %v", want, got)
	}
}
