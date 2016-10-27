package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
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

// DefaultReportInterval value is used when REPORT_INTERVAL
// environment variable is not set
var DefaultReportInterval = "10"

// Payload describes request data to be sent to netchecker server
type Payload struct {
	ReportInterval string    `json:"report_interval"`
	PodName        string    `json:"podname"`
	HostDate       time.Time `json:"hostdate"`
}

// EnvVars retrieves needed environment variables
// MY_POD_NAME must be set. In case it is not, error is returned
func EnvVars() (reportInterval, podName string, err error) {
	if podName = os.Getenv("MY_POD_NAME"); len(podName) == 0 {
		return "", "", errors.New("Environment variable MY_POD_NAME is not set")
	}
	if reportInterval = os.Getenv("REPORT_INTERVAL"); len(reportInterval) == 0 {
		reportInterval = DefaultReportInterval
	}

	return reportInterval, podName, nil
}

// PrepareAndSendRequest completes the payload with dynamically changing data
// and constructs POST request that is sent by HTTP client singleton
func PrepareAndSendRequest(c *http.Client, p *Payload, url string) (*http.Response, error) {
	p.HostDate = time.Now()

	m, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("Marshaled payload --> %s\n", m)

	req, err := http.NewRequest("POST", url, bytes.NewReader(m))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	return resp, err
}

func main() {
	flag.Parse()

	glog.V(4).Info("Starting agent")

	ReportInterval, PodName, err := EnvVars()
	if err != nil {
		glog.Errorf("Error while retrieving environment variables. Err --> %v\n", err)
		os.Exit(1)
	}

	url := "http://" + NetcheckerServiceName + ":" + NetcheckerPort + NetcheckerAgentsEndpoint + PodName
	glog.V(4).Infof("Netchecker server URL --> %s", url)

	P := &Payload{
		ReportInterval: ReportInterval,
		PodName:        PodName,
	}

	client := &http.Client{}

	sleepSeconds, err := strconv.Atoi(ReportInterval)
	if err != nil {
		glog.Errorf("Fail to convert report interavl to string. Err --> %v\n", err)
	}

	for {
		// the application is run as daemon thus periodically flush the logs
		glog.Flush()
		_, err := PrepareAndSendRequest(client, P, url)
		if err != nil {
			glog.Errorf("Error while sending request. Err --> %v\n", err)
			os.Exit(1)
		}

		glog.V(4).Infof("Sleep for %v second(s)\n", sleepSeconds)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}
