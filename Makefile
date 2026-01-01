MAIN_PATH = "cmd/pgxcli/main.go"
BUILD_PATH = "bin"

.PHONY: build clean run update runc

build:
	@mkdir -p $(BUILD_PATH)
	@/usr/bin/time -f "Time: %E" \
		go build -o $(BUILD_PATH)/app $(MAIN_PATH)
	@du -sh $(BUILD_PATH)

runc: build
	@./bin/app $(DB)

run:
	@./bin/app $(DB)

clean:
	@rm -rf $(BUILD_PATH)

update:
	@go get -u ./...
