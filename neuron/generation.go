package neuron

type BrainResult struct {
	id      IDType
	inputs  []SignalType
	outputs []SignalType
	steps   int
}

type GenerationConfig struct {
	MaxSteps int
}

type Generation struct {
	config GenerationConfig
	brains map[IDType]*Brain
}

func NewGeneration(gconf GenerationConfig, codes map[IDType]*DNA) *Generation {
	g := &Generation{
		config: gconf,
		brains: make(map[IDType]*Brain, len(codes)),
	}

	for id, dna := range codes {
		g.brains[id] = Flourish(dna)
	}
	return g
}

func (g *Generation) FireBrains(inputs map[IDType][]SignalType) map[IDType]*BrainResult {
	// Simulate all brains in separate goroutines.
	resChan := make(chan BrainResult)
	for id := range g.brains {
		go g.fireBrain(id, inputs[id], resChan)
	}

	// Wait for all the results to come in before returning.
	results := make(map[IDType]*BrainResult, len(g.brains))
	for i := 0; i < len(g.brains); i++ {
		result := <-resChan
		results[result.id] = &result
	}

	return results
}

func (g *Generation) fireBrain(id IDType, inputs []SignalType, resChan chan BrainResult) {
	brain := g.brains[id]
	brain.SeeInput(inputs)

	result := BrainResult{
		id:     id,
		inputs: inputs,
	}

	for step := 0; step < g.config.MaxSteps; step++ {
		outputs := brain.StepFunction()
		if len(outputs) > 0 {
			result.outputs = outputs
			result.steps = step
			resChan <- result
			return
		}
	}

	result.outputs = make([]SignalType, 0)
	result.steps = g.config.MaxSteps
	resChan <- result
}
