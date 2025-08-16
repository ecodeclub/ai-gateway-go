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
	@$(MAKE) e2e_down
	@$(MAKE) e2e_up
	@go test -race -failfast -tags=e2e -count=1 -coverprofile=cover.out -coverpkg=./... ./...
	@$(MAKE) e2e_down

.PHONY: e2e_up
e2e_up:
	docker compose -p ai_gateway_platform -f ./.script/docker-compose.yaml up -d

.PHONY: e2e_down
e2e_down:
	docker compose -p ai_gateway_platform -f ./.script/docker-compose.yaml down -v

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
	#@$(MAKE) lint

# 生成gRPC相关文件
.PHONY: grpc
grpc:
	@buf format -w api/proto
	@buf lint api/proto
	@buf generate api/proto
