# Makefile for Tharsis CLI

MODULE = $(shell go list -m)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "1.0.0")
PACKAGES := $(shell go list ./... | grep -v /vendor/)
BINARY=tharsis
BUILD_PATH=$(MODULE)/cmd/tharsis
GCFLAGS:=-gcflags all=-trimpath=${PWD}
LDFLAGS := -ldflags "-X main.Version=${VERSION}"

## build the binaries
.PHONY: build
build:  ## build the Tharsis CLI binary
	CGO_ENABLED=0 go build ${LDFLAGS} -a -o ${BINARY} $(BUILD_PATH)

.PHONY: lint
lint: ## run golint on all Go package
	@revive $(PACKAGES)

.PHONY: vet
vet: ## run golint on all Go package
	@go vet $(PACKAGES)

.PHONY: fmt
fmt: ## run "go fmt" on all Go packages
	@go fmt $(PACKAGES)

.PHONY: test
test: ## run unit tests
	go test -v ./...

release:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_darwin_amd64  $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_darwin_arm64  $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_arm   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_386     $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_amd64   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_arm     $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_arm64   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_openbsd_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_openbsd_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_solaris_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_windows_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_windows_amd64 $(BUILD_PATH)

# The End.
