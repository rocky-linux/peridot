# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules

GO_FLAGS= CGO_ENABLED=0
GO_LDFLAGS= -ldflags=""
GO_BUILD_CMD=$(GO_FLAGS) go build $(GO_LDFLAGS)

BINARY_NAME=go-xmlrpc
BUILD_DIR=build

.PHONY: all
all: clean lint test build

#--------------------------------
# Validation steps
#--------------------------------

.PHONY: lint
lint:
	@echo "Linting code..."
	@go vet ./...

.PHONY: test
test: pre-build
	@echo "Running tests..."
	@go test -short -coverprofile=build/coverage.txt -covermode=atomic ./...

#--------------------------------
# Build steps
#--------------------------------

.PHONY: pre-build
pre-build:
	@mkdir -p $(BUILD_DIR)

.PHONY: build
build:
	@echo "Building..."
	$(GO_BUILD_CMD)

#--------------------------------
# Docker steps
#--------------------------------

.PHONY: docker
docker:
# Build a new image (delete old one)
	docker build --force-rm --build-arg GOPROXY -t $(BINARY_NAME) .

.PHONY: build-in-docker
build-in-docker: docker
# Force-stop any containers with this name
	docker rm -f $(BINARY_NAME) || true
# Create a new container with newly built image (but don't run it)
	docker create --name $(BINARY_NAME) $(BINARY_NAME)
# Copy over the binary to disk (from container)
	docker cp '$(BINARY_NAME):/opt/' $(BUILD_DIR)
# House-keeping: removing container
	docker rm -f $(BINARY_NAME)

#--------------------------------
# Cleanup steps
#--------------------------------

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -Rf $(BUILD_DIR)
