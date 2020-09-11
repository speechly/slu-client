VERSION     := $(shell git describe --tags)
AUTHOR      := $(shell git config user.email)
TIME        := $(shell date +%FT%T%z)

NAME        := speechly-slu
BUILDNAME   ?= ./bin/$(NAME)
INSTALLNAME := ${GOBIN}/$(NAME)
LINTNAME    := .linted
PROTONAME   := pkg/speechly/*.pb.go

CMD         := ./cmd
SOURCES     := $(shell find . -name '*.go')
GOPKGS      := $(shell go list ./... | grep -v "/test")
GOTEST      := -race -cover -covermode=atomic -v $(GOPKGS)
GOFLAGS     := "-X github.com/speechly/slu-client/internal/application.BuildVersion=$(VERSION) -X github.com/speechly/slu-client/internal/application.BuildTime=$(TIME) -X github.com/speechly/slu-client/internal/application.BuildAuthor=$(AUTHOR)"

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
	go build -i -o $(INSTALLNAME) -ldflags $(GOFLAGS) $(CMD)

$(BUILDNAME): $(SOURCES)
	go build -o $(BUILDNAME) -ldflags $(GOFLAGS) $(CMD)

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
