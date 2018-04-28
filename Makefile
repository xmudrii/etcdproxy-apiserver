ifndef VERBOSE
	MAKEFLAGS += --silent
endif

CI_PKGS=$(shell go list ./... | grep -v /vendor)
FMT_PKGS=$(shell go list -f {{.Dir}} ./... | grep -v vendor | tail -n +2)

default: authorsfile compile # Create the etcdproxy executable in the ./bin directory and regenerate the AUTHORS file.

all: default install # Create the etcdproxy executable and move it to the $GOPTAH/bin.

compile: ## Create the etcdproxy executable in the ./bin directory.
	go build -o bin/etcdproxy main.go

install: ## Create the etcdproxy executable in $GOPATH/bin directory.
	install -m 0755 bin/etcdproxy ${GOPATH}/bin/etcdproxy

authorsfile: ## Update the AUTHORS file from the git logs.
	git log --all --format='%aN <%cE>' | sort -u > AUTHORS

clean: ## Clean the project tree from binary files.
	rm -rf bin/*

gofmt: install-tools ## gofmt your code.
	echo "Fixing format of go files..."; \
	for package in $(FMT_PKGS); \
	do \
		gofmt -w $$package ; \
		goimports -l -w $$package ; \
	done

lint: install-tools ## Check for style mistakes all Go files using golint,
	golint $(PKGS)

vet: ## Apply go vet to all the Go files,
	@go vet $(PKGS)

check-headers: ## Check if the headers are valid.
	./hack/check-headers.sh

update-headers: ## Update the headers in the repository. Required for all new files.
	./hack/update-headers.sh

.PHONY: test-ci
test-ci: ## Run the CI tests.
	go test -timeout 20m -v $(CI_PKGS)

.PHONY: check-ci
check-ci: install-tools ## Run code checks
	PKGS="${FMT_PKGS}" GOFMT="gofmt" GOLINT="golint" ./hack/check-ci.sh

.PHONY: install-tools
install-tools:
	GOIMPORTS_CMD=$(shell command -v goimports 2> /dev/null)
ifndef GOIMPORTS_CMD
	go get golang.org/x/tools/cmd/goimports
endif
	GOLINT_CMD=$(shell command -v golint 2> /dev/null)
ifndef GOLINT_CMD
	go get github.com/golang/lint/golint
endif

.PHONY: help
help:  ## Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'
