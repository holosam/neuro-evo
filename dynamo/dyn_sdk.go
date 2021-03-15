package dynamo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	inputFileName  = "config.json"
	outputFileName = "results.json"
)

// Use a hacky global variable to store the specific filepath while testing
// multiple instances on the same server.
var DynDir = ""

type intConfig struct {
	Name      string    `json:"name"`
	Value     int       `json:"value"`
	Timestamp time.Time `json:"mtime"`
}

type configsWrapper struct {
	Configs []intConfig `json:"configs"`
}

func Get(configName string, defaultVal int) int {
	if DynDir == "" {
		return defaultVal
	}

	inputFilePath := fmt.Sprintf("%s%s", DynDir, inputFileName)
	file, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", inputFilePath, err)
		return defaultVal
	}

	configs := configsWrapper{}
	err = json.Unmarshal([]byte(file), &configs)
	if err != nil {
		fmt.Printf("Error unmarshaling %s: %v\n", inputFilePath, err)
		return defaultVal
	}

	for _, config := range configs.Configs {
		if config.Name == configName {
			return config.Value
		}
	}

	fmt.Printf("Couldn't find config for: %s\n", configName)
	return defaultVal
}

func Record(session string, value int) {
	if DynDir == "" {
		fmt.Printf("Not writing result for %s because dir is empty\n", session)
		return
	}

	result := intConfig{
		Name:      session,
		Value:     value,
		Timestamp: time.Now(),
	}

	output, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Error marshaling result %+v: %v\n", result, err)
		return
	}

	outputFilePath := fmt.Sprintf("%s%s", DynDir, outputFileName)
	err = ioutil.WriteFile(outputFilePath, output, 0644)
	if err != nil {
		fmt.Printf("Error writing %s: %v\n", outputFilePath, err)
		return
	}
}
