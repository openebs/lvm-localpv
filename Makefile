# list only csi source code directories
PACKAGES = $(shell go list ./... | grep -v 'pkg/generated')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

GOLANGCI_LINT_VERSION = 2.5.0

# Tools required for different make
# targets or for development purposes
EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/cover \
	golang.org/x/lint/golint \
	github.com/axw/gocov/gocov@v1.1 \
	github.com/matm/gocov-html/cmd/gocov-html \
	github.com/onsi/ginkgo/ginkgo@v1.16.5 \
	github.com/onsi/gomega/...@v1.35

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in
# the `build/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique
# repository under an organization.
# Putting all this together, an unique uri for a given
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization
# under which images should be pushed.
# By default the organization name is `openebs`.

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG="openebs"
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/lvm-localpv"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

# Set the path to the Chart.yaml file
ROOT_DIR:=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))
CHART_YAML:=${ROOT_DIR}/deploy/helm/charts/Chart.yaml

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG := $(shell awk -F': ' '/^version:/ {print $$2}' $(CHART_YAML))
  export IMAGE_TAG
endif

# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS
ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH
ARCH:=${XC_OS}_${XC_ARCH}
export ARCH

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg BRANCH=${BRANCH} --build-arg RELEASE_TAG=${RELEASE_TAG}

# Specify the name for the binary
CSI_DRIVER=lvm-driver

GEN_SRC=openebs.io/lvm/v1alpha1 

.PHONY: all
all: golint test manifests lvm-driver-image fio-image

.PHONY: clean
clean:
	@echo "--> Cleaning Directory" ;
	go clean -testcache
	rm -rf bin
	./ci/ci-test.sh clean
	chmod -R u+w ${GOPATH}/bin/${CSI_DRIVER} 2>/dev/null || true
	chmod -R u+w ${GOPATH}/pkg/* 2>/dev/null || true
	rm -rf ${GOPATH}/bin/${CSI_DRIVER}
	rm -rf ${GOPATH}/pkg/*

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

.PHONY: test
test: format
	@echo "--> Running go test" ;
	@./buildscripts/test-cov.sh


.PHONY: deps
deps:
	@echo "--> Tidying up submodules"
	@go mod tidy
	@echo "--> Verifying submodules"
	@go mod verify

.PHONY: verify-deps
verify-deps: deps
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please commit the changes to go.mod and go.sum"; exit 1; \
	fi

# Bootstrap downloads tools required
# during build
.PHONY: bootstrap
bootstrap: controller-gen install-golangci-lint
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		if ! echo $$tool | grep "@"; then \
			tool=$$tool@latest ; \
		fi ; \
		GO111MODULE=on go install $$tool; \
	done

## golangci-lint tool used to check linting tools in codebase
## Example: golangci-lint document is not recommending
##			to use `go get <path>`. For more info:
##          https://golangci-lint.run/usage/install/#install-from-source
##
## Install golangci-lint only if tool doesn't exist in system
.PHONY: install-golangci-lint
install-golangci-lint:
	$(if $(shell which golangci-lint), echo "golangci-lint already exist in system", (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOPATH}/bin" v$(GOLANGCI_LINT_VERSION)))

.PHONY: controller-gen
controller-gen:
	@go install -mod=mod sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0

# SRC_PKG is the path of code files
SRC_PKG := github.com/openebs/lvm-localpv/pkg

# code generation for custom resources
.PHONY: kubegen
kubegen: kubegendelete deepcopy-install clientset-install lister-install informer-install
	make deepcopy clientset lister informer

# deletes generated code by codegen
.PHONY: kubegendelete
kubegendelete:
	@rm -rf pkg/generated/clientset
	@rm -rf pkg/generated/lister
	@rm -rf pkg/generated/informer

.PHONY: deepcopy-install
deepcopy-install:
	@go install k8s.io/code-generator/cmd/deepcopy-gen 

.PHONY: deepcopy
deepcopy:
	@echo "+ Generating deepcopy funcs for $(GEN_SRC)"
	@deepcopy-gen \
		--output-file zz_generated.deepcopy.go \
 		--go-header-file ./buildscripts/custom-boilerplate.go.txt \
		$(SRC_PKG)/apis/$(GEN_SRC)

.PHONY: clientset-install
clientset-install:
	@go install k8s.io/code-generator/cmd/client-gen

.PHONY: clientset
clientset:
	@echo "+ Generating clientsets for $(GEN_SRC)"
	@client-gen \
		--fake-clientset \
		--input-base $(SRC_PKG)/apis \
		--output-dir pkg/generated/clientset \
		--output-pkg $(SRC_PKG)/generated/clientset \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt \
		--input $(GEN_SRC)

.PHONY: lister-install
lister-install:
	@go install k8s.io/code-generator/cmd/lister-gen

.PHONY: lister
lister:
	@echo "+ Generating lister for $(GEN_SRC)"
	@lister-gen \
		--output-dir pkg/generated/lister \
		--output-pkg $(SRC_PKG)/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt \
		$(SRC_PKG)/apis/$(GEN_SRC)

.PHONY: informer-install
informer-install:
	@go install k8s.io/code-generator/cmd/informer-gen

.PHONY: informer
informer:
	@echo "+ Generating informer for $(GEN_SRC)"
	@informer-gen \
		--versioned-clientset-package $(SRC_PKG)/generated/clientset/internalclientset \
		--listers-package $(SRC_PKG)/generated/lister \
		--output-dir pkg/generated/informer \
		--output-pkg $(SRC_PKG)/generated/informer \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt \
		$(SRC_PKG)/apis/$(GEN_SRC)

manifests:
	@echo "--------------------------------"
	@echo "+ Generating LVM LocalPV crds"
	@echo "--------------------------------"
	./buildscripts/generate-manifests.sh

.PHONY: lvm-driver
lvm-driver: format
	@echo "--------------------------------"
	@echo "--> Building ${CSI_DRIVER}        "
	@echo "--------------------------------"
	@PNAME=${CSI_DRIVER} CTLNAME=${CSI_DRIVER} sh -c "'./buildscripts/build.sh'"

.PHONY: lvm-driver-image
lvm-driver-image: lvm-driver
	@echo "--------------------------------"
	@echo "+ Generating ${CSI_DRIVER} image"
	@echo "--------------------------------"
	@cp bin/${CSI_DRIVER}/${CSI_DRIVER} buildscripts/${CSI_DRIVER}/
	cd buildscripts/${CSI_DRIVER} && docker build -t ${IMAGE_ORG}/${CSI_DRIVER}:${IMAGE_TAG} ${DBUILD_ARGS} . && docker tag ${IMAGE_ORG}/${CSI_DRIVER}:${IMAGE_TAG} quay.io/${IMAGE_ORG}/${CSI_DRIVER}:${IMAGE_TAG}
	@rm buildscripts/${CSI_DRIVER}/${CSI_DRIVER}

.PHONY: fio-image
fio-image:
	@echo "--------------------------------"
	@echo "+ Generating fio image"
	@echo "--------------------------------"
	cd buildscripts/fio && docker build -t ${IMAGE_ORG}/fio:latest ${DBUILD_ARGS} .

.PHONY: image-tag
image-tag:
	@echo ${IMAGE_TAG}

.PHONY: image-repo
image-repo:
	@echo ${IMAGE_ORG}/${CSI_DRIVER}

.PHONY: fio-image-repo
fio-image-repo:
	@echo ${IMAGE_ORG}/fio

.PHONY: image-ref
image-ref:
	@echo docker.io/${IMAGE_ORG}/${CSI_DRIVER}:${IMAGE_TAG}

.PHONY: ci
ci:
	@echo "--> Running ci test";
	TEST_ARGS=""; \
	if [ "${PROVIDER}" = minikube ]; then \
		TEST_ARGS="${TEST_ARGS} --provider minikube"; \
	elif [ -n "${DATADIR}" ]; then \
		TEST_ARGS="${TEST_ARGS} --datadir ${DATADIR}"; \
	fi; \
	./ci/ci-test.sh$$TEST_ARGS run

.PHONY: minikube-ci
minikube-ci:
	@echo "--> Running ci test";
	./ci/ci-test.sh --provider minikube --reset --datadir .tmp run

# Push lvm-driver images
deploy-images:
	@DIMAGE="${IMAGE_ORG}/lvm-driver" ./buildscripts/push

# Push lvm-localpv-e2e-tests images
deploy-e2e-images:
	@DIMAGE="${IMAGE_ORG}/lvm-localpv-e2e" ./buildscripts/push

## Currently we are running with Default options + other options
## Explanation for explicitly mentioned linters:
## dupl: Tool for code clone detection within repo
## revive: Drop-in replacement of golint. It allows to enable or disable
##         rules using configuration file.
## bodyclose: checks whether HTTP response body is closed successfully
## goconst: Find repeated strings that could be replaced by a constant
## misspell: Finds commonly misspelled English words in comments
.PHONY: golint
golint:
	@echo "--> Running golint"
	CGO_ENABLED=0 golangci-lint run -E dupl,revive,bodyclose,goconst,misspell --timeout 5m0s
	@echo "Completed golangci-lint no recommendations !!"
	@echo "--------------------------------"
	@echo ""

include Makefile.buildx.mk
