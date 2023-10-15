GOCMD=go
GOTEST=$(GOCMD) test -v -count 1
GOVET=$(GOCMD) vet
GOBUILD=$(GOCMD) build
GOFMT=gofmt

PACKAGES = ./checker ./badge
CMDS = ./cmd/checker ./cmd/badge
TARGETS = $(PACKAGES) $(CMDS)

PREFLAGS += GOOS=linux GOARCH=amd64
LDFLAGS = "-s -w"
override CC := /usr/bin/gcc

all:
	@echo "No default operation"

checker: bin Makefile
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

badge: bin Makefile
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

cmd: Makefile
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/cmd/checker ./cmd/checker
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/cmd/badge ./cmd/badge

fmt:
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

lint:
	bash ./scripts/lint-check

vet:
	$(GOVET) $(TARGETS)

test:
	$(GOTEST) $(PACKAGES)

bin:
	mkdir -p $@

clean: bin
	rm -rf ./bin/*

.PHONY: fmt vet bin test clean all checker badge cmd