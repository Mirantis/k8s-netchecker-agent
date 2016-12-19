BUILD_DIR=_output
BUILD_CONTAINER_NAME=mcp-netchecker-agent.build
DEPLOY_CONTAINER_NAME=aateem/mcp-netchecker-agent
DEPLOY_CONTAINER_TAG=golang

prepare-build-container: Dockerfile.build
	docker build -f Dockerfile.build -t $(BUILD_CONTAINER_NAME) .

dist:
	mkdir -p dist

build-containerized: prepare-build-container dist
	docker run --rm  \
		-v ${PWD}:/go/src/github.com/aateem/mcp-netchecker-agent:ro \
		-v ${PWD}/dist:/go/src/github.com/aateem/mcp-netchecker-agent/dist \
		-w /go/src/github.com/aateem/mcp-netchecker-agent/ \
		$(BUILD_CONTAINER_NAME) bash -c '\
	    	CGO_ENABLED=0 go build -x -o dist/agent -ldflags "-s -w" cmd/agent.go &&\
			chown -R $(shell id -u):$(shell id -u) dist'

prepare-deploy-container: build-containerized
	docker build -t $(DEPLOY_CONTAINER_NAME):$(DEPLOY_CONTAINER_TAG) .

test-containerized: prepare-build-container
	docker run --rm \
		-v ${PWD}:/go/src/github.com/aateem/mcp-netchecker-agent:ro \
		$(BUILD_CONTAINER_NAME) go test ./tests

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

clean :
	rm -rf dist
