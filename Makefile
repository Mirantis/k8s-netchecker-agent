BUILD_DIR=_output
UTILITY_CONTAINER_NAME=k8s-netchecker-agent.build
RELEASE_CONTAINER_NAME=aateem/k8s-netchecker-agent
RELEASE_CONTAINER_TAG=golang

build-utility-image: Dockerfile.build
	docker build -f Dockerfile.build -t $(UTILITY_CONTAINER_NAME) .

go-build-containerized:  $(BUILD_DIR) build-utility-image
	docker run --rm  \
		-v $(PWD):/go/src/github.com/aateem/mcp-netchecker-agent:ro \
		-v $(PWD)/$(BUILD_DIR):/go/src/github.com/aateem/mcp-netchecker-agent/$(BUILD_DIR) \
		-w /go/src/github.com/aateem/mcp-netchecker-agent/ \
		$(UTILITY_CONTAINER_NAME) bash -c '\
	    	CGO_ENABLED=0 go build -x -o $(BUILD_DIR)/agent -ldflags "-s -w" agent.go &&\
			chown -R $(shell id -u):$(shell id -u) $(BUILD_DIR)'

build-release-image: go-build-containerized
	docker build -t $(RELEASE_CONTAINER_NAME):$(RELEASE_CONTAINER_TAG) .

test-containerized: build-utility-image
	docker run --rm \
		-v $(PWD):/go/src/github.com/Mirantis/k8s-netchecker-agent:ro \
		$(UTILITY_CONTAINER_NAME) go test -v $(glide novendor)

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

go-build-local: $(BUILD_DIR)
	go build -v -o $(BUILD_DIR)/agent agent.go

go-rebuild-local: clean-build build-local

.PHONY: clean-build
clean-build:
	rm -rf $(BUILD_DIR)

.PHONY: test-local
test-local:
	go test -v $(glide novendor)

.PHONY: clean-all
clean-all: clean-build
	docker rmi $(UTILITY_CONTAINER_NAME)
	docker rmi $(RELEASE_CONTAINER_NAME):$(RELEASE_CONTAINER_TAG)

.PHONY: get-deps
get-deps:
	glide install --strip-vendor
