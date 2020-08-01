BINARY_NAME=clinvar-matcher
MAIN=main.go
BUILD_DIR=builds

build:
	GOOS=darwin go build -o $(BUILD_DIR)/osx/$(BINARY_NAME) -v $(MAIN)
	zip -r -j $(BUILD_DIR)/osx/$(BINARY_NAME)_osx.zip $(BUILD_DIR)/osx/$(BINARY_NAME)
	GOOS=linux go build -o $(BUILD_DIR)/linux/$(BINARY_NAME) -v $(MAIN)
	zip -r -j $(BUILD_DIR)/linux/$(BINARY_NAME)_linux.zip $(BUILD_DIR)/linux/$(BINARY_NAME)
	GOOS=windows go build -o $(BUILD_DIR)/win/$(BINARY_NAME).exe -v $(MAIN)
	zip -r -j $(BUILD_DIR)/win/$(BINARY_NAME)_win.zip $(BUILD_DIR)/win/$(BINARY_NAME).exe

release:
	hub release create -a $(BUILD_DIR)/linux/$(BINARY_NAME)_linux.zip -a $(BUILD_DIR)/osx/$(BINARY_NAME)_osx.zip -a $(BUILD_DIR)/win/$(BINARY_NAME)_win.zip -m 'Initial Release' v0.0.1

tests:
	go test -v ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf $(BUILD_DIR)