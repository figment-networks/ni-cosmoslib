
LDFLAGS      := -w -s
MODULE       := github.com/figment-networks/ni-cosmoslib
VERSION_FILE ?= ./VERSION


# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

ifneq (,$(wildcard $(VERSION_FILE)))
VERSION ?= $(shell head -n 1 $(VERSION_FILE))
else
VERSION ?= n/a
endif

all: test


.PHONY: test
test:
	cd api && go test ./... -cover -race && cd ..


