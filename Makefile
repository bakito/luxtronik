# Include toolbox tasks
include ./.toolbox.mk

lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

test:
	go test -v ./...