.PHONY: all build login run test clean

# 编译主程序
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o heybox-mcp .

# 编译登录工具
login:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o heybox-login ./cmd/login

# 编译所有
all: build login

# 运行主程序
run:
	go run .

# 运行登录工具
run-login:
	go run ./cmd/login/main.go

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -f heybox-mcp heybox-login
	go clean

# 安装依赖
deps:
	go mod download
	go mod tidy

# Docker 构建
docker-build:
	docker build -t heybox-mcp:latest .

# Docker 运行
docker-run:
	docker run -p 18060:18060 -v $(PWD)/data:/app/data heybox-mcp:latest
