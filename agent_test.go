package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestLoadAndMarshallConfig(t *testing.T) {
	var agent Agent
	var buffer bytes.Buffer
	agent = LoadAgent("./agentconfig.json")
	body, _ := json.Marshal(agent.Config)
	var jsonDoc = `{
					"account": ["hello", "world"],
					"augmentations" : {
							"frequency-cap-ex" : {
									"config" : 42,
									"filters" : {
											"include" : [ "pass-frequency-cap-ex" ]
									},
									"required" : true
							},
							"random" : null
					},
					"bidControl" : {
							"fixedBidCpmInMicros" : 0,
							"type" : "RELAY"
					},
					"bidProbability" : 0.1,
					"creatives" : [
							{
									"format" : "728x90",
									"id" : 2,
									"name" : "LeaderBoard"
							},
							{
									"format" : "160x600",
									"id" : 0,
									"name" : "LeaderBoard"
							},
							{
									"format" : "300x250",
									"id" : 1,
									"name" : "BigBox"
							}
					],
					"errorFormat" : "lightweight",
					"external" : false,
					"externalId" : 0,
					"lossFormat" : "lightweight",
					"minTimeAvailableMs" : 5,
					"winFormat" : "full"
			}`
	json.Compact(&buffer, []byte(jsonDoc))
	jsonResult := (&buffer).Bytes()
	if !bytes.Equal(body, jsonResult) {
		t.Errorf("Expected to be equal, but it was: %s == %s", body, jsonResult)
	}
}

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
	fmt.Println(agents[0].Name)
	// Output: my_http_config
}

func ExampleLoadAgentsFromFile() {
	agents := LoadAgentsFromFile("./agents.json")
	fmt.Println(agents[0].Name)
	// Output: my_http_config
}
