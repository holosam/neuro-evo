package neuron

// AccuracyFunc scores the output outputs. Closer to 0 is better.
type AccuracyFunc func(inputs []SignalType, outputs []SignalType) int

type BrainResult struct {
	id      IDType
	outputs []SignalType
	steps   int
}

type GenerationConfig struct {
	MaxSteps int
}

type Generation struct {
	config GenerationConfig

	brains  map[IDType]*Brain
	resChan chan BrainResult
}

func NewGeneration(gconf GenerationConfig, codes map[IDType]*DNA, resChan chan BrainResult) *Generation {
	g := &Generation{
		config:  gconf,
		brains:  make(map[IDType]*Brain, len(codes)),
		resChan: resChan,
	}

	for id, dna := range codes {
		g.brains[id] = Flourish(dna)
	}
	return g
}

func (g *Generation) FireBrains(inputs []SignalType) map[IDType]*BrainResult {
	// Simulate all brains in separate goroutines.
	resChan := make(chan BrainResult)
	for id := range g.brains {
		go g.fireBrain(id, inputs)
	}

	// Wait for all the results to come in before returning.
	results := make(map[IDType]*BrainResult, len(g.brains))
	for i := 0; i < len(g.brains); i++ {
		result := <-resChan
		results[result.id] = &result
	}

	return results
}

func (g *Generation) fireBrain(id IDType, inputs []SignalType) {
	brain := g.brains[id]
	brain.SeeInput(inputs)

	for step := 0; step < g.config.MaxSteps; step++ {
		outputs := brain.StepFunction()
		if len(outputs) > 0 {
			g.resChan <- BrainResult{
				id:      id,
				outputs: outputs,
				steps:   step,
			}
			return
		}
	}

	g.resChan <- BrainResult{
		id:      id,
		outputs: make([]SignalType, 0),
		steps:   g.config.MaxSteps,
	}
}

func RunGeneration(codes map[IDType]*DNA, envInputs []SignalType, maxSteps int) map[int]*BrainResult {
	// Simulate all brains in separate goroutines.
	resChan := make(chan BrainResult)
	for id, code := range codes {
		go FireBrain(id, code, envInputs, maxSteps, resChan)
	}

	// Wait for all the scores to come in.
	results := make(map[int]*BrainResult, len(codes))
	for i := 0; i < len(codes); i++ {
		result := <-resChan
		results[result.id] = &result
	}

	return results
}

func FireBrain(id IDType, dna *DNA, envInputs []SignalType, maxSteps int, resChan chan BrainResult) {
	brain := Flourish(dna)

	// At some point it would be good to add support for
	// multiple rounds of seeing. Maybe [][]SignalType
	brain.SeeInput(envInputs)
	for step := 0; step < maxSteps; step++ {
		outputs := brain.StepFunction()
		if len(outputs) > 0 {
			resChan <- BrainResult{
				id:      id,
				outputs: outputs,
				steps:   step,
			}
			return
		}
	}

	resChan <- BrainResult{
		id:      id,
		outputs: make([]SignalType, 0),
		steps:   maxSteps,
	}
}

func FireBrainBlock(dna *DNA, envInputs []SignalType, maxSteps int) []SignalType {
	brain := Flourish(dna)
	brain.SeeInput(envInputs)
	for step := 0; step < maxSteps; step++ {
		outputs := brain.StepFunction()
		if len(outputs) > 0 {
			return outputs
		}
	}
	return make([]SignalType, 0)
}
