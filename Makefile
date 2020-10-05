VERSION     := $(shell git describe --tags)
TIME        := $(shell date +%FT%T%z)

NAME        := speechly-slu
BUILDNAME   ?= ./bin/$(NAME)
INSTALLNAME := ${GOBIN}/$(NAME)
LINTNAME    := .linted

CMD         := ./cmd
SOURCES     := $(shell find . -name '*.go')
GOPKGS      := $(shell go list ./... | grep -v "/test")
GOTEST      := -race -cover -covermode=atomic -v $(GOPKGS)
GOFLAGS     := "-X github.com/speechly/slu-client/internal/application.BuildVersion=$(VERSION) -X github.com/speechly/slu-client/internal/application.BuildTime=$(TIME)"

all: vendor lint test build
.PHONY: all

clean:
	rm -rf $(BUILDNAME) $(LINTNAME)
.PHONY: clean

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

$(LINTNAME): $(SOURCES) .golangci.yml
	golangci-lint run --exclude-use-default=false
	@touch $@

vendor: $(SOURCES)
	go mod tidy
	go mod vendor
	@touch $@
