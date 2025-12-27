MAIN_PATH = "cmd/pgcli/main.go"
BUILD_PATH = "bin"

.PHONY: build clean run update

# check if the dir exits
build:
	@mkdir -p $(BUILD_PATH)
	@/usr/bin/time -f "Time: %E" \
		@go build -o $(BUILD_PATH)/app $(MAIN_PATH)
	@du -sh $(BUILD_PATH)

run: build
	./bin/app $(DB)

clean:
	@rm -rf $(BUILD_PATH)

update:
	@go get -u ./...

