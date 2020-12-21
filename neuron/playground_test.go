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
			ChangeOp:  0.5,
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

func CreateTestPlayground() *Playground {
	p := NewPlayground(PlaygroundConfig{
		NumInputs:  2,
		NumOutputs: 1,

		NumVariants: 10,
		Generations: 3,

		Mconf: MutationConfig{
			// Guaranteed at least 1 addition regardless of these values.
			NeuronExpansion:  0.01,
			SynapseExpansion: 0.01,

			AddNeuron:  1.0,
			AddSynapse: 1.0,

			ChangeOp:  0.5,
			SetSeed:   0.5,
			UnsetSeed: 0.5,
		},

		Econf: EvolutionConfig{
			Parents:           2,
			BottomTierPercent: 0.25,
			DistanceThreshold: 0.2,
		},

		Gconf: GenerationConfig{
			Rounds:   2,
			MaxSteps: 20,

			InputsFn: func(actions int) []SignalType {
				if actions >= 2 {
					return []SignalType{}
				}
				return []SignalType{SignalType(actions + 1), SignalType(actions + 2)}
			},

			FitnessFn: func(outputs []SignalType) ScoreType {
				return ScoreType(outputs[0])
			},
		},
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

	for i := 0; i < p.config.NumVariants; i++ {
		dna := p.codes[i]
		p.mutateDNAStructure(dna)
		p.mutateDNAStructure(dna)
		p.mutateNeurons(dna)
	}

	return p
}

func TestSimulatePlayground(t *testing.T) {
	p := CreateTestPlayground()
	arbitaryDNA := p.codes[0]

	p.SimulatePlayground()

	if arbitaryDNA == p.codes[0] {
		t.Errorf("Expected evolution, got nothing")
	}
}

func TestSpeciation(t *testing.T) {
	p := CreateTestPlayground()

	scores := make([]BrainScore, p.config.NumVariants)
	for i := 0; i < p.config.NumVariants; i++ {
		scores[i] = BrainScore{id: i, score: ScoreType(p.rnd.Intn(100))}
	}

	p.species[0] = &Species{
		rep: p.codes[0],
	}

	speciesOffspring := p.speciation(scores)
	fmt.Printf("speciesOffspring: %v\n", speciesOffspring)

	if p.species[0].scores[0].id != 0 {
		t.Errorf("0 didn't get grouped")
	}

	sumOffspring := 0
	for _, offspring := range speciesOffspring {
		sumOffspring += offspring
	}
	if got, want := sumOffspring, p.config.NumVariants; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestDNADistance(t *testing.T) {
	p := CreateTestPlayground()

	a := NewDNA(p.source)
	a.AddNeuron(0, OR)
	a.AddNeuron(1, OR)
	a.AddNeuron(2, OR)
	a.AddNeuron(3, OR)
	a.AddSynapse(0)
	a.AddSynapse(1)
	a.AddSynapse(2)
	a.AddSynapse(3)

	b := NewDNA(p.source)
	b.AddNeuron(0, OR)
	b.AddNeuron(1, OR)
	b.AddNeuron(2, OR)
	b.AddNeuron(4, OR)
	b.AddSynapse(0)
	b.AddSynapse(8)
	b.AddSynapse(4)
	b.AddSynapse(5)

	if got, want := dnaDistance(a, b), float32(6)/float32(8); got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestReproduction(t *testing.T) {
	p := CreateTestPlayground()
	species := &Species{
		scores: []BrainScore{
			{id: 0, score: 200},
			{id: 1, score: 400},
			{id: 2, score: 300},
			{id: 3, score: 100},
		},
	}

	newCodes := p.reproduction(species, 2)
	if got, want := len(newCodes), 2; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
	// The variant with score 100 dies off.
	if got, want := species.Size(), 3; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestCreateOffspring(t *testing.T) {
	p := CreateTestPlayground()

	fmt.Printf("dna 0: %s\n", p.codes[0].PrettyPrint())
	fmt.Printf("dna 1: %s\n", p.codes[1].PrettyPrint())

	parentScores := []BrainScore{{id: 0, score: 60}, {id: 1, score: 40}}
	child := p.createOffspring(parentScores)
	fmt.Printf("child: %s\n", child.PrettyPrint())

	if _, ok := child.Neurons[2]; !ok {
		t.Errorf("Child didn't get a motor neuron")
	}
}

func TestShiftConglomerate(t *testing.T) {
	p := CreateTestPlayground()

	numInterBefore := p.source.NeuronIDs[INTER].Length()
	nextIDBefore := p.source.Synapses.nextID
	p.shiftConglomerate()

	if got, want := p.source.NeuronIDs[INTER].Length(), numInterBefore+1; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}

	// The nextID will be +2 because of the new inter neuron, then +1 because of
	// the new synapse being added.
	if got, want := p.source.Synapses.nextID, nextIDBefore+1+2; got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
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
	p := CreateTestPlayground()

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
	p := CreateTestPlayground()

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

	outputIndices := make([]int, len(inputScores))
	for i := 0; i < 200; i++ {
		outputIndices[p.randomParentGene(inputScores)]++
	}
	if outputIndices[0] < outputIndices[1] || outputIndices[1] < outputIndices[2] {
		t.Errorf("Got %v, expected ratio more like %v", outputIndices, expectedChances)
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
