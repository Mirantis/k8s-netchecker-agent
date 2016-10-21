package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// NetcheckerServiceName is DNS name for netchecker service
// and is used in URL that provides access to netchecker server
// var NetcheckerServiceName = "netchecker-service"
var NetcheckerServiceName = "127.0.0.1"

// NetcheckerPort is exposed by netchecker service port and
// is used in URI that provides access to netchecker server
var NetcheckerPort = "8081"

// NetcheckerAgentsEndpoint is part of URI providing access to
// netchecker server that designates endpoint for retrieving
// and updating of the agents
var NetcheckerAgentsEndpoint = "/api/v1/agents/"

// Payload describes request data to be sent to netchecker server
type Payload struct {
	ReportInterval string    `json:"report_interval"`
	PodName        string    `json:"podname"`
	HostDate       time.Time `json:"hostdate"`
}

func main() {
	var ReportInterval string
	if envVar := os.Getenv("REPORT_INTERVAL"); envVar != "" {
		ReportInterval = envVar
	} else {
		ReportInterval = "10"
	}

	P := &Payload{
		ReportInterval: ReportInterval,
		PodName:        "my_pod",
		// PodName:        os.Getenv("MY_POD_NAME"),
	}

	url := "http://" + NetcheckerServiceName + ":" + NetcheckerPort + NetcheckerAgentsEndpoint + P.PodName
	fmt.Printf("netchecker-server URI --> %s\n", url)

	client := &http.Client{}

	for {
		P.HostDate = time.Now()

		m, _ := json.Marshal(P)

		fmt.Printf("Marshaled payload --> %s\n", m)

		req, _ := http.NewRequest("POST", url, bytes.NewReader(m))
		req.Header.Add("Content-Type", "application/json")

		client.Do(req)

		secondsNumber, _ := strconv.Atoi(ReportInterval)
		fmt.Printf("Sleep for %d second(s)\n", secondsNumber)
		time.Sleep(time.Duration(secondsNumber) * time.Second)
	}
}
