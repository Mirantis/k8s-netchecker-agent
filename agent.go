// Copyright 2017 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/tcnksm/go-httpstat"
	"io/ioutil"
)

const (
	// EnvVarPodName is a pod name variable in pod's environment
	EnvVarPodName = "MY_POD_NAME"
	// EnvVarNodeName is a node name variable in pod's environment
	EnvVarNodeName = "MY_NODE_NAME"
	// NetcheckerAgentsEndpoint is a server URI where keepalive message is sent to
	NetcheckerAgentsEndpoint = "/api/v1/agents"
	// NetcheckerProbeEndpoint is a server URI that just provides simple 200 answer
	NetcheckerProbeEndpoint = "/api/v1/ping"
)

// Client is a REST API client interface that matches standard http.Client struct and
// references only Do() method from there.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Payload structure for keepalive message sent from agent to server
type Payload struct {
	ReportInterval int                 `json:"report_interval"`
	NodeName       string              `json:"nodename"`
	PodName        string              `json:"podname"`
	HostDate       time.Time           `json:"hostdate"`
	LookupHost     map[string][]string `json:"nslookup"`
	IPs            map[string][]string `json:"ips"`
	NetworkProbes  []ProbeResult       `json:"network_probes"`
	ZeroExtender   []int8              `json:"zero_extender"`
}

// ProbeResult structure for network probing results
type ProbeResult struct {
	EndPoint        string
	Total           int
	ContentTransfer int
}

func sendInfo(srvEndpoint, podName string, nodeName string, probeRes []ProbeResult,
	repIntl int, extenderLength int, cl Client) (*http.Response, error) {

	reqURL := (&url.URL{
		Scheme: "http",
		Host:   srvEndpoint,
		Path:   strings.Join([]string{NetcheckerAgentsEndpoint, podName}, "/"),
	}).String()

	payload := &Payload{
		HostDate:       time.Now(),
		IPs:            linkV4Info(),
		ReportInterval: repIntl,
		NodeName:       nodeName,
		PodName:        podName,
		LookupHost:     nsLookUp(srvEndpoint),
		NetworkProbes:  probeRes,
		ZeroExtender:   make([]int8, extenderLength),
	}

	glog.V(10).Infof("Request payload before marshaling: %v", payload)
	marshaled, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	glog.V(5).Infof("Send payload via URL: %v", reqURL)
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(marshaled))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := cl.Do(req)
	return resp, err
}

func nsLookUp(endpoint string) map[string][]string {
	hostname, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		glog.Errorf("Error while splitting endpont %v. Details: %v", endpoint, err)
	}
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		glog.Errorf("DNS look up host error. Details: %v", err)
	}
	result := map[string][]string{
		hostname: addrs,
	}

	return result
}

func linkV4Info() map[string][]string {
	ifaces, err := net.Interfaces()
	if err != nil {
		glog.Errorf("Fail to collect information on ifaces. Details: %v", err)
		return nil
	}

	result := map[string][]string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			glog.Errorf("Fail get addresses for iface %v. Details: %v", i.Name, err)
			continue
		}

		addrArr := []string{}
		for _, a := range addrs {
			addrArr = append(addrArr, a.String())
		}
		result[i.Name] = addrArr
	}
	glog.V(10).Infof("Addresses of host's links: %v", result)

	return result
}

func httpProbe(endpoints []string, probeRes []ProbeResult, timeout time.Duration) {
	for idx, ep := range endpoints {
		reqURL := (&url.URL{
			Scheme: "http",
			Host:   ep,
			Path:   NetcheckerProbeEndpoint,
		}).String()

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			glog.Fatal(err)
		}

		// Create go-httpstat powered context and pass it to http.Request
		var result httpstat.Result
		ctx := httpstat.WithHTTPStat(req.Context(), &result)
		req = req.WithContext(ctx)

		client := http.DefaultClient
		client.Timeout = timeout
		res, err := client.Do(req)
		if err != nil {
			glog.Fatal(err)
		}

		if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
			glog.Fatal(err)
		}
		res.Body.Close()
		t := time.Now()
		result.End(t)

		curRes := new(ProbeResult)
		curRes.EndPoint = ep
		curRes.Total = int(result.Total(t) / time.Millisecond)
		curRes.ContentTransfer = int(result.ContentTransfer(t) / time.Millisecond)
		probeRes[idx] = *curRes

		glog.V(5).Infof("HTTP Probe (%v): Total: %d ms, Content Transfer: %d ms",
			reqURL, curRes.Total, curRes.ContentTransfer)
	}
}

func main() {
	var (
		serverEndpoint string
		reportInterval int
		extenderLength int
	)

	flag.StringVar(&serverEndpoint, "serverendpoint", "netchecker-service:8081", "Netchecker server endpoint (host:port)")
	flag.IntVar(&reportInterval, "reportinterval", 60, "Agent report interval")
	flag.IntVar(&extenderLength, "zeroextenderlength", 1500,
		fmt.Sprint(
			"Length of zero bytes extender array ",
			"that will be added to the agent's payload ",
			"in case its size is less than MTU value. ",
			"Is used to reveal problems with network packets ",
			"fragmentation",
		),
	)
	flag.Parse()

	glog.V(5).Infof("Provided server endpoint: %v", serverEndpoint)
	glog.V(5).Infof("Provided report interval: %v", reportInterval)

	glog.Info("Starting agent")

	var podName string
	var nodeName string
	if podName = os.Getenv(EnvVarPodName); len(podName) == 0 {
		glog.Error("Environment variable %s is not set. No point in sending info. Exiting", EnvVarPodName)
		os.Exit(1)
	}
	if nodeName = os.Getenv(EnvVarNodeName); len(nodeName) == 0 {
		glog.Error("Environment variable %s is not set.", EnvVarNodeName)
	}

	probeRes := make([]ProbeResult, 1)
	endPoints := []string{serverEndpoint}
	client := &http.Client{}
	for {
		go httpProbe(endPoints, probeRes, time.Duration(reportInterval-1)*time.Second)
		glog.V(4).Infof("Sleep for %v second(s)", reportInterval)
		time.Sleep(time.Duration(reportInterval) * time.Second)

		resp, err := sendInfo(serverEndpoint, podName, nodeName, probeRes, reportInterval, extenderLength, client)
		if err != nil {
			glog.Errorf("Error while sending info. Details: %v", err)
		} else {
			glog.Infof("Response status code: %v", resp.StatusCode)
		}
	}
}
