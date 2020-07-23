CMD         := ./cmd
NAME        := speechly-slu
BUILDNAME   ?= ./bin/$(NAME)
INSTALLNAME := ${GOBIN}/$(NAME)
LINTNAME    := .linted
PROTONAME   := pkg/speechly/*.pb.go
SOURCES     := $(shell find . -name '*.go')
GOPKGS      := $(shell go list ./... | grep -v "/test")
GOTEST      := -race -cover -covermode=atomic -v $(GOPKGS)

all: vendor proto lint test build
.PHONY: all

clean:
	rm -rf $(BUILDNAME) $(LINTNAME) $(PROTONAME)
.PHONY: clean

proto: $(PROTONAME)
build: $(BUILDNAME)
install: $(INSTALLNAME)
lint: $(LINTNAME)

test:
	go test $(GOTEST)
.PHONY: test

$(INSTALLNAME): $(SOURCES)
	go build -i -o $(INSTALLNAME) $(CMD)

$(BUILDNAME): $(SOURCES)
	go build -o $(BUILDNAME) $(CMD)

$(PROTONAME): api/speechly/*.proto
	protoc \
	-I api/speechly/ \
	-I vendor/ \
	--gogofaster_out=plugins=grpc:pkg/speechly \
	$^

$(LINTNAME): $(SOURCES) .golangci.yml
	golangci-lint run --exclude-use-default=false
	@touch $@

vendor: $(SOURCES)
	go mod tidy
	go mod vendor
	@touch $@
