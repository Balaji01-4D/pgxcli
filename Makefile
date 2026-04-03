MAIN_PATH = "cmd/pgxcli/main.go"
BUILD_PATH = "bin"

.PHONY: build clean run update runc lint test precommit

build:
	@mkdir -p $(BUILD_PATH)
	@go build -o $(BUILD_PATH)/app $(MAIN_PATH)

lint:
	@golangci-lint run

test:
	@go test ./...

precommit: lint test

runc: build
	@./bin/app $(DB)

run:
	@./bin/app $(DB)

clean:
	@rm -rf $(BUILD_PATH)

update:
	@go get -u ./...
