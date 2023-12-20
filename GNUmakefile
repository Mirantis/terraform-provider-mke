default: test

LOCAL_TAG=$(shell git describe --tags)

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env .GOARCH)

# Run local unit tests
.PHONY: test
test:
	go test ./... -timeout 120m

# Run local unit tests (which use the acceptance test suite)
# @NOTE these are run with the provider/resources in testing mode
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# FMT the source code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint by running golangci-lint in a docker container
.PHONY: lint
lint:
	docker run -ti --rm -v "$(CURDIR):/data" -w "/data" golangci/golangci-lint:latest golangci-lint run

# Local install of the plugin
.PHONY: local
local:
	GORELEASER_CURRENT_TAG="$(LOCAL_TAG)" goreleaser build --clean --single-target --skip=validate
	# @SEE README.md on how to use the locally built plugin
