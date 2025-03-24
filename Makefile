ROOT_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
build:
	@export PKG_CONFIG_PATH=$(ROOT_DIR)/app/internal/getter/ffmpeg/lib/pkgconfig && go build -o ./bin/$(program)/$(program).exe ./app/cmd/$(program)

run: build
	@./bin/$(program)/$(program).exe

test-cover:
	@go test -coverprofile=c.out ./... -v
	@go tool cover -html=c.out

test:
	@go test ./... -v