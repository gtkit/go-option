.PHONY: manual tool check

LocalApp=optioner

manual:
	go build  -ldflags "-s -w" -gcflags="-m"  -o ${LocalApp} cmd/go-option/main.go && upx -9 ${LinuxApp}

LINT_TARGETS ?= ./...
tool: ## Lint Go code with the installed golangci-lint
	@ echo "▶️ golangci-lint run"
	golangci-lint run $(LINT_TARGETS)
	gofumpt -l -w .
	@ echo "✅ golangci-lint run"

## govulncheck 检查漏洞 go install golang.org/x/vuln/cmd/govulncheck@latest
check:
	govulncheck ./...
	gosec ./...
