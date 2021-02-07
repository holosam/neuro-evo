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
// 1) The scores are directly affecting totalFitness so they must be positive
//    numbers. Should be more resiliant like base everything off the minumum
//    fitness and make that = 1
// 2) Swarm intelligence idea: After rounds of evolution, would be cool to
//    check what the consensus answer is among all variants (each vote on a
//    response and pick the winner). They may be smarter together that any
//    individual.

func main() {
	env.RunAdder()
}
