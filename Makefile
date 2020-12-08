SHELL=/bin/bash -o pipefail

export GO111MODULE := on
export PATH := .bin:${PATH}

SRCROOT ?= $(realpath .)

# These are paths used in the docker image
SRCROOT_D = /go/src/github.com/ory/oathkeeper

BUILD_NUMBER ?= 0
BUILD_IDENTIFIER = _${BUILD_NUMBER}

MAJOR_VERSION = 1
MINOR_VERSION = 0
PATCH_VERSION = $(BUILD_NUMBER)

REVISION ?= $$(git rev-parse --short HEAD)
VERSION ?= $(MAJOR_VERSION).$(MINOR_VERSION).$(PATCH_VERSION)
CGO_ENABLED ?= 0

LD_FLAGS ?= -ldflags "-X github.com/ory/oathkeeper/x.Commit=$(REVISION) -X github.com/ory/oathkeeper/x.Version=$(VERSION)"

.PHONY: deps
deps:
ifneq ("$(shell base64 Makefile))","$(shell cat .bin/.lock)")
		go build -o .bin/go-acc github.com/ory/go-acc
		go build -o .bin/goreturns github.com/sqs/goreturns
		go build -o .bin/listx github.com/ory/x/tools/listx
		go build -o .bin/mockgen github.com/golang/mock/mockgen
		go build -o .bin/swagger github.com/go-swagger/go-swagger/cmd/swagger
		go build -o .bin/goimports golang.org/x/tools/cmd/goimports
		go build -o .bin/ory github.com/ory/cli
		go build -o .bin/packr2 github.com/gobuffalo/packr/v2/packr2
		go build -o .bin/go-bindata github.com/go-bindata/go-bindata/go-bindata
		echo "v0" > .bin/.lock
		echo "$$(base64 Makefile)" > .bin/.lock
endif

# Formats the code
.PHONY: format
format: deps
		goreturns -w -local github.com/ory $$(listx .)

.PHONY: gen
gen:
		mocks sdk

# Generates the SDKs
.PHONY: sdk
sdk: deps
		swagger generate spec -m -o ./.schema/api.swagger.json -x internal/httpclient
		ory dev swagger sanitize ./.schema/api.swagger.json
		swagger flatten --with-flatten=remove-unused -o ./.schema/api.swagger.json ./.schema/api.swagger.json
		swagger validate ./.schema/api.swagger.json
		rm -rf internal/httpclient
		mkdir -p internal/httpclient
		swagger generate client -f ./.schema/api.swagger.json -t internal/httpclient -A Ory_Oathkeeper
		make format

.PHONY: install-stable
install-stable: deps
		OATHKEEPER_LATEST=$$(git describe --abbrev=0 --tags)
		git checkout $$OATHKEEPER_LATEST
		packr2
		GO111MODULE=on go install \
				-ldflags "-X github.com/ory/oathkeeper/x.Version=$$OATHKEEPER_LATEST -X github.com/ory/oathkeeper/x.Date=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/oathkeeper/x.Commit=`git rev-parse HEAD`" \
				.
		packr2 clean
		git checkout master

.PHONY: install
install: deps
		packr2 || (GO111MODULE=on go install github.com/gobuffalo/packr/v2/packr2 && packr2)
		GO111MODULE=on go install .
		packr2 clean

.PHONY: docker
docker: deps
		packr2 || (GO111MODULE=on go install github.com/gobuffalo/packr/v2/packr2 && packr2)
		CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build
		packr2 clean
		docker build -t oryd/oathkeeper:dev .
		docker build -t oryd/oathkeeper:dev-alpine -f Dockerfile-alpine .
		rm oathkeeper

.PHONY: clean
clean:
		rm -f oathkeeper

.PHONY: build
build: clean
		which packr2; if [ $$? -eq 1 ]; then \
				go get -u github.com/gobuffalo/packr/v2/packr2; \
		fi
		packr2 clean
		packr2
	 	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) go build $(LD_FLAGS)

.PHONY: docker.build
docker.build: clean
		docker build -f Dockerfile-builder \
			-t oathkeeper$(BUILD_IDENTIFIER) \
			--build-arg VERSION=$(VERSION) \
			--build-arg REVISION=$(REVISION) \
			--build-arg CGO_ENABLED=$(CGO_ENABLED) \
			--build-arg BUILDER_IMAGE=$(BUILDER_IMAGE) .
		docker create -it --name tocopy-oathkeeper$(BUILD_IDENTIFIER) oathkeeper$(BUILD_IDENTIFIER) bash
		docker cp tocopy-oathkeeper$(BUILD_IDENTIFIER):$(SRCROOT_D)/oathkeeper $(SRCROOT)/
		docker rm -f tocopy-oathkeeper$(BUILD_IDENTIFIER)
		docker rmi -f oathkeeper$(BUILD_IDENTIFIER)
