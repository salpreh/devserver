.PHONY: build test

build:
	@go build -o build/devserver ./cmd

run:
	@go run ./cmd

test:
	@go test -v ./test/...

install: build _install

_install:
ifeq ($(GOBIN),)
	@echo "GOBIN is not set"
else
	@echo "Installing devserver in $(GOBIN)"
	@cp ./build/devserver $(GOBIN)
endif

uninstall:
ifeq ($(GOBIN),)
	@echo "GOBIN is not set"
else
	@echo "Uninstalling devserver from $(GOBIN)"
	@rm $(GOBIN)/devserver
endif