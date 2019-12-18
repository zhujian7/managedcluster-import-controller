# Copyright 2019 The Kubernetes Authors.
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

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overridding the IMG and REGISTRY environment variable.
IMG ?= multicloud-operators-cluster-controller
REGISTRY ?= quay.io/rhibmcollab

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/rh-ibm-synergy

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
	git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

# enable GO MOD
export GO111MODULE := on

ifeq (,$(wildcard go.mod))
ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
    $(error Please run 'make' from $(DEST). Current directory is $(PWD))
endif
endif

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_GIT_TARGET = config-git
endif

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
endif

config-git:
	@mkdir /root/.ssh && cat /home/ssh/ssh-key > /root/.ssh/id_rsa && chmod 600 /root/.ssh/id_rsa && ssh-keyscan -H github.com >> /root/.ssh/known_hosts
	@git config --global url.git@github.com:rh-ibm-synergy/.insteadOf https://github.com/rh-ibm-synergy/

############################################################
# all section
############################################################

all: fmt check test coverage build images

############################################################
# install git hooks
############################################################
INSTALL_HOOKS := $(shell find .git/hooks -type l -exec rm {} \; && \
                         find common/scripts/.githooks -type f -exec ln -sf ../../{} .git/hooks/ \; )

include common/Makefile.common.mk

############################################################
# work section
############################################################

$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

############################################################
# format section
############################################################

# All available format: format-go format-protos format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go format-protos format-python

############################################################
# check section
############################################################

check: lint

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: $(CONFIG_GIT_TARGET) lint-all

############################################################
# test section
############################################################

test: $(CONFIG_GIT_TARGET)
	@go test ${TESTARGS} ./...

############################################################
# coverage section
############################################################

coverage:
	@common/scripts/codecov.sh

############################################################
# install operator sdk section
############################################################

install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

############################################################
# build section
############################################################

build: install-operator-sdk $(CONFIG_GIT_TARGET)
	operator-sdk build $(REGISTRY)/$(IMG):$(VERSION)

############################################################
# run section
############################################################

export ENDPOINT_CRD_FILE=$(PWD)/build/resources/multicloud_v1beta1_endpoint_crd.yaml

run:
	@operator-sdk up local --namespace="" --operator-flags="--zap-devel=true"

############################################################
# images section
############################################################

images: build build-push-images

build-push-images: $(CONFIG_DOCKER_TARGET)
	docker tag $(REGISTRY)/$(IMG):$(VERSION) $(REGISTRY)/$(IMG)
	@if [ $(BUILD_LOCALLY) -ne 1 ]; then docker push $(REGISTRY)/$(IMG):$(VERSION); docker push $(REGISTRY)/$(IMG); fi

############################################################
# deploy section
############################################################

deploy:
	mkdir -p overlays/deploy
	cp overlays/template/kustomization.yaml overlays/deploy
	cd overlays/deploy
	kustomize build overlays/deploy | kubectl apply -f -
	rm -rf overlays/deploy

############################################################
# dev section
############################################################

dev-build: build

dev-images: build dev-build-push-images

dev-build-push-images:
	docker push $(REGISTRY)/$(IMG):$(VERSION)

dev-deploy: dev-images
	mkdir -p overlays/deploy
	cp overlays/template/kustomization.yaml overlays/deploy
	cd overlays/deploy && kustomize edit set image quay.io/rhibmcollab/multicloud-operators-cluster-controller:latest=$(REGISTRY)/$(IMG):$(VERSION)
	kustomize build overlays/deploy | kubectl apply -f -
	rm -rf overlays/deploy

############################################################
# regcred section
############################################################

regcred-patch-sa:
	kubectl patch sa multicloud-operators-cluster-controller --type='json' -p='[{"op": "add", "path": "/imagePullSecrets/1", "value": {"name": "$(IMAGE_PULL_SECRET)" } }]'

############################################################
# clean section
############################################################

clean:
	rm -f build/_output/bin/$(IMG)

.PHONY: all build check lint test coverage images deploy deploy-build
