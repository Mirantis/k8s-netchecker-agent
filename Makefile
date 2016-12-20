BUILD_DIR=_output
BUILD_CONTAINER_NAME=mcp-netchecker-agent.build
DEPLOY_CONTAINER_NAME=aateem/mcp-netchecker-agent
DEPLOY_CONTAINER_TAG=golang

prepare-build-container: Dockerfile.build
	docker build -f Dockerfile.build -t $(BUILD_CONTAINER_NAME) .

build-containerized:  $(BUILD_DIR) prepare-build-container
	docker run --rm  \
		-v $(PWD):/go/src/github.com/aateem/mcp-netchecker-agent:ro \
		-v $(PWD)/$(BUILD_DIR):/go/src/github.com/aateem/mcp-netchecker-agent/$(BUILD_DIR) \
		-w /go/src/github.com/aateem/mcp-netchecker-agent/ \
		$(BUILD_CONTAINER_NAME) bash -c '\
	    	CGO_ENABLED=0 go build -x -o $(BUILD_DIR)/agent -ldflags "-s -w" agent.go &&\
			chown -R $(shell id -u):$(shell id -u) $(BUILD_DIR)'

prepare-deploy-container: build-containerized
	docker build -t $(DEPLOY_CONTAINER_NAME):$(DEPLOY_CONTAINER_TAG) .

test-containerized: prepare-build-container
	docker run --rm \
		-v $(PWD):/go/src/github.com/aateem/mcp-netchecker-agent:ro \
		$(BUILD_CONTAINER_NAME) go test ./...

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

build-local: clean-build $(BUILD_DIR)
	go build -v -o $(BUILD_DIR)/agent agent.go

.PHONY: clean-build
clean-build:
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean-all
clean-all: clean-build
	docker rmi $(BUILD_CONTAINER_NAME)
	docker rmi $(DEPLOY_CONTAINER_NAME):$(DEPLOY_CONTAINER_TAG)

.PHONY: get-deps
get-deps:
	go get github.com/golang/glog
