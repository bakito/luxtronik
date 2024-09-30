# Include toolbox tasks
include ./.toolbox.mk

lint: golangci-lint
	$(GOLANGCI_LINT) run --fix

