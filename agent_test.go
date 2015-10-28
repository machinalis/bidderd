package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"gopkg.in/bsm/openrtb.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// this is from openrtb, useful :-)
func fixture(fname string, v interface{}) error {
	data, err := ioutil.ReadFile(filepath.Join("testdata", fname+".json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func TestLoadAndMarshallConfig(t *testing.T) {
	var agent Agent
	var buffer bytes.Buffer
	agents, _ := LoadAgentsFromFile("./agents.json")
	agent = agents[0]
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

func ExampleLoadAgentsFromFile() {
	agents, err := LoadAgentsFromFile("./agents.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(agents[0].Name)
	// Output: my_http_config
}

var _ = Describe("Agent", func() {
	var res *openrtb.Response
	var req *openrtb.Request
	var a Agent

	BeforeEach(func() {
		config := AgentConfig{Creatives: []Creative{Creative{Id: 1}}}
		a = Agent{Name: "test_agent", Config: config, Price: 1.0, Period: 30000, Balance: 15000}
		err := fixture("openrtb1_req", &req)
		res = emptyResponseWithOneSeat(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(req.Imp[0].Id).NotTo(BeNil())
	})

	It("bid should have a price", func() {
		ids := externalIdsFromRequest(req)
		a.DoBid(req, res, ids)
		Expect(*res.Seatbid[0].Bid[0].Price).To(Equal(float32(1.0)))
	})
})
