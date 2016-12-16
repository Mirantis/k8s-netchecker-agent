package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	//"github.com/vishvananda/netlink"
)

// NetcheckerAgentsEndpoint is part of URI providing access to
// netchecker server that designates endpoint for retrieving
// and updating of the agents
var NetcheckerAgentsEndpoint = "/api/v1/agents"

// EnvVarPodName is name of environment variable that stores k8s pod name
// inside which the agent is running
var EnvVarPodName = "MY_POD_NAME"

// Payload describes request data to be sent to netchecker server
type Payload struct {
	ReportInterval string              `json:"report_interval"`
	PodName        string              `json:"podname"`
	HostDate       time.Time           `json:"hostdate"`
	LookupHost     []string            `json:"nslookup"`
	IPs            map[string][]string `json:"ips"`
}

// Client represents HTTP client interface
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// SendMarshaled marshals given payload, constructs http request and
// sends to the server URL
func SendMarshaled(c Client, p *Payload, requestURL string) (*http.Response, error) {
	glog.V(10).Infof("Request payload before marshaling: %v", p)
	m, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	glog.V(5).Infof("Send payload via URL: %v", requestURL)
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(m))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	return resp, err
}

// AnalyzeResponse is a place holder for more sophisticated response analysis
func AnalyzeResponse(resp *http.Response) error {
	if resp.StatusCode != 200 {
		return errors.New("Response from the server is not OK")
	}
	return nil
}

// StartSending constructs and sends requests to the server in infinite loop
func StartSending(serverEndpoint, reportInterval string) error {
	var podName string
	if podName = os.Getenv(EnvVarPodName); len(podName) == 0 {
		return errors.New("Environment variable MY_POD_NAME is not set")
	}

	P := &Payload{
		ReportInterval: reportInterval,
		PodName:        podName,
	}

	reqURL := url.URL{
		Scheme: "http",
		Host:   serverEndpoint,
		Path:   strings.Join([]string{NetcheckerAgentsEndpoint, podName}, "/"),
	}
	client := &http.Client{}

	for {
		// periodically flush cache to log file because app is run as
		// daemon
		glog.Flush()

		P.HostDate = time.Now()

		addrs, err := net.LookupHost(strings.Split(serverEndpoint, ":")[0])
		if err != nil {
			glog.Errorf("DNS look up host error. Details: %v", err)
		}
		P.LookupHost = addrs

		P.IPs = linkV4Info()

		resp, err := SendMarshaled(client, P, reqURL.String())
		if err != nil {
			return err
		}

		// let's just treat unsuccessful response as a warning
		err = AnalyzeResponse(resp)
		if err != nil {
			glog.Warning("Analyzing response fails. Detals: %v", err)
		}

		sleepSeconds, err := strconv.Atoi(reportInterval)
		if err != nil {
			return err
		}

		glog.V(4).Infof("Sleep for %v second(s)", sleepSeconds)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
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
