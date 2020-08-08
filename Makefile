GOCMD=go
GOCMDUSR=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOBUILDUSR=$(GOCMDUSR) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GO111MODULE=auto
CGO_ENABLED=1
BINARY_NAME=tccbot-backend
PACKAGE=$(BINARY_NAME)
GOPATH=$(HOME)/go
SRC=.

VERSION=$(shell git describe --tags --match 'v*' --always --abbrev=0)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GITHASH=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION} -X main.DateBuild=${BUILD_TIME}  -X main.GitHash=${GITHASH}"

build:
	@GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -v ${LDFLAGS} -o $(SRC)/bin/$(BINARY_NAME) $(SRC)/cmd

build-usr:
	@GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) $(GOBUILDUSR) -v ${LDFLAGS} -o $(SRC)/bin/$(BINARY_NAME) $(SRC)/cmd

test-unit:
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./...

test-unit-cover:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -cover ./...

test-examples-tradeapi-bitmex:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -tags examples ./pkg/tradeapi/bitmex

build-image:
	docker build -t tccbot -f Dockerfile .

lint:
	docker run --rm -it \
		-v $(shell go env GOCACHE):/cache/go \
		-e GOCACHE=/cache/go \
		-e GOLANGCI_LINT_CACHE=/cache/go \
		-v $(shell go env GOPATH)/pkg:/go/pkg \
		-w /app \
		-v $(shell pwd):/app \
		golangci/golangci-lint:v1.27.0-alpine golangci-lint run --config .golangci.yaml