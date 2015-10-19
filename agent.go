package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
)

type Creative struct {
	Format string `json:"format"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
}

type AgentConfig struct {
	// This is the agent configuration that will be sent to RTBKIT's ACS

	// We use `RawMessage` for Augmentations and BidcControl, because we
	// don't need it.
	Account            []string        `json:"account"`
	Augmentations      json.RawMessage `json:"augmentations"`
	BidControl         json.RawMessage `json:"bidControl"`
	BidProbability     float64         `json:"bidProbability"`
	Creatives          []Creative      `json:"creatives"`
	ErrorFormat        string          `json:"errorFormat"`
	External           bool            `json:"external"`
	ExternalId         int             `json:"externalId"`
	LossFormat         string          `json:"lossFormat"`
	MinTimeAvailableMs float64         `json:"minTimeAvailableMs"`
	WinFormat          string          `json:"full"`
}

type Agent struct {
	Name   string      `json:"name"`
	Config AgentConfig `json:"config"`

	// This is the price the agent will pay per impression. "Fixed price bidder".
	Price float64 `json:"price"`

	// For pacing the budgeting
	Period  int `json:"period"`
	Balance int `json:"balance"`

	// private state of each agent
	registered bool      // did we register the configuration in the ACS?
	pacer      chan bool // go routine updating balance in the banker
	bidId      int       // unique id for response
}

func LoadAgent(filepath string) Agent {
	var agent Agent
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &agent)
	if err != nil {
		log.Fatal(err)
	}
	return agent
}

func LoadAgents(r io.Reader) map[string]Agent {
	agents := make(map[string]Agent)
	dec := json.NewDecoder(r)
	for {
		var a Agent
		if err := dec.Decode(&a); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		agents[a.Name] = a
	}
	return agents
}
