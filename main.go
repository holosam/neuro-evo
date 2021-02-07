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
	$ nohup ./evolve > out.txt &  ()
*/

// Current tasks:
// 1) Figure out a way to do NeuronExpansion and SynapseExpansion
//    without insane exponential growth (logarathmic growth?)
//    For example: generation 50 / 25 = 2 new neurons
// 3) Observation: species are churning a lot, but it appears that an early
//    variant in generation #13 had the best score. Doesn't always need to
//    grow by a lot (or at all)
//  - Generally need to find a way to decide when there's been "enough"
//    evolution and the simulation can end with a winner.
// 4) The scores are directly affecting totalFitness so they must be positive
//    numbers. Should be more resiliant like base everything off the minumum
//    fitness and make that = 1

func main() {
	env.RomanNumeralConverter()
}
