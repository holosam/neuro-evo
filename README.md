Experimenting with evolving the neural network structures to solve generic problems.

### Deploy on Digital Ocean
1) Build binary for ubuntu
	```bash
	$ cd ~/src/hackathon
	$ env GOOS=linux GOARCH=amd64 go build
	```
2) Copy binary to droplet
	* sign in to digitalocean.com
	* copy external IP address of the droplet
	```bash
	$ scp evolve root@0.0.0.0:~ 
	```
3) Run binary in background
	```bash
	$ ssh root@0.0.0.0
	$ nohup ./evolve > out.txt &
	```
	* show the process ID if it needs to be killed: `$ ps aux | grep evolve`

### Project improvement ideas
* [Sparse Categorical Cross Entropy Loss](https://machinelearningmastery.com/how-to-choose-loss-functions-when-training-deep-learning-neural-networks/) - loss function for scoring neural networks with multi-class outputs.
* Swarm Intelligence - all the neural nets in a generation vote on an answer.
* Protobuf Input - take a proto as input and use reflection to extract the fields.
