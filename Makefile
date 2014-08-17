.PHONY: clean

default: test

test: deps
	@echo "===> Running tests..."
	go test -v ./...

install:
	@echo "===> Installing command..."
	go install

deps:
	@echo "===> Installing dependencies..."
	go get -d -v ./...

updatedeps:
	@echo "===> Updating dependencies..."
	go get -u -v ./...

build: deps
	@echo "===> Beginning compile..."
	gox -os "darwin linux windows" -output "pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

clean:
	go clean
	rm -rf pkg
