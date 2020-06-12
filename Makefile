GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GO111MODULE=auto
CGO_ENABLED=0
BINARY_NAME=tccbot-backend
PACKAGE=$(BINARY_NAME)
GOPATH=$(HOME)/go
SRC=.

build:
	@GO111MODULE=$(GO111MODULE) $(GOBUILD) -o $(SRC)/$(BINARY_NAME) -ldflags "-s -w" $(SRC)/cmd

test-unit:
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./...

test-unit-cover:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -cover ./...

test-examples-tradeapi-bitmex:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -tags examples ./pkg/tradeapi/bitmex

build-image:
	docker build -t tccbot -f Dockerfile .
