## toolbox - start
## Generated with https://github.com/bakito/toolbox

## Current working directory
TB_LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
TB_LOCALBIN ?= $(TB_LOCALDIR)/bin
$(TB_LOCALBIN):
	if [ ! -e $(TB_LOCALBIN) ]; then mkdir -p $(TB_LOCALBIN); fi

## Tool Binaries
TB_GOLANGCI_LINT ?= $(TB_LOCALBIN)/golangci-lint

## Tool Versions
TB_GOLANGCI_LINT_VERSION ?= v2.11.2

## Tool Installer
.PHONY: tb.golangci-lint
tb.golangci-lint: ## Download golangci-lint locally if necessary.
	@test -s $(TB_GOLANGCI_LINT) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(TB_GOLANGCI_LINT_VERSION)

## Reset Tools
.PHONY: tb.reset
tb.reset:
	@rm -f \
		$(TB_GOLANGCI_LINT)

## Update Tools
.PHONY: tb.update
tb.update: tb.reset
	toolbox makefile -f $(TB_LOCALDIR)/Makefile \
		github.com/golangci/golangci-lint/v2/cmd/golangci-lint
## toolbox - end
