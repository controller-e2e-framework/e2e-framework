##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Prepare

.PHONY: prepare
prepare: ## Prepare the running environment by installing the required tooling
	./hack/scripts/install_tools.sh

##@ House Keeping

.PHONY: tidy
tidy:  ## Run go mod tidy
	rm -f go.sum; go mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Tests

.PHONY: test
test: ## Run all of the tests
	go test ./...

