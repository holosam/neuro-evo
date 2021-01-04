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
// 4) Find out why generations go from taking 1 minute to 30 minutes
// Optional) Give droplet more CPU before running again

func main() {
	env.StockSimulation()
}
