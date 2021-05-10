### Overview
Experimenting with evolving the neural network structures to solve generic problems.

Filename | Description
-------- | -----------
neuron.go | Performs mathematical operations on a series of input values.
brain.go | DNA class which encodes a series of connected neurons.
playground.go | Handles the speciation, reproduction, and mutation of Brains in one generation.
runner.go | Runs the playground over many generations.
env.go | Sets up the environment for the runner.

The speciation concept comes from [NEAT method](http://nn.cs.utexas.edu/downloads/papers/stanley.ec02.pdf) for neural net evolution, however the underlying neuron operations and evolution method are original. This evolution method involves keeping track of a "conglomerate" network, which is a superset of all the neurons from every genome. During the reproduction step, the networks can be overlaid on this conglomerate to match up the underlying genes.

### Deploy on Digital Ocean
1) Build binary for ubuntu
	```bash
	$ cd ~/neuro-evo
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
