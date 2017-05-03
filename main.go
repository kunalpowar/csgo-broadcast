package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type fullDataDetails struct {
	at   time.Time
	tick int
	data []byte
}

var (
	fullData  = make(map[string]fullDataDetails)
	deltaData = make(map[string][]byte)
	startData = make(map[string][]byte)

	startTime, latestFullData time.Time
	syncDetails               SyncDetails

	receivedStartReq bool
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/match/{token}/sync", sync)

	r.HandleFunc("/match/{token}/{fragment_number}/start", clientStart)
	r.HandleFunc("/match/{token}/{fragment_number}/full", clientFull)
	r.HandleFunc("/match/{token}/{fragment_number}/delta", clientDelta)

	r.HandleFunc("/{token}/{fragment_number}/start", start).Methods("POST")
	r.HandleFunc("/{token}/{fragment_number}/full", full).Methods("POST")
	r.HandleFunc("/{token}/{fragment_number}/delta", delta).Methods("POST")

	addr := "0.0.0.0:3090"
	log.Printf("starting server at %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

type SyncDetails struct {
	Tick           int   `json:"tick"`
	Rtdelay        int64 `json:"rtdelay"`
	Rcvage         int64 `json:"rcvage"`
	Fragment       int   `json:"fragment"`
	SignupFragment int   `json:"signup_fragment"`
	Tps            int   `json:"tps"`
	Protocol       int   `json:"protocol"`
}

func start(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	f := vars["fragment_number"]
	log.Printf("received start request with token %s and fragment number %s and query %v", vars["token"], vars["fragment_number"], q)

	startTime = time.Now()

	tps, err := strconv.ParseFloat(q.Get("tps"), 64)
	if err != nil {
		panic(err)
	}

	protocol, err := strconv.Atoi(q.Get("protocol"))
	if err != nil {
		panic(err)
	}

	fg, err := strconv.Atoi(f)
	if err != nil {
		panic(err)
	}

	syncDetails.SignupFragment = fg
	syncDetails.Tps = int(tps)
	syncDetails.Protocol = protocol

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("could not read from req: %v", err)
	}
	defer r.Body.Close()

	startData[f] = bs
	receivedStartReq = true

	w.WriteHeader(http.StatusOK)
}

func full(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	f := vars["fragment_number"]
	log.Printf("received full request with token %s and fragment number %s and query %v", vars["token"], f, r.URL.Query())

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("could not read from req: %v", err)
	}
	defer r.Body.Close()

	tick, err := strconv.Atoi(q.Get("tick"))
	if err != nil {
		panic(err)
	}

	now := time.Now()
	fullData[f] = fullDataDetails{at: now, tick: tick, data: bs}

	latestFullData = now

	if !receivedStartReq {
		// This will ask the server to make a start request.
		w.WriteHeader(http.StatusResetContent)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func delta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f := vars["fragment_number"]
	log.Printf("received delta request with token %s and fragment number %s and query %v", vars["token"], f, r.URL.Query())

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("could not read from req: %v", err)
	}
	defer r.Body.Close()

	deltaData[f] = bs

	if !receivedStartReq {
		// This will ask the server to make a start request.
		w.WriteHeader(http.StatusResetContent)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sync(w http.ResponseWriter, r *http.Request) {
	log.Printf("received sync request from client")
	for f := range fullData {
		syncDetails.Tick = fullData[f].tick
		syncDetails.Rtdelay = int64(time.Since(fullData[f].at).Seconds())
		syncDetails.Rcvage = int64(time.Since(latestFullData).Seconds())

		fg, err := strconv.Atoi(f)
		if err != nil {
			panic(err)
		}
		syncDetails.Fragment = fg

		break
	}

	bs, err := json.Marshal(&syncDetails)
	if err != nil {
		panic(err)
	}

	log.Printf("sending %s as response to sync", string(bs))

	w.Write(bs)
}

func clientStart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f := vars["fragment_number"]
	log.Printf("received start request with fragment number %s from client", f)

	if _, ok := startData[f]; ok {
		w.Write(startData[f])
	} else {
		log.Printf("missing start data for fragment %s", f)
		w.WriteHeader(http.StatusOK)
	}
}

func clientFull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f := vars["fragment_number"]
	log.Printf("received full request with fragment number %s from client", f)

	if _, ok := fullData[f]; ok {
		w.Write(fullData[f].data)
	} else {
		log.Printf("missing full data for fragment %s", f)
		w.WriteHeader(http.StatusOK)
	}
}

func clientDelta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f := vars["fragment_number"]
	log.Printf("received delta request with fragment number %s from client", f)

	if _, ok := deltaData[f]; ok {
		w.Write(deltaData[f])
	} else {
		log.Printf("missing delta data for fragment %s", f)
		w.WriteHeader(http.StatusOK)
	}
}
