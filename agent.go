package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
)

const (
	EnvVarPodName            = "MY_POD_NAME"
	NetcheckerAgentsEndpoint = "/api/v1/agents"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type Payload struct {
	ReportInterval int                 `json:"report_interval"`
	PodName        string              `json:"podname"`
	HostDate       time.Time           `json:"hostdate"`
	LookupHost     map[string][]string `json:"nslookup"`
	IPs            map[string][]string `json:"ips"`
}

func sendInfo(srvEndpoint, podName string, repIntl int, cl Client) (*http.Response, error) {
	reqURL := (&url.URL{
		Scheme: "http",
		Host:   srvEndpoint,
		Path:   strings.Join([]string{NetcheckerAgentsEndpoint, podName}, "/"),
	}).String()

	payload := &Payload{
		HostDate:       time.Now(),
		IPs:            linkV4Info(),
		ReportInterval: repIntl,
		PodName:        podName,
		LookupHost:     nsLookUp(srvEndpoint),
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
	hostname := strings.Split(endpoint, ":")[0]
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

func main() {
	var (
		serverEndpoint string
		reportInterval int
	)

	flag.StringVar(&serverEndpoint, "serverendpoint", "netchecker-service:8081", "Netchecker server endpoint (host:port)")
	flag.IntVar(&reportInterval, "reportinterval", 60, "Agent report interval")
	flag.Parse()

	glog.V(5).Infof("Provided server endpoint: %v", serverEndpoint)
	glog.V(5).Infof("Provided report interval: %v", reportInterval)

	glog.Info("Starting agent")

	var podName string
	if podName = os.Getenv(EnvVarPodName); len(podName) == 0 {
		glog.Error("Environment variable MY_POD_NAME is not set. No point in sending info. Exiting")
		os.Exit(1)
	}

	client := &http.Client{}
	for {
		glog.V(4).Infof("Sleep for %v second(s)", reportInterval)
		time.Sleep(time.Duration(reportInterval) * time.Second)

		resp, err := sendInfo(serverEndpoint, podName, reportInterval, client)
		if err != nil {
			glog.Errorf("Error while sending info. Details: %v", err)
		} else {
			glog.Infof("Response status code: %v", resp.StatusCode)
		}
	}
}
