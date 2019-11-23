# Project information
OWNER      = "tomohiro"
REPOSITORY = $(shell basename $(PWD))
VERSION    = $(shell grep "const Version " $(PWD)/version.go | sed -E 's/.*"(.+)"$$/\1/')

# Build information
DIST_DIR   = $(PWD)/dist
ASSETS_DIR = $(DIST_DIR)/$(VERSION)
XC_OS      = "linux darwin windows"
XC_ARCH    = "386 amd64"

# Tasks
help:
	@echo "Please type: make [target]"
	@echo "  setup        Setup development environment"
	@echo "  deps         Install runtime dependencies"
	@echo "  updatedeps   Update runtime dependencies"
	@echo "  dist         Ship packages as release assets"
	@echo "  release      Publish release assets to GitHub"
	@echo "  clean        Cleanup assets"
	@echo "  help         Show this help messages"

setup:
	@echo "===> Setup development tools..."

	# goxz - Just do cross building and archiving go tools conventionally
	GO111MODULE=off go get github.com/Songmu/goxz/cmd/goxz

	# ghr - Upload multiple artifacts to GitHub Release in parallel
	GO111MODULE=off go get github.com/tcnksm/ghr

deps:
	@echo "===> Installing runtime dependencies..."
	go mod download

updatedeps:
	@echo "===> Updating runtime dependencies..."
	go get -u

dist:
	@echo "===> Shipping packages as release assets..."
	goxz -d $(ASSETS_DIR) -os $(XC_OS) -arch $(XC_ARCH) -pv $(VERSION) -z
	pushd $(ASSETS_DIR); \
	shasum -a 256 *.zip > ./$(VERSION)_SHA256SUMS; \
	popd; \

release:
	@echo "===> Publishing to GitHub..."
	ghr -u $(OWNER) -r $(REPOSITORY) $(VERSION) $(ASSETS_DIR)

clean:
	@echo "===> Cleaning assets..."
	go clean ./...
	rm -rf $(DIST_DIR)

.PHONY: help setup deps updatedeps dist release clean
