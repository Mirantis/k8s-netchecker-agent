package netcheckclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
)

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

func sendMarshaled(c *http.Client, p *Payload, url string) (*http.Response, error) {
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

// Place holder for more sophisticated response analysis
func analyzeResponse(resp *http.Response) {
	if resp.StatusCode != 200 {
		glog.Warning("Response from the server is not OK")
	}
}

// StartSending constructs and sends requests to the server in infinite loop
func StartSending(serverEndpoint, reportInterval string) error {
	var podName string
	if podName = os.Getenv("MY_POD_NAME"); len(podName) == 0 {
		return errors.New("Environment variable MY_POD_NAME is not set")
	}

	P := &Payload{
		ReportInterval: reportInterval,
		PodName:        podName,
		HostDate:       time.Now(),
	}

	url := "http://" + serverEndpoint + NetcheckerAgentsEndpoint + podName
	glog.V(4).Infof("Netchecker server URL --> %s", url)

	client := &http.Client{}

	for {
		// periodically flush cache to log file because app is run as
		// daemon
		glog.Flush()

		resp, err := sendMarshaled(client, P, url)
		if err != nil {
			return err
		}
		analyzeResponse(resp)

		sleepSeconds, err := strconv.Atoi(reportInterval)
		if err != nil {
			return err
		}

		glog.V(4).Infof("Sleep for %v second(s)\n", sleepSeconds)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}
