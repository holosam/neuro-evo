package main

import (
	"fmt"
	"strconv"
)

// import "hackathon/sam/evolve/neuron"

func main() {
	var andValue int64
	andValue = 5 & 4 & 15
	fmt.Println(strconv.FormatInt(andValue, 2))
}
