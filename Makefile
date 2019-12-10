# Project information
OWNER      := tomohiro
REPOSITORY := $(shell basename $(PWD))
VERSION    := $(shell grep "const Version " $(PWD)/version.go | sed -E 's/.*"(.+)"$$/\1/')
REVISION   := $(shell git rev-parse --short HEAD)

# Build information
DIST_DIR      := $(PWD)/dist
ASSETS_DIR    := $(DIST_DIR)/$(VERSION)
XC_OS         := "linux darwin windows"
XC_ARCH       := "386 amd64"
BUILD_LDFLAGS := "-w -s -X main.revision=$(REVISION)"

.PHONY: help
help:
	@echo "Please type: make [target]"
	@echo "  setup        Setup development environment"
	@echo "  deps         Install runtime dependencies"
	@echo "  updatedeps   Update runtime dependencies"
	@echo "  dist         Ship packages as release assets"
	@echo "  release      Publish release assets to GitHub"
	@echo "  clean        Cleanup assets"
	@echo "  help         Show this help messages"

.PHONY: setup
setup:
	@echo "===> Setup development tools..."
	# Install goxz - Just do cross building and archiving go tools conventionally
	GO111MODULE=off go get github.com/Songmu/goxz/cmd/goxz
	# Install ghr - Upload multiple artifacts to GitHub Release in parallel
	GO111MODULE=off go get github.com/tcnksm/ghr

.PHONY: deps
deps:
	@echo "===> Installing runtime dependencies..."
	go mod download

.PHONY: updatedeps
updatedeps:
	@echo "===> Updating runtime dependencies..."
	go get -u

.PHONY: dist
dist:
	@echo "===> Shipping packages as release assets..."
	goxz -d=$(ASSETS_DIR) -os=$(XC_OS) -arch=$(XC_ARCH) --build-ldflags=$(BUILD_LDFLAGS) -pv=$(VERSION) -z
	pushd $(ASSETS_DIR); \
	shasum -a 256 *.zip > ./$(VERSION)_SHA256SUMS; \
	popd

.PHONY: release
release:
	@echo "===> Publishing to GitHub..."
	ghr -u=$(OWNER) -r=$(REPOSITORY) $(VERSION) $(ASSETS_DIR)

.PHONY: clean
clean:
	@echo "===> Cleaning assets..."
	go clean ./...
	rm -rf $(DIST_DIR)
