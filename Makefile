projectname?=clispot

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build golang binary with embedded JS backend
	@echo "--> Building Go application..."
	@go build -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) -X main.spotifyClientID=$(SPOTIFY_CLIENT_ID) -X main.spotifyClientSecret=$(SPOTIFY_CLIENT_SECRET)" -o $(projectname)

.PHONY: install
install: build ## install binary to /usr/local/bin
	@echo "--> Installing clispot to /usr/local/bin..."
	@sudo cp $(projectname) /usr/local/bin/
	@echo "--> Installation complete. Run 'clispot' to start."

.PHONY: run
run: build ## build and run the app
	@./$(projectname)

.PHONY: bootstrap
bootstrap: ## install build deps
	go generate -tags tools tools/tools.go

.PHONY: test
test: clean ## display test coverage
	go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -rnk3

.PHONY: clean
clean: ## clean up environment
	@echo "--> Cleaning up..."
	@rm -rf coverage.out dist/ $(projectname)
	@rm -rf embed
	@rm -rf js/bin
	@rm -rf js/node_modules

.PHONY: cover
cover: ## display test coverage
	go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: fmt
fmt: ## format go files
	gofumpt -w .
	gci write .

.PHONY: lint
lint: ## lint go files
	golangci-lint run -c .golangci.yml

.PHONY: pre-commit
pre-commit:	## run pre-commit hooks
	pre-commit run --all-files

.PHONY: hooks
hooks: ## install git commit-msg hook for commitlint (local)
	@chmod +x scripts/hooks/commit-msg
	@git config core.hooksPath scripts/hooks
	@echo "--> Git hooks installed (commit-msg)."