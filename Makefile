# Package configuration
PROJECT = dockership
COMMANDS = dockership dockershipd
DEPENDENCIES = gopkg.in/check.v1 \
code.google.com/p/go.tools/cmd/cover \
github.com/jteeuwen/go-bindata/... \
github.com/gorilla/mux \
github.com/laher/goxc

# Environment
BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/build
VERSION ?= $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')
BUILD ?= $(shell date)
ASSETS := static

# PACKAGES
PKG_OS = darwin linux
PKG_ARCH = amd64
PKG_CONTENT = README.md LICENSE

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
	for cmd in $(COMMANDS); do \
		$(GOCMD) build $${cmd}.go; \
	done

test: dependencies
	cd $(BASE_PATH)/http; $(BINDATA) -pkg=http --debug $(ASSETS)
	cd $(BASE_PATH)/core; $(GOTEST) -v . --github --slow
	cd $(BASE_PATH)/config; $(GOTEST) -v .

dependencies:
	$(GOGET) -d -v ./...
	for i in $(DEPENDENCIES); do $(GOGET) $$i; done

packages: clean assets
	for os in $(PKG_OS); do \
		for arch in $(PKG_ARCH); do \
			cd $(BASE_PATH); \
			mkdir -p $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}; \
			for cmd in $(COMMANDS); do \
				GOOS=$${os} GOARCH=$${arch} $(GOCMD) build -ldflags "-X main.version $(VERSION) -X main.build \"$(BUILD)\"" -o $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}/$${cmd} $${cmd}.go ; \
			done; \
			for content in $(PKG_CONTENT); do \
				cp -rf $${content} $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}/; \
			done; \
			cd  $(BUILD_PATH) && tar -cvzf $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}.tar.gz $(PROJECT)_$(VERSION)_$${os}_$${arch}/; \
		done; \
	done;


clean:
	echo $(VERSION)
	rm -rf $(BUILD_PATH)

	$(GOCLEAN) .
