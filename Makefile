# PREAMBLE
#//////////////////////////////////////////////////////////////////////////////
#
SHELL := bash

# VARIABLES, CONFIG, & SETTINGS
#//////////////////////////////////////////////////////////////////////////////
#
BINDIR       := $(CURDIR)/bin
BINNAME      ?= lula
INSTALL_PATH ?= /usr/local/bin

# Git vars
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
CLI_VERSION ?= $(if $(shell git describe --tags),$(shell git describe --tags),"unset")

# Go CLI options
PKG         := ./...
UNIT_PKG    := $(shell go list ./... | grep -v 'e2e')
TAGS        :=
TESTS       := .
TESTFLAGS   ?= -race -v
LDFLAGS     := -w -s -X 'github.com/mike-winberry/lulalib/src/config.CLIVersion=$(CLI_VERSION)'
GOFLAGS     ?=
CGO_ENABLED ?= 0
FUZZTIME := 10s

# Allows us to set VERSION from the command line.
# Otherwise, if BINARY_VERSION is not set, use the current git tag.
ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

# Project sources.
SRC := $(shell find . -type f -name '*.go' -print) go.mod go.sum

# TASKS
#//////////////////////////////////////////////////////////////////////////////
#
.PHONY: all
all: build test

.PHONY: help
help: ## Show this help message.
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "; printf "\nUsage:\n"}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo

.PHONY: clean
clean: ## Remove generated artifacts.
	go clean
	rm -rf $(BINDIR)

.PHONY: build
build: $(BINDIR)/$(BINNAME) ## Build the project.

$(BINDIR)/$(BINNAME): $(SRC)
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)/$(BINNAME)' .

.PHONY: test
test: 
	go clean -testcache && go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: test-unit
test-unit: # Run tests excluding those in the e2e folder.
	go clean -testcache && go test $(GOFLAGS) -run $(TESTS) $(UNIT_PKG) $(TESTFLAGS)

.PHONY: test-e2e
test-e2e: 
	cd src/test/e2e && go clean -testcache && go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: test-cmd
test-cmd: 
	cd src/cmd && go clean -testcache && go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: test-fuzz
test-fuzz:
	cd src && $(SHELL) ../build/scripts/fuzz.sh $(FUZZTIME)

.PHONY: install
install: ## Install binary to $INSTALL_PATH.
	@install "$(BINDIR)/$(BINNAME)" "$(INSTALL_PATH)/$(BINNAME)"
