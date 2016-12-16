package agent

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
)

const NetcheckerAgentsEndpoint = "/api/v1/agents"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type Payload struct {
	ReportInterval string              `json:"report_interval"`
	PodName        string              `json:"podname"`
	HostDate       time.Time           `json:"hostdate"`
	LookupHost     []string            `json:"nslookup"`
	IPs            map[string][]string `json:"ips"`
}

func SendInfo(serverEndpoint, reportInterval, podName string, client Client) (*http.Response, error) {
	reqURL := (&url.URL{
		Scheme: "http",
		Host:   serverEndpoint,
		Path:   strings.Join([]string{NetcheckerAgentsEndpoint, podName}, "/"),
	}).String()

	payload := &Payload{
		HostDate:       time.Now(),
		IPs:            linkV4Info(),
		ReportInterval: reportInterval,
		PodName:        podName,
	}
	addrs, err := net.LookupHost(strings.Split(serverEndpoint, ":")[0])
	if err != nil {
		glog.Errorf("DNS look up host error. Details: %v", err)
	}
	payload.LookupHost = addrs

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

	resp, err := client.Do(req)
	return resp, err
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
