package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aateem/mcp-netchecker-agent/agent"
	"github.com/golang/glog"
)

const EnvVarPodName = "MY_POD_NAME"

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

		resp, err := agent.SendInfo(serverEndpoint, reportInterval, podName, client)
		if err != nil {
			glog.Errorf("Error while sending info. Details: %v", err)
		} else {
			glog.Infof("Response status code: %v", resp.StatusCode)
		}
	}
}
