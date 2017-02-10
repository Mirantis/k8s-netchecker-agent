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
	"github.com/vishvananda/netlink"
)

const (
	EnvVarPodName            = "MY_POD_NAME"
	NetcheckerAgentsEndpoint = "/api/v1/agents"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type Payload struct {
	ReportInterval     int                 `json:"report_interval"`
	PodName            string              `json:"podname"`
	HostDate           time.Time           `json:"hostdate"`
	LookupHost         map[string][]string `json:"nslookup"`
	IPs                map[string][]string `json:"ips"`
	ZeroExtenderLength int                 `json:"zero_extender_length"`
}

type IfaceProcessor struct {
	CommLinkAddr string
	CommLinkMTU  int
}

func (ifp *IfaceProcessor) ProcessIifaces() (map[string][]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			glog.Errorf("Failed to get addresses of interface %v. Details: %v",
				iface.Name, err)
			continue
		}

		addrRes := []string{}
		for _, addr := range addrs {
			addrS := addr.String()
			if addrS == ifp.CommLinkAddr {
				ifp.CommLinkMTU = iface.MTU
			}
			addrRes = append(addrRes, addrS)
		}
		result[iface.Name] = addrRes
	}
	return result, nil
}

func sendInfo(srvEndpoint, podName string, repIntl int, checkMTU bool, cl Client) (*http.Response, error) {
	reqURL := (&url.URL{
		Scheme: "http",
		Host:   srvEndpoint,
		Path:   strings.Join([]string{NetcheckerAgentsEndpoint, podName}, "/"),
	}).String()

	hostname, _, err := net.SplitHostPort(srvEndpoint)
	ips, err := net.LookupIP(hostname)
	if err != nil {
		glog.Errorf("DNS look up host error. Details: %v", err)
	}

	lookup := map[string][]string{hostname: []string{}}
	for _, ip := range ips {
		lookup[hostname] = append(lookup[hostname], ip.String())
	}

	//to test that there are no problems with network packet's
	//fragmentation let's extended marshaled payload so it
	//exceeds the active (which is in the communication with the server)
	//link MTU value in case it is bigger than the data being sent
	linkAddr := ""
	if a := getRouteSrc(ips); len(a) != 0 && checkMTU {
		linkAddr = a
	}
	iProcessor := &IfaceProcessor{CommLinkAddr: linkAddr}
	linkInfo, err := iProcessor.ProcessIifaces()
	if err != nil {
		glog.Errorf("Error while processing interfaces. Details: %v", err)
	}

	data := &Payload{
		HostDate:           time.Now(),
		IPs:                linkInfo,
		ReportInterval:     repIntl,
		PodName:            podName,
		LookupHost:         lookup,
		ZeroExtenderLength: iProcessor.CommLinkMTU,
	}

	glog.V(10).Infof("Request payload before marshaling: %v", data)
	marshaled, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	glog.V(10).Infof("Extend payload by %v zero bytes", iProcessor.CommLinkMTU)
	zeroExtender := make([]byte, iProcessor.CommLinkMTU)
	marshaled = bytes.Join([][]byte{marshaled, zeroExtender}, []byte{})

	glog.V(5).Infof("Send payload via URL: %v", reqURL)
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(marshaled))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := cl.Do(req)
	return resp, err
}

func getRouteSrc(dsts []net.IP) (src string) {
	for _, dst := range dsts {
		route, err := netlink.RouteGet(dst)
		if err != nil {
			glog.Errorf("Failed to get route to the server's IP. Details: %v", err)
			continue
		}
		if len(route) == 0 {
			glog.Infof("There are no routes for this IP")
			continue
		}
		return route[0].Src.String()
	}
	return
}

func main() {
	var (
		serverEndpoint string
		reportInterval int
		checkMTU       bool
	)

	flag.StringVar(&serverEndpoint, "serverendpoint", "netchecker-service:8081", "Netchecker server endpoint (host:port)")
	flag.IntVar(&reportInterval, "reportinterval", 60, "Agent report interval")
	flag.BoolVar(
		&checkMTU, "checkmtu", false,
		"Extend sent payload by zero bytes of MTU value length. Used in case when agent's data is less then MTU size to test problems with the fragmentation.")
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

		resp, err := sendInfo(serverEndpoint, podName, reportInterval, checkMTU, client)
		if err != nil {
			glog.Errorf("Error while sending info. Details: %v", err)
		} else {
			glog.Infof("Response status code: %v", resp.StatusCode)
		}
	}
}
