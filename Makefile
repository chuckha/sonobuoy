# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

TARGET = sonobuoy
E2ETARGET = e2e.test
GOTARGET = github.com/heptio/$(TARGET)
E2ESRCS = $(GOTARGET)/pkg/battery
BUILDMNT = /go/src/$(GOTARGET)
REGISTRY ?= gcr.io/heptio-images
VERSION ?= v0.1
TESTARGS ?= -v
IMAGE = $(REGISTRY)/$(BIN)
BUILD_IMAGE ?= golang:1.7
TEST_PKGS ?= ./pkg/agent/... ./pkg/aggregator/...
# NOTE - the only reason we don't choose alpine is it's missing gcc deps
# properly compile and we'd need to update. https://github.com/docker-library/golang/issues/153
# BUILD_IMAGE ?= golang:1.7-alpine3.5
DOCKER ?= docker
DIR := ${CURDIR}
BUILD = go build -v -ldflags "-X github.com/heptio/sonobuoy/pkg/buildinfo.Version=$(VERSION) -X github.com/heptio/sonobuoy/pkg/buildinfo.DockerImage=$(REGISTRY)/$(TARGET)" && go test -i -c -o $(E2ETARGET) $(E2ESRCS)
TEST = go test $(TEST_PKGS) $(TESTARGS)

local:
	$(BUILD)

test:
	$(TEST)

all: cbuild container

cbuild:
	$(DOCKER) run --rm -v $(DIR):$(BUILDMNT) -w $(BUILDMNT) $(BUILD_IMAGE) /bin/sh -c '$(BUILD)'

container: cbuild
	$(DOCKER) build -t $(REGISTRY)/$(TARGET):latest -t $(REGISTRY)/$(TARGET):$(VERSION) .

# TODO: Determine tagging mechanics
push:
	docker -- push $(REGISTRY)/$(TARGET)

.PHONY: all local container cbuild push test

clean:
	rm -f $(TARGET) $(TESTTARGET)
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):$(VERSION)
