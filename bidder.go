package main

import (
	"io"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"strings"
	"strconv"
	"time"
	"gopkg.in/bsm/openrtb.v1"
	"math/rand"
	"os"
	"os/signal"
)

const (
	ACSIp string = "127.0.0.1"
	ACSPort = 9986
	BankerIp = "127.0.0.1"
	BankerPort = 9985
	BidderPort = 7654
	BidderWin = 7653
	BidderEvent = 7652
	AgentConfigFile = "agentconfig.json"
)


type Agent struct {
	Name string
	ExternalId int

	Config map[string]interface{}

	// Fixed price
	Price float64

	// For pacing the banking
	Period, Budget int

	registed bool

	pacer chan bool

	BidId int
}

func (agent *Agent) RegisterAgent(httpClient *http.Client) {
	url := fmt.Sprintf("http://%s:%d/v1/agents/%s/config", ACSIp, ACSPort, agent.Name)
	body, _ := json.Marshal(agent.Config)
	reader := bytes.NewReader(body)
	req, _ := http.NewRequest("POST", url, reader)
	req.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("ACS registration failed with %s", err)
		return
	}
	agent.registed = true
	res.Body.Close()
}

func (agent *Agent) StartPacer(httpClient *http.Client) {

	// Convert a slice of interface{} to a slice of string.
	accounts := make([]string, 2)
	for i, account := range agent.Config["account"].([]interface{}) {
		accounts[i] = account.(string)
	}

	url := fmt.Sprintf("http://%s:%d/v1/accounts/%s/balance",
		BankerIp, BankerPort, strings.Join(accounts, ":"))
	body := fmt.Sprintf("{\"USD/1M\": %d}", agent.Budget)
	ticker := time.NewTicker(15000 * time.Millisecond)
	agent.pacer = make(chan bool)

	go func() {
		for {
			select {
			case <- ticker.C:
				// make this a new go routine?
				go func() {
					fmt.Println("Pacing...")
					req, _ := http.NewRequest("POST", url, strings.NewReader(body))
					req.Header.Add("Accept", "application/json")
					res, err := httpClient.Do(req)
					if err != nil {
						fmt.Println("Balance failed with %s", err)
						return
					}
					res.Body.Close()
				}()
			case <- agent.pacer:
				ticker.Stop()
				return
			}
		}
	}()
}


func (agent *Agent) StopPacer() {
	close(agent.pacer)
}


func (agent *Agent) DoBid(
	req *openrtb.Request, res *openrtb.Response, ids *map[Key]interface{}) (*openrtb.Response, bool) {


	for _, imp := range req.Imp {
		key := Key{ImpId: *imp.Id, ExtId: agent.ExternalId}
		if (*ids)[key] == nil {
			continue
		}
		creativeList := (*ids)[key].([]interface{})
		n := rand.Intn(len(creativeList))
		// json reads numbers as float64, which I guess is a good default
		// but they are really integers, because it's an index to
		// the creatives list of the agent config

		cridx := int(creativeList[n].(float64))
		creative := (agent.Config["creatives"].([]interface{}))[cridx]
		crid := strconv.Itoa(int(creative.(map[string]interface{})["id"].(float64)))
		bidId := strconv.Itoa(agent.BidId)
		price := float32(agent.Price)
		ext := map[string]interface{}{"priority": 1.0, "external-id": agent.ExternalId}
		bid := openrtb.Bid{Id: &bidId, Impid: imp.Id, Crid: &crid, Price: &price, Ext: ext}
		agent.BidId += 1
		res.Seatbid[0].Bid = append(res.Seatbid[0].Bid, bid)
	}
	return res, len(res.Seatbid[0].Bid) > 0
}

type Key struct {
	ImpId string
	ExtId int
}

func ExternalIdsFromRequest(req *openrtb.Request) *map[Key]interface{} {
	ids := make(map[Key]interface{})

	for _, imp := range req.Imp {
		for _, extId := range imp.Ext["external-ids"].([]interface{}) {
			// types, types and more types... *sigh*
			key := Key{ImpId: *imp.Id, ExtId: int(extId.(float64))}  // json turns it into a float even though it's an int.
			creatives := (imp.Ext["creative-indexes"].(map[string]interface{}))[strconv.Itoa(int(extId.(float64)))]
			ids[key] = creatives.(interface{})
		}
	}
	return &ids
}


func EmptyOneSeatResponse(req *openrtb.Request) *openrtb.Response {

	seat := openrtb.Seatbid{Bid: make([]openrtb.Bid, 0)}
	seatbid := []openrtb.Seatbid{seat}
	res := &openrtb.Response{Id: req.Id, Seatbid: seatbid}
	return res

}


type AgentConfig map[string]interface{};

func loadAgentConfig() AgentConfig {
	var conf AgentConfig
	data, err := ioutil.ReadFile(AgentConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	return conf
}


func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s request took %s", name, elapsed)
}

func track(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer timeTrack(time.Now(), name)
		fn(w, req)
	}
}


func main() {
	// http client to pace agents (note that it's pointer)
	client := &http.Client{}

	// load configuration
	conf := loadAgentConfig()

	agent := Agent{Name: "my_http_config", Config: conf, ExternalId: 0, Price: 1.0, Period: 30000, BidId: 1}
	agent.RegisterAgent(client)
	agent.StartPacer(client)

	mux := http.NewServeMux()

	mux.HandleFunc("/", track(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var ids *map[Key]interface{}
		enc := json.NewEncoder(w)

		req, err := openrtb.ParseRequest(r.Body)

		if err != nil {
			log.Println("ERROR", err.Error())
			w.WriteHeader(204) // respond with 'no bid'
			return
		}

		log.Println("INFO Received bid request", *req.Id)

		ids = ExternalIdsFromRequest(req)
		res := EmptyOneSeatResponse(req)

		if res, ok := agent.DoBid(req, res, ids); ok {
			w.Header().Set("Content-type", "application/json")
			w.Header().Add("x-openrtb-version", "2.1")
			w.WriteHeader(http.StatusOK)
			enc.Encode(res)
			return
		}

		w.WriteHeader(204)

	}, "bidding"))

	go http.ListenAndServe(fmt.Sprintf(":%d", BidderPort), mux)


	evemux := http.NewServeMux()
	evemux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "")
		log.Println("Event!")
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), evemux)

	winmux := http.NewServeMux()
	winmux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "")
		log.Println("Win!")
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", BidderWin), winmux)


	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <- c:
		// Implement remove agent from ACS
		fmt.Println("Leaving...")
	}
}
