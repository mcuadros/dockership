# Package configuration
PROJECT = dockership
COMMANDS = dockership dockershipd
DEPENDENCIES = \
github.com/jteeuwen/go-bindata/... \

# Environment
BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/build
VERSION ?= $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')
BUILD ?= $(shell date)
ASSETS := static

# PACKAGES
PKG_OS = darwin linux
PKG_ARCH = amd64
PACKAGES = $(foreach os, $(PKG_OS), $(foreach arch, $(PKG_ARCH), $(os)_$(arch)))
PKG_CONTENT = README.md LICENSE

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get
GOTEST = $(GOCMD) test
BINDATA = go-bindata

.PHONY: dependencies $(DEPENDENCIES) packages $(PACKAGES)

all: test build

build-assets:
	cd $(BASE_PATH)/client; npm install; npm install gulp -g; gulp javascript

assets:
	cd $(BASE_PATH)/http; $(BINDATA) -pkg=http $(ASSETS)

dependencies: $(DEPENDENCIES)
	$(GOGET) -d -v -t ./...

$(DEPENDENCIES):
	$(GOGET) $@

build: dependencies build-assets assets  $(COMMANDS)

$(COMMANDS): %: %.go
	$(GOCMD) build -ldflags "-X main.version $(VERSION) -X main.build \"$(BUILD)\"" $@.go

full-test: dependencies
	cd $(BASE_PATH)/http; $(BINDATA) -pkg=http --debug $(ASSETS)
	cd $(BASE_PATH)/core; $(GOTEST) -v . --github --slow
	cd $(BASE_PATH)/config; $(GOTEST) -v .

test: dependencies
	cd $(BASE_PATH)/http; $(BINDATA) -pkg=http --debug $(ASSETS)
	cd $(BASE_PATH)/core; $(GOTEST) -v .
	cd $(BASE_PATH)/config; $(GOTEST) -v .

install: $(COMMANDS)
	cp -rf $^ /usr/bin/

packages: clean dependencies build-assets assets $(PACKAGES)

$(PACKAGES):
	cd $(BASE_PATH)
	mkdir -p $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$@
	for cmd in $(COMMANDS); do \
		GOOS=`echo $@ | sed 's/_.*//'` \
		GOARCH=`echo $@ | sed 's/.*_//'` \
		$(GOCMD) build -ldflags "-X main.version $(VERSION) -X main.build \"$(BUILD)\"" -o $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$@/$${cmd} $${cmd}.go ; \
	done
	cp -rf $(PKG_CONTENT) $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$@/
	cd  $(BUILD_PATH) && tar -cvzf $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$@.tar.gz $(PROJECT)_$(VERSION)_$@/

clean:
	echo $(VERSION)
	rm -rf $(BUILD_PATH)

	$(GOCLEAN) .
