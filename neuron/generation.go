package neuron

type Generation struct {
	brains map[int]*Brain
}

func NewGeneration() *Generation {
	g := Generation{
		brains: make(map[int]*Brain),
	}
	return &g
}

func (g *Generation) AddSpecies(id int, dna *DNA) {
	g.brains[id] = Flourish(dna)
}

func (g *Generation) FireBrains() {
	for _, brain := range g.brains {
		brain.StepFunction()
	}
}
