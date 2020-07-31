BINARY_NAME=clinvar-matcher
MAIN=main.go
BUILD_DIR=builds

build:
	GOOS=darwin go build -o $(BUILD_DIR)/osx/$(BINARY_NAME) -v $(MAIN)
	GOOS=linux go build -o $(BUILD_DIR)/linux/$(BINARY_NAME) -v $(MAIN)
	GOOS=windows go build -o $(BUILD_DIR)/win/$(BINARY_NAME).exe -v $(MAIN)

tests:
	go test -v ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf $(BUILD_DIR)