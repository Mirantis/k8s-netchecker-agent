Network checker agent
=====================

The agent is a simple application that collects network related
information from a host and sends it to designated network checker
server end point.

Usage
=====

`agent -v=5 -alsologtostderr=true -serverendopoint=0.0.0.0:8888 -reportinterval=5`

Building binary, running tests and preparing docker image
=======================================================

Build static binary inside of intermediate build container
`make build-containerized`

Prepare docker image
`make prepare-deployment-container`

Run tests inside intermediate container
`make tests`
