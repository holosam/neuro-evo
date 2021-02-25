package main

import (
	"hackathon/sam/evolve/env"
)

/*
To deploy:
1) Build binary
	$ cd ~/src/hackathon
	$ env GOOS=linux GOARCH=amd64 go build  (builds it for ubuntu)
2) Copy binary to droplet
	- sign in to digitalocean.com
	- copy external IP address of the droplet
	$ scp evolve root@0.0.0.0:~  (replace with IP)
3) Run binary in background
	$ ssh root@0.0.0.0
	$ nohup ./evolve > out.txt &
	$ ps aux | grep evolve  (shows the process ID if it needs to be killed)
*/

// Current tasks:
// 1) The scores are directly affecting totalFitness so they must be positive
//    numbers. Should be more resiliant like base everything off the minumum
//    fitness and make that = 1. (Or even better, use a true loss function
//    where 0 is best. UNLESS going with #5: optimization implies high score)
// 1.5) The current implementation of geneChance() implies that the scores are
//    all positive, and that they are linearly scaled (1000 is 2x better than
//    500).
// 2) Swarm intelligence idea: After rounds of evolution, would be cool to
//    check what the consensus answer is among all variants (each vote on a
//    response and pick the winner). They may be smarter together than any
//    individual.
//  - a voting system like this may require all outputs to be included in
//    the BrainScores.
// 3) It would be really nice to be able to print and read the game status
//    during evolution. Could add Print() to the interface.
// 4) Use Sparse Categorical Cross Entropy as the loss function to select for
//    variants that give the closest output for each input. It would be ideal
//    to have a general fitness function instead of needing a new one for each
//    game type.
// 5) Find business applications for this project. Some potential options:
//  - Thundier: codify logic based on expected inputs/outputs entered by the
//    user. Expected loss should be zero and needs to be called by other APIs
//  - Focus on optimization problems. Remember lyft surge pricing? There must
//    be many other applications for this that low-code founders need help with
//    and can be trained. Doesn't require a perfect loss, just needs to be
//    better than what the user can do themselves.
// 6) Possible memory leak. Failed during shiftConglomerate() of gen 761 while
//    trying to add a new synapse. It's adding neurons/syns too quickly still,
//    since the winner of that gen only had up to neuron #88 but the conglom
//    had 2632.
// 7) The new streaming inputs/outputs are an improvement, but I think they're
//    still insufficient to handle larger problems. It can't be streaming in
//    ascii values for numbers >=256. And ideally string fields would go in
//    together too. At some point I thought of using a bloom filter.

func main() {
	env.RunHealthChecker()
}
