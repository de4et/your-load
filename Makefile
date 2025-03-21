ROOT_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
build:
	@export PKG_CONFIG_PATH=$(ROOT_DIR)getter/ffmpeg/lib/pkgconfig && go build -o ./bin/$(program)/$(program).exe ./$(program)/cmd

run: build
	@./bin/$(program)/$(program).exe

test-cover:
	@go test -coverprofile=c.out ./... -v
	@go tool cover -html=c.out

test:
	@go test ./... -v