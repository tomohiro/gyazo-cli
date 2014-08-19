.PHONY: clean

# Project information
OWNER      = $(shell whoami)
TOP        = $(shell pwd)
REPOSITORY = $(shell basename $(TOP))
VERSION    = $(shell grep "const Version " $(TOP)/version.go | sed -E 's/.*"(.+)"$$/\1/')

# Build information
OUTPUT    = gyazo
BUILDTOOL = gox
BUILDDIR  = $(TOP)/pkg
XC_OS     = "darwin linux windows"
XC_ARCH   = "386 amd64 arm"
DISTDIR   = $(BUILDDIR)/dist/$(VERSION)


default: test

test: deps
	@echo "===> Running tests..."
	go test -v ./...

setup:
	@echo "===> Setup development tools..."

	# Gox - Simple Go Cross Compilation
	go get github.com/mitchellh/gox
	gox -build-toolchain

	# ghr - Easy to ship your project on GitHub to your user
	go get github.com/tcnksm/ghr

install:
	@echo "===> Installing '$(OUTPUT)' to $(GOPATH)/bin..."
	go build -o $(OUTPUT)
	mv $(OUTPUT) $(GOPATH)/bin/

deps:
	@echo "===> Installing runtime dependencies..."
	go get -d -v ./...

updatedeps:
	@echo "===> Updating runtime dependencies..."
	go get -u -v ./...

build: deps
	@echo "===> Beginning compile..."
	gox -os $(XC_OS) -arch $(XC_ARCH) -output "pkg/{{.OS}}_{{.Arch}}/$(OUTPUT)"

dist: build
	@echo "===> Shipping packages..."
	rm -rf $(DISTDIR)
	mkdir -p $(DISTDIR)
	@for dir in $$(find $(BUILDDIR) -mindepth 1 -maxdepth 1 -type d); do \
		platform=`basename $$dir`; \
		if [ $$platform = "dist" ]; then \
			continue; \
		fi; \
		archive=$(OUTPUT)_$(VERSION)_$$platform; \
		zip -j $(DISTDIR)/$$archive.zip $$dir/*; \
		pushd $(DISTDIR); \
		shasum -a 256 * > ./$(VERSION)_SHA256SUMS; \
		popd; \
	done

release:
	@echo "===> Releasing to GitHub..."
	ghr -u $(OWNER) -r $(REPOSITORY) $(VERSION) $(DISTDIR)

clean:
	go clean
	rm -rf $(BUILDDIR)
