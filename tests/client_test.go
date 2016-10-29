package clienttest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aateem/mcp-netchecker-agent/netcheckclient"
)

func TestStartSendingFailIfEnvVarNotSet(t *testing.T) {
	if _, set := os.LookupEnv(netcheckclient.EnvVarPodName); set == false {
		os.Unsetenv(netcheckclient.EnvVarPodName)
	}

	err := netcheckclient.StartSending("", "")
	if err == nil {
		t.Error("Error is expected to be returned when $MY_POD_NAME is unset")
	}
}

func TestAnalyzeResponse(t *testing.T) {
	fakeResp := &http.Response{
		StatusCode: 400,
	}

	err := netcheckclient.AnalyzeResponse(fakeResp)
	if err == nil {
		t.Error("Fake resp must return error in case resp is not OK")
	}

	fakeResp.StatusCode = 200
	err = netcheckclient.AnalyzeResponse(fakeResp)

	if err != nil {
		t.Error("Fake resp must not return error in case resp is OK")
	}
}

type FakeHTTPClient struct {
	recordedRequest *http.Request
}

func (c *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.recordedRequest = req
	return &http.Response{}, nil
}

func TestSendMarshaled(t *testing.T) {
	fakeClient := &FakeHTTPClient{}

	fakePayload := &netcheckclient.Payload{
		PodName:        "test_pod",
		ReportInterval: "10",
		HostDate:       time.Now(),
	}

	fakeURL := "http://fake-url.edu"

	resp, err := netcheckclient.SendMarshaled(fakeClient, fakePayload, fakeURL)

	if resp == nil && err != nil {
		t.Error("SendMarshalled should not return error with given data")
	}

	if parsedURL, _ := url.Parse(fakeURL); *parsedURL != *fakeClient.recordedRequest.URL {
		t.Errorf("Request does not use provided URL")
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

	fakeMarshaled, _ := json.Marshal(fakePayload)
	byteBody := make([]byte, fakeClient.recordedRequest.ContentLength)
	fakeClient.recordedRequest.Body.Read(byteBody)

	if equal := bytes.Equal(byteBody, fakeMarshaled); !equal {
		t.Error("Request's body is not as expected")
	}
}
