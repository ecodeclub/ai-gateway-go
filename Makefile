#.PHONY:	bench
#bench:
#	@go test -bench=. -benchmem  ./...

#.PHONY:	ut
#ut:
#	@go test -race -v ./... -failfast

.PHONY: e2e
e2e:
	@docker compose -f docker-compose.yaml up -d
	@go	test -race -v -failfast ./...
	@docker compose -f docker-compose.yaml down

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
	@$(MAKE) lint