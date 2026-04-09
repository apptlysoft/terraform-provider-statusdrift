BINARY=terraform-provider-statusdrift
VERSION=$(shell cat VERSION)
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_DIR=~/.terraform.d/plugins/registry.terraform.io/apptlysoft/statusdrift/$(VERSION)/$(OS_ARCH)
DIST_DIR=dist

PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	linux/386 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/386 \
	windows/arm64 \
	freebsd/amd64 \
	freebsd/386 \
	freebsd/arm

.PHONY: build install test lint clean docs generate build-all

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)

build-all: clean
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=$(DIST_DIR)/$(BINARY)_$(VERSION)_$${os}_$${arch}; \
		if [ "$$os" = "windows" ]; then output=$${output}.exe; fi; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "-X main.version=$(VERSION)" -o $$output || exit 1; \
	done
	@echo "All binaries built in $(DIST_DIR)/"

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)

generate:
	go generate ./...

docs: generate
