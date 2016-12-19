package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	ReportInterval string              `json:"report_interval"`
	PodName        string              `json:"podname"`
	HostDate       time.Time           `json:"hostdate"`
	LookupHost     []string            `json:"nslookup"`
	IPs            map[string][]string `json:"ips"`
}

func sendInfo(serverEndpoint, reportInterval, podName string, client Client) (*http.Response, error) {
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

func main() {
	var (
		serverEndpoint string
		reportInterval string
	)

	flag.StringVar(&serverEndpoint, "serverendpoint", "127.0.0.1:8081", "Host address and port on which netchecker server is listening")
	flag.StringVar(&reportInterval, "reportinterval", "60", "Agent report interval")
	flag.Parse()

	glog.V(5).Infof("Provided server endpoint: %v", serverEndpoint)
	glog.V(5).Infof("Provided report interval: %v", reportInterval)

	glog.Info("Starting agent")

	var podName string
	if podName = os.Getenv(EnvVarPodName); len(podName) == 0 {
		glog.Error("Environment variable MY_POD_NAME is not set. No point in sending info. Exiting")
		os.Exit(1)
	}

	sleepSeconds, err := strconv.Atoi(reportInterval)
	if err != nil {
		glog.Errorf("Error while processing report interval. Details: %v", err)
		os.Exit(1)
	}

	client := &http.Client{}
	for {
		glog.V(4).Infof("Sleep for %v second(s)", sleepSeconds)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)

		resp, err := sendInfo(serverEndpoint, reportInterval, podName, client)
		if err != nil {
			glog.Errorf("Error while sending info. Details: %v", err)
		} else {
			glog.Infof("Response status code: %v", resp.StatusCode)
		}
	}
}
