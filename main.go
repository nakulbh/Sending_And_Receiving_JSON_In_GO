package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Request struct {
	Format string `json:"format"`
	TZ     string `json:"tz"`
}

type Resp struct {
	Time time.Time `json:"time"`
}

type Error struct {
	Error string `json:"error"`
}

type Response struct {
	Time string
}

func getTime(w http.ResponseWriter, r *http.Request) {
	var req Request

	w.Header().Set("Content-Type", "enconding/json")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}
	r.Body.Close()

	var tz *time.Location = time.Local

	if req.TZ != "" {
		var err error
		tz, err = time.LoadLocation(req.TZ)
		if err != nil || tz == nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(Error{err.Error()})
			return
		}
	}

	format := time.RFC3339
	if req.Format != "" {
		format = req.Format
	}

	resp := Response{Time: time.Now().In(tz).Format(format)}
	json.NewEncoder(w).Encode(resp)
}

var client = &http.Client{Timeout: 2 * time.Second}

func sendRequest(tz, format string) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(Request{TZ: tz, Format: format})
	log.Printf("request body: %v", body)
	req, err := http.NewRequestWithContext(context.TODO(), "GET", "http://localhost:8080", body)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Write(os.Stdout)
	resp.Body.Close()
}

func main() {
	server := http.Server{Addr: ":8080", Handler: http.HandlerFunc(getTime)}

	go server.ListenAndServe()

	sendRequest("", "")

	sendRequest("America/Los_Angels", time.RFC3339)

	sendRequest("America/New_York", time.RFC3339)

	sendRequest("faketz", "")
}
