GOCMD=go
GOBUILD=$(GOCMD) build
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
	@GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -v ${LDFLAGS} -o $(SRC)/$(BINARY_NAME) $(SRC)/cmd

test-unit:
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./...

test-unit-cover:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -cover ./...

test-examples-tradeapi-bitmex:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -tags examples ./pkg/tradeapi/bitmex

build-image:
	docker build -t tccbot -f Dockerfile .
