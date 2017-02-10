package main

import (
	"encoding/json"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type FakeHTTPClient struct {
	recordedRequest *http.Request
}

func (c *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.recordedRequest = req
	return nil, nil
}

func TestSendInfo(t *testing.T) {
	fakeClient := &FakeHTTPClient{}

	serverEndPoint := "localhost:8888"
	reportInterval := 5
	podName := "test-pod"
	_, err := sendInfo(serverEndPoint, podName, reportInterval, false, fakeClient)
	if err != nil {
		t.Errorf("sendInfo should not return error. Details: %v", err)
	}

	expectedURL := "http://" + serverEndPoint + NetcheckerAgentsEndpoint + "/" + podName
	if reqURL := fakeClient.recordedRequest.URL.String(); reqURL != expectedURL {
		t.Errorf("URL used in the request is not as expected. Actual %v", reqURL)
	}

	if fakeClient.recordedRequest.Method != "POST" {
		t.Error("Request does not use proper method (should be POST)")
	}

	ctJSON := false
	for _, ct := range fakeClient.recordedRequest.Header["Content-Type"] {
		if ct == "application/json" {
			ctJSON = true
			break
		}
	}

	if ctJSON == false {
		t.Error("Content-Type header must be properly set for header (correct - application/json)")
	}

	body := make([]byte, fakeClient.recordedRequest.ContentLength)
	_, err = fakeClient.recordedRequest.Body.Read(body)
	if err != nil {
		t.Errorf("Error should not occur while reading fake requests's body. Details: %v", err)
	}

	payload := &Payload{}
	err = json.Unmarshal(body, payload)
	if err != nil {
		t.Errorf("Error should not occur while unmarshaling fake request's payload. Details: %v", err)
	}

	iproc := &IfaceProcessor{}
	expectedIPs, err := iproc.ProcessIifaces()
	if err != nil {
		t.Errorf("Error should not occured while retrieving expected ifaces. Details: %v", err)
	}
	if !reflect.DeepEqual(expectedIPs, payload.IPs) {
		t.Errorf("IPs data from payload is not as expected. expected %v\n actual %v", expectedIPs, payload.IPs)
	}

	expectedHost := strings.Split(serverEndPoint, ":")[0]
	addrs, err := net.LookupHost(expectedHost)
	if err != nil {
		t.Errorf("DNS look up error should not occur. Details: %v", err)
	}
	if !reflect.DeepEqual(payload.LookupHost, map[string][]string{expectedHost: addrs}) {
		t.Errorf("LookupHost data from the payload is not as expected")
	}

	if payload.ZeroExtenderLength != 0 {
		t.Errorf("ZeroExtenderLength should be %v instead it is %v",
			0, payload.ZeroExtenderLength)
	}
}
