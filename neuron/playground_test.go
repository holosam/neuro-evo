package neuron

import (
	"fmt"
	"reflect"
	"testing"
)

func TestInitDNA(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		NumInputs:   2,
		NumOutputs:  1,
		NumVariants: 5,

		Mconf: MutationConfig{
			ChangeOp:  1.0,
			SetSeed:   0.5,
			UnsetSeed: 0.5,
		},
	})

	p.InitDNA()

	if got, want := len(p.codes), p.config.NumVariants; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	sampleDNA := p.codes[p.rnd.Intn(p.config.NumVariants)]
	expectedNeurons := make([]bool, 3)
	for neuronID := range sampleDNA.Neurons {
		expectedNeurons[neuronID] = true
	}
	if !reflect.DeepEqual(expectedNeurons, []bool{true, true, true}) {
		t.Errorf("Got neurons %v", expectedNeurons)
	}
	foundSyns := make([]bool, 2)
	for synID := range sampleDNA.Synpases.idMap {
		foundSyns[synID] = true
	}
	if !reflect.DeepEqual(foundSyns, []bool{true, true}) {
		t.Errorf("Got syns %v", foundSyns)
	}

	if p.codes[0].PrettyPrint() == p.codes[1].PrettyPrint() {
		t.Errorf("Expected DNA have different ops/seeds")
	}
}

/*
func TestSimulatePlayground(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		DnaSeedSnippets:  10,
		DnaSeedMutations: 10,

		NumSpecies:   10,
		Generations:  5,
		RoundsPerGen: 3,
		GenInputsFn: func(round int) []SignalType {
			return []SignalType{1, 2}
		},

		FitnessFn: func(inputs []SignalType, outputs []SignalType) ScoreType {
			if len(outputs) == 0 {
				return math.MaxUint64
			}
			return ScoreType(outputs[0])
		},
		NumParents: 3,

		Gconf: GenerationConfig{
			MaxSteps: 10,
		},
	})
	p.SeedRandDNA()
	arbitaryDNA := p.codes[0]

	p.SimulatePlayground()

	if arbitaryDNA == p.codes[0] {
		t.Errorf("Expected evolution, got nothing")
	}
}
*/

// func TestAddSynIfNotExists(t *testing.T) {
// 	p := NewPlayground(PlaygroundConfig{
// 		NumInputs:  2,
// 		NumOutputs: 1,
// 	})
// 	p.InitDNA()
// 	p.source.AddInterNeuron(0)
// 	p.source.AddInterNeuron(1)
// 	//    0    1
// 	// (2)|    |(4)
// 	//    3    4
// 	//   (3)\/(5)
// 	//0-(0)- 2 -(1)-1

// 	candidates := make(map[IDType]IDSet, 0)
// 	expected := make(map[IDType]IDSet, 0)

// 	// DstType is SENSE, do not add.
// 	p.addSynIfNotExists(0, 1, SENSE, SENSE, candidates)
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}

// 	// Syn exists, do not add.
// 	p.addSynIfNotExists(0, 3, SENSE, INTER, candidates)
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}

// 	// SrcType is MOTOR, do not add.
// 	p.addSynIfNotExists(2, 4, MOTOR, INTER, candidates)
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}

// 	// Syn doesn't exist, add it.
// 	p.addSynIfNotExists(0, 4, SENSE, INTER, candidates)
// 	expected[0] = make(IDSet, 1)
// 	expected[0][4] = member
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}

// 	// Syn doesn't exist, add it.
// 	p.addSynIfNotExists(4, 3, INTER, INTER, candidates)
// 	expected[4] = make(IDSet, 1)
// 	expected[4][3] = member
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}

// 	// Already added, nothing should change.
// 	p.addSynIfNotExists(4, 3, INTER, INTER, candidates)
// 	if !reflect.DeepEqual(candidates, expected) {
// 		t.Errorf("Got %v, want %v", candidates, expected)
// 	}
// }

func CreateTestConglomerate() *Playground {
	p := NewPlayground(PlaygroundConfig{
		NumInputs:  2,
		NumOutputs: 1,
	})
	p.InitDNA() // Makes vision N0, N1, and motor N2 with Syn0 and Syn1.

	p.source.AddInterNeuron(0) // Syn0 connects N0 to N2, makes N3 (Syn2 and Syn3)
	p.source.AddInterNeuron(1) // Syn1 connects N1 to N2, makes N4 (Syn4 and Syn5)
	p.source.AddInterNeuron(5) // Syn5 connects N4 to N2, makes N5 (Syn6 and Syn7)

	p.source.Synapses.AddNewSynapse(0, 4) // makes Syn8 between N0 and N4

	/*
					0 (8)  1
			 (2)|  \  /(4)
					3   4
					|   | (6)
			 (3)|   5
					| / (7)
					2

		Also:
		0->2 Syn0
		1->2 Syn1
		4->2 Syn5
	*/

	return p
}

func testAddToIDSet(idset IDSet, ids ...IDType) IDSet {
	for _, id := range ids {
		idset[id] = member
	}
	return idset
}

func testMakeIDSet(ids ...IDType) IDSet {
	return testAddToIDSet(make(IDSet), ids...)
}

func TestNearbyNeurons(t *testing.T) {
	p := CreateTestConglomerate()

	expected := make(map[IDType]IDSet)
	expected[0] = testMakeIDSet(2, 3, 4)
	expected[1] = testMakeIDSet(2, 4)
	expected[2] = testMakeIDSet(0, 1, 3, 4, 5)
	expected[3] = testMakeIDSet(0, 2)
	expected[4] = testMakeIDSet(0, 1, 2, 5)
	expected[5] = testMakeIDSet(2, 4)

	if got := p.nearbyNeurons(1); !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %v, want %v", got, expected)
	}

	fmt.Printf("-------- next test ----------\n")

	// Because everything is 1 hop away from N2, then at 2 hops everything is
	// "nearby" everything else.
	expected[0] = testAddToIDSet(expected[0], 1, 5)
	expected[1] = testAddToIDSet(expected[1], 0, 3, 5)
	expected[3] = testAddToIDSet(expected[3], 1, 4, 5)
	expected[4] = testAddToIDSet(expected[4], 3)
	expected[5] = testAddToIDSet(expected[5], 0, 1, 3)

	if got := p.nearbyNeurons(2); !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %v, want %v", got, expected)
	}
}

// func TestDownstreamNeurons(t *testing.T) {
// 	p := CreateTestConglomerate()

// 	expected := make(IDSet)
// 	expected[2] = member // Syn0 connects N0 to N2
// 	expected[3] = member // Syn2 connects N0 to N3
// 	expected[4] = member // Syn8 connects N0 to N4

// 	// One hop away.
// 	if got := p.downstreamNeurons(0, 1); !reflect.DeepEqual(got, expected) {
// 		t.Errorf("Got %v, want %v", got, expected)
// 	}

// 	// Two hops away.
// 	expected[5] = member // Syn7 connects N4 to N5
// 	if got := p.downstreamNeurons(0, 2); !reflect.DeepEqual(got, expected) {
// 		t.Errorf("Got %v, want %v", got, expected)
// 	}
// }

func TestMutateDNAStructure(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{
		NumInputs:   2,
		NumOutputs:  1,
		NumVariants: 1,

		Mconf: MutationConfig{
			// Will round up to 1
			NeuronExpansion: 0.01,
			// Ensure both synapses are added so the test is deterministic.
			SynapseExpansion: 2.0,

			AddNeuron: 1.0,
			// Starts at 0 so the removed synapse doesn't get immediately added back.
			AddSynapse: 0.0,

			ChangeOp:  1.0,
			SetSeed:   0.5,
			UnsetSeed: 0.5,
		},
	})

	p.InitDNA()
	dna := p.codes[0]

	// There are 2 vision and 1 motor neurons.
	// Add an inter neuron from V0->M0
	newInterID := p.source.AddInterNeuron(0)
	p.mutateDNAStructure(dna)

	if _, ok := dna.Neurons[newInterID]; !ok {
		t.Fatalf("Expected neuron %d to be added", newInterID)
	}

	foundSyns := make([]bool, 4)
	for synID := range dna.Synpases.idMap {
		foundSyns[synID] = true
	}
	// The V0->M0 syn should be removed when I0 is added in between.
	if !reflect.DeepEqual(foundSyns, []bool{false, true, true, true}) {
		t.Errorf("Got syns %v", foundSyns)
	}

	// Then add a synapse from V1->I0
	p.config.Mconf.AddSynapse = 1.0
	p.source.Synapses.AddNewSynapse(1, newInterID)
	p.mutateDNAStructure(dna)

	foundSyns = make([]bool, 5)
	for synID := range dna.Synpases.idMap {
		foundSyns[synID] = true
	}
	if !reflect.DeepEqual(foundSyns, []bool{true, true, true, true, true}) {
		t.Errorf("Got syns %v", foundSyns)
	}
}

func TestMutateNeurons(t *testing.T) {
	dna := SimpleTestDNA()
	p := NewPlayground(PlaygroundConfig{
		Mconf: MutationConfig{
			ChangeOp:  1.0,
			SetSeed:   0.5,
			UnsetSeed: 0.5,
		},
	})

	p.mutateNeurons(dna)
	if got, want := dna.PrettyPrint(), SimpleTestDNA().PrettyPrint(); got == want {
		t.Errorf("Got: %v\n want: %v", got, want)
	}
}

func TestHelperFns(t *testing.T) {
	p := NewPlayground(PlaygroundConfig{})

	if got, want := p.mutationOccurs(1.0), true; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := p.mutationOccurs(0.0), false; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	inputScores := []BrainScore{{score: 600}, {score: 300}, {score: 100}}
	expectedChances := []float32{0.6, 0.3, 0.1}
	if got := geneChance(inputScores); !reflect.DeepEqual(got, expectedChances) {
		t.Errorf("Got %v, want %v", got, expectedChances)
	}

	if got, want := percentageOfWithMin1(100, 0.32), 32; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := percentageOfWithMin1(5, 0.05), 1; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	if got, want := removeIndexFromIDSlice([]IDType{0, 1, 2, 3}, 1), []IDType{0, 3, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := removeIndexFromIDSlice([]IDType{0}, 0), []IDType{}; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
	if got, want := removeIndexFromSynSlice([]Synapse{{src: 1, dst: 2}, {src: 2, dst: 3}}, 1), []Synapse{{src: 1, dst: 2}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}
