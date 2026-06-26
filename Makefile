BINARY     := advisory
BUILD_DIR  := .
GO_FLAGS   := -trimpath -ldflags="-s -w"
LAMBDA_ZIP := bootstrap.zip

.PHONY: all run build clean test lint lambda help

all: build

build:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY) .

run: build
	npx @modelcontextprotocol/inspector ./$(BINARY)

test:
	go test ./test/... -v -count=1

lint:
	golangci-lint run ./...

lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build $(GO_FLAGS) -o bootstrap .
	zip -j $(LAMBDA_ZIP) bootstrap
	rm -f bootstrap
	@echo "Lambda 部署包已生成: $(LAMBDA_ZIP)"


clean:
	rm -f $(BUILD_DIR)/$(BINARY) bootstrap $(LAMBDA_ZIP)

help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
