To deploy:
1) Build binary for ubuntu
	$ cd ~/src/hackathon
	$ env GOOS=linux GOARCH=amd64 go build
2) Copy binary to droplet
	- sign in to digitalocean.com
	- copy external IP address of the droplet
	$ scp evolve root@0.0.0.0:~  (replace with IP)
3) Run binary in background
	$ ssh root@0.0.0.0
	$ nohup ./evolve > out.txt &
	$ ps aux | grep evolve  (shows the process ID if it needs to be killed)

Main improvement ideas:
- Sparse Categorical Cross Entropy Loss
- Swarm Intelligence
- Proto input