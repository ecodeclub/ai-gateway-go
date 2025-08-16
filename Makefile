.PHONY:	bench
bench:
	@go test -bench=. -benchmem  ./...

#.PHONY:	ut
#ut:
#	@go test -race -v ./... -failfast

# 定义操作系统相关的睡眠命令
ifeq ($(OS),Windows_NT)  # 检测 Windows 系统
    SLEEP_CMD = powershell -Command Start-Sleep -Seconds 10
else                     # 其他系统默认为 Unix-like
    SLEEP_CMD = sleep 10
endif

.PHONY: e2e
e2e:
	@docker compose -f ./.script/docker-compose.yaml up -d
	@go	test -race -v -failfast -coverprofile=cover.out ./...
	@docker compose -f ./.script/docker-compose.yaml down


.PHONY:	fmt
fmt:
	@goimports -l -w $$(find . -type f -name '*.go'  -not -path "./.idea/*" -not -name '*.pb.go' -not -name '*mock*.go')

.PHONY:	lint
lint:
	@golangci-lint run -c .golangci.yml

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: check
check:
	@$(MAKE) fmt
	@$(MAKE) tidy
	@$(MAKE) lint

# 生成gRPC相关文件
.PHONY: grpc
grpc:
	@buf format -w api/proto
	@buf lint api/proto
	@buf generate api/proto
