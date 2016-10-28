package main

import (
	"flag"
	"os"

	"github.com/aateem/mcp-netchecker-agent/netcheckclient"
	"github.com/golang/glog"
)

func main() {
	var (
		serverEndpoint string
		reportInterval string
	)

	flag.StringVar(&serverEndpoint, "serverendpoint", "127.0.0.1:8081", "Host address and port on which netchecker server is listening")
	flag.StringVar(&reportInterval, "reportinterval", "10", "Agent report interval")
	flag.Parse()

	glog.V(4).Info("Starting agent")

	err := netcheckclient.StartSending(serverEndpoint, reportInterval)
	if err != nil {
		glog.Errorf("Cancel sending due to error. Err --> %v\n", err)
		os.Exit(1)
	}
}
