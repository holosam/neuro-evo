package neuron

// AccuracyFunc scores the output moves. Closer to 0 is better.
type AccuracyFunc func(inputs []SignalType, outputs []SignalType) int

type BrainResult struct {
	id    int
	moves []SignalType
	steps int
}

func RunGeneration(codes map[int]*DNA, envInputs []SignalType, maxSteps int) map[int]*BrainResult {
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

func FireBrain(id int, dna *DNA, envInputs []SignalType, maxSteps int, resChan chan BrainResult) {
	brain := Flourish(dna)

	// At some point it would be good to add support for
	// multiple rounds of seeing. Maybe [][]SignalType
	brain.SeeInput(envInputs...)
	for step := 0; step < maxSteps; step++ {
		moves := brain.StepFunction()
		if len(moves) > 0 {
			resChan <- BrainResult{
				id:    id,
				moves: moves,
				steps: step,
			}
			return
		}
	}

	resChan <- BrainResult{
		id:    id,
		moves: make([]SignalType, 0),
		steps: maxSteps,
	}
}

func FireBrainBlock(dna *DNA, envInputs []SignalType, maxSteps int) []SignalType {
	brain := Flourish(dna)
	brain.SeeInput(envInputs...)
	for step := 0; step < maxSteps; step++ {
		moves := brain.StepFunction()
		if len(moves) > 0 {
			return moves
		}
	}
	return make([]SignalType, 0)
}
