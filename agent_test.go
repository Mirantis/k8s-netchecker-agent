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
	nodeName := strings.Split(serverEndPoint, ":")[0]
	podName := "test-pod"
	extenderLength := 100
	netProbes := []ProbeResult{
		{"0.0.0.0:8081", 50, 1, 0, 0, 0, 0},
	}
	_, err := sendInfo(serverEndPoint, podName, nodeName, netProbes, reportInterval, extenderLength, fakeClient)
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

	if expectedIPs := linkV4Info(); !reflect.DeepEqual(expectedIPs, payload.IPs) {
		t.Errorf("IPs data from payload is not as expected. expected %v\n actual %v", expectedIPs, payload.IPs)
	}

	expectedHost := nodeName
	addrs, err := net.LookupHost(expectedHost)
	if err != nil {
		t.Errorf("DNS look up error should not occur. Details: %v", err)
	}
	if !reflect.DeepEqual(payload.LookupHost, map[string][]string{expectedHost: addrs}) {
		t.Errorf("LookupHost data from the payload is not as expected")
	}

	if len(payload.ZeroExtender) != extenderLength {
		t.Errorf("Extender should be of %v len instead it is %v", extenderLength,
			len(payload.ZeroExtender))
	}
	if payload.NodeName != nodeName {
		t.Errorf("Node name from payload (%v) does not match expected one (%v)", payload.NodeName, nodeName)
	}
	if !reflect.DeepEqual(payload.NetworkProbes, netProbes) {
		t.Errorf("NetworkProbes data from the payload is not as expected")
	}
}
