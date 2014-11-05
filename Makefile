# Package configuration
DEPENDENCIES = gopkg.in/check.v1 \
code.google.com/p/go.tools/cmd/cover \
github.com/jteeuwen/go-bindata/... \
github.com/gorilla/mux \
github.com/laher/goxc

# Environment
BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/build
VERSION ?= $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')
ASSETS := static

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get
GOTEST = $(GOCMD) test
GOXC = goxc
GOXC_CONFIG = .goxc.json
BINDATA = go-bindata

# Rules
all: test build

assets:
	cd $(BASE_PATH)/http; $(BINDATA) -pkg=http $(ASSETS)

build: assets dependencies
	$(GOCMD) build dockership.go
	$(GOCMD) build dockershipd.go

test: dependencies
	cd $(BASE_PATH)/http; $(BINDATA) --debug $(ASSETS)
	cd $(BASE_PATH)/core; $(GOTEST) -v . --github --slow
	cd $(BASE_PATH)/config; $(GOTEST) -v .

dependencies:
	$(GOGET) -d -v ./...
	for i in $(DEPENDENCIES); do $(GOGET) $$i; done

packages: assets
	$(GOXC) -d="$(BUILD_PATH)" -c $(GOXC_CONFIG) -pv="$(VERSION)"

clean:
	echo $(VERSION)
	rm -rf $(BUILD_PATH)

	$(GOCLEAN) .
