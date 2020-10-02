package neuron

import "fmt"

// AccuracyFunc scores the output moves. Closer to 0 is better.
type AccuracyFunc func(moves []SignalType) int64

type BrainResult struct {
	id    int
	moves []SignalType
	steps int
}

func RunGeneration(codes map[int]*DNA, envInputs []SignalType) map[int]*BrainResult {
	// Simulate all brains in separate goroutines.
	resChan := make(chan BrainResult)
	for id, code := range codes {
		go FireBrain(id, code, envInputs, resChan)
	}

	// Wait for all the scores to come in.
	results := make(map[int]*BrainResult, len(codes))
	for i := 0; i < len(codes); i++ {
		result := <-resChan
		results[result.id] = &result
	}

	return results
}

func FireBrain(id int, dna *DNA, envInputs []SignalType, resChan chan BrainResult) {
	brain := Flourish(dna)

	// At some point it would be good to add support for
	// multiple rounds of seeing. Maybe [][]SignalType
	brain.SeeInput(envInputs...)
	for step := 0; step < maxStepsPerGen; step++ {
		moves := brain.StepFunction()
		if len(moves) > 0 {
			fmt.Printf("Stopping id=%d with moves=%v\n", id, moves)
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
		steps: maxStepsPerGen,
	}
}
