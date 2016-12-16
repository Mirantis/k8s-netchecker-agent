package main

import (
	"flag"
	"os"

	"github.com/aateem/mcp-netchecker-agent/agent"
	"github.com/golang/glog"
)

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

	err := agent.StartSending(serverEndpoint, reportInterval)
	if err != nil {
		glog.Errorf("Cancel sending due to error. Details: %v", err)
		os.Exit(1)
	}
}
