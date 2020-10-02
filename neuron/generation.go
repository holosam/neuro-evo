package neuron

// AccuracyFunc scores the output moves. Closer to 0 is better.
type AccuracyFunc func(moves []SignalType) int64

type BrainResult struct {
	id    int
	moves []SignalType
	steps int
}

func RunGeneration(codes map[int]*DNA) map[int]*BrainResult {
	// Simulate all brains in separate goroutines.
	resChan := make(chan BrainResult)
	for id, code := range codes {
		go FireBrain(id, code, resChan)
	}

	// Wait for all the scores to come in.
	results := make(map[int]*BrainResult, len(codes))
	for i := 0; i < len(codes); i++ {
		result := <-resChan
		results[result.id] = &result
	}

	return results
}

func FireBrain(id int, dna *DNA, scoreChan chan BrainResult) {
	brain := Flourish(dna)
	for step := 0; step < maxStepsPerGen; step++ {
		moves := brain.StepFunction()
		if len(moves) > 0 {
			// score := 1000000 * accuracy(moves)
			// score += 1000 * int64(step)
			// score += int64(dnaComplexity(dna))
			scoreChan <- BrainResult{
				id:    id,
				moves: moves,
				steps: step,
			}
			return
		}
	}

	scoreChan <- BrainResult{
		id:    id,
		moves: make([]SignalType, 0),
		steps: maxStepsPerGen,
	}
}
