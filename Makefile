BINARY        := advisory
BUILD_DIR     := .
GO_FLAGS      := -trimpath -ldflags="-s -w"
LAMBDA_ZIP    := bootstrap.zip
LAMBDA_ZIP_ARM := bootstrap-arm64.zip

.PHONY: all run build clean test lint lambda lambda-arm64 help

all: build

build:
	go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY) .

run: build
	npx @modelcontextprotocol/inspector ./$(BINARY)

test:
	go test ./test/... -v -count=1

lint:
	golangci-lint run ./...

# lambda: 为 x86_64 架构的 Lambda 函数构建部署包（默认）
lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build $(GO_FLAGS) -o bootstrap .
	zip -j $(LAMBDA_ZIP) bootstrap
	rm -f bootstrap
	@echo "Lambda 部署包已生成 (x86_64): $(LAMBDA_ZIP)"

# lambda-arm64: 为 arm64 架构的 Lambda 函数构建部署包（Graviton2，成本更低）
lambda-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build $(GO_FLAGS) -o bootstrap .
	zip -j $(LAMBDA_ZIP_ARM) bootstrap
	rm -f bootstrap
	@echo "Lambda 部署包已生成 (arm64): $(LAMBDA_ZIP_ARM)"

clean:
	rm -f $(BUILD_DIR)/$(BINARY) bootstrap $(LAMBDA_ZIP) $(LAMBDA_ZIP_ARM)

help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
