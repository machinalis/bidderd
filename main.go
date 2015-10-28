package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gopkg.in/bsm/openrtb.v1"
)

const (
	ACSIp       string = "127.0.0.1"
	ACSPort            = 9986
	BankerIp           = "127.0.0.1"
	BankerPort         = 9985
	BidderPort         = 7654
	BidderWin          = 7653
	BidderEvent        = 7652
)

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
	var agentsConfigFile = flag.String("config", "", "Configuration file in JSON.")
	flag.Parse()
	if *agentsConfigFile == "" {
		log.Fatal("You should provide a configuration file.")
	}

	// http client to pace agents (note that it's pointer)
	client := &http.Client{}

	// load configuration
	agents, err := LoadAgentsFromFile(*agentsConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, agent := range agents {
		agent.RegisterAgent(client, ACSIp, ACSPort)
		agent.StartPacer(client, BankerIp, BankerPort)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", track(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var (
			ok    bool = true
			tmpOk bool = true
		)
		enc := json.NewEncoder(w)
		req, err := openrtb.ParseRequest(r.Body)

		if err != nil {
			log.Println("ERROR", err.Error())
			w.WriteHeader(204) // respond with 'no bid'
			return
		}

		log.Println("INFO Received bid request", *req.Id)

		ids := externalIdsFromRequest(req)
		res := emptyResponseWithOneSeat(req)

		for _, agent := range agents {
			res, tmpOk = agent.DoBid(req, res, ids)
			ok = tmpOk || ok
		}

		if ok {
			w.Header().Set("Content-type", "application/json")
			w.Header().Add("x-openrtb-version", "2.1")
			w.WriteHeader(http.StatusOK)
			enc.Encode(res)
			return
		}
		log.Println("No bid.")
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
	case <-c:
		// Implement remove agent from ACS
		for _, agent := range agents {
			agent.UnregisterAgent(client, ACSIp, ACSPort)
		}
		fmt.Println("Leaving...")
	}
}
