build:
	@go build -o ./bin/$(program)/$(program).exe $(program)/cmd/main.go

run: build
	@./bin/$(program)/$(program).exe
