package main

import (
	"fmt"
	"log"
	"os"
)

func ExampleLoadAgent() {
	var agent Agent
	agent = LoadAgent("./agentconfig.json")
	fmt.Println(agent.Name)
	// Output: my_http_config
}

func ExampleLoadAgents() {
	f, err := os.Open("./agents.json")
	if err != nil {
		log.Fatal(err)
	}
	agents := LoadAgents(f)
	for key, value := range agents {
		fmt.Println(key, value.Config.Account)
	}
	// Output: my_http_config ["hello", "world"]
}
