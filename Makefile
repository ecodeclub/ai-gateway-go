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
	@echo "等待 10 秒确保容器启动完成..."
	@$(SLEEP_CMD)  # 根据系统动态选择命令
	@go	test -race -v -failfast ./...
	@docker compose -f ./.script/docker-compose.yaml down


.PHONY:	fmt
fmt:
	@goimports -l -w $$(find . -type f -name '*.go' -not -path "./.idea/*")

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