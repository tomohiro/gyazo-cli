.PHONY: clean

install: deps
	@go install

deps:
	@go get -d -v ./...
	@go build -v ./...

test: deps
	@go test -v ./...

build:
	@gox -os "darwin linux windows" -output "pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

clean:
	@go clean
	@rm -rf pkg
