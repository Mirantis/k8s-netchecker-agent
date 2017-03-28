# Network checker agent

[![Build Status](https://goo.gl/ZeP2WM)](https://goo.gl/gQxrfz)
[![Stories in Progress](https://goo.gl/D8FyiN)](https://goo.gl/kJ8CYj)
[![Go Report Card](https://goo.gl/eHnKRa)](https://goo.gl/Q6HZdP)
[![Code Climate](https://goo.gl/51gpev)](https://goo.gl/n5nWM4)
[![License Apache 2.0](https://goo.gl/joRzTI)](https://goo.gl/QKY5kg)
[![Docker Pulls](https://goo.gl/bsXWBB)](https://goo.gl/U0l9UK)

The agent is a simple application that collects network related information
from a host and sends it to designated network checker server end point.

## Usage

`agent -v=5 -alsologtostderr=true -serverendopoint=0.0.0.0:8888  
-reportinterval=5`

## Building binary, running tests and preparing docker image

Build static binary inside of intermediate build container
`make build-containerized`

Prepare docker image
`make build-image`

Run tests inside intermediate container
`export DOCKER_BUILD=yes; make unit`
