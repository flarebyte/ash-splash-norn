.DEFAULT_GOAL := help

.PHONY: build test test-go test-unit test-race test-fixtures \
	lint lint-go lint-ts format format-go format-ts \
	typecheck-ts review coverage coverage-go coverage-critical coverage-threshold \
	doc-design doc-decision cue-input dup complexity release sec \
	thoth-meta thoth-meta-go thoth-meta-go-test \
	check-tools install-tools-help help

GO := go
BUN := bun
GOLINT := golangci-lint
ROOT_DIR := $(CURDIR)
TMP_DIR := $(ROOT_DIR)/tmp
GO_CACHE_DIR := $(ROOT_DIR)/.gocache
GO_MOD_CACHE_DIR := $(ROOT_DIR)/.gomodcache
GO_LINT_CACHE_DIR := $(ROOT_DIR)/.golangci-lint-cache
E2E_BIN_DIR := $(ROOT_DIR)/.e2e-bin
GO_PACKAGES := ./...
GO_ENV := GOTOOLCHAIN=local GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)
BUN_ENV := TMPDIR=$(TMP_DIR)
BIOME := $(BUN_ENV) $(BUN) run biome
THOTH := thoth
FLYB := flyb
GOLINT_ENV := $(GO_ENV) GOLANGCI_LINT_CACHE=$(GO_LINT_CACHE_DIR)
COVER_PROFILE := $(TMP_DIR)/test-unit.coverage.out
COVER_HTML := $(TMP_DIR)/test-unit.coverage.html
CRITICAL_COVER_PROFILE := $(TMP_DIR)/critical.coverage.out
CRITICAL_PACKAGES := . ./internal/schema ./internal/argv ./internal/validators
COVERAGE_MIN := 98.0
DESIGN_META_CONFIGS := $(wildcard doc/design-meta/*.cue)

build:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) build $(GO_PACKAGES)

test: test-go

test-go: test-unit

test-fixtures:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) test -run 'TestCommandDocumentCompileFromFixture|TestEndToEndValidateWithDocumentJSON|TestSecurityRuntimeFixtures' ./...

test-unit:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) test -v -coverprofile=$(COVER_PROFILE) -covermode=count $(GO_PACKAGES)
	$(GO_ENV) $(GO) tool cover -func=$(COVER_PROFILE)

test-race:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) test -race $(GO_PACKAGES)

coverage: coverage-go

coverage-go: test-unit
	$(GO_ENV) $(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)
	@printf "Coverage HTML: %s\n" "$(COVER_HTML)"

coverage-critical:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) test -coverprofile=$(CRITICAL_COVER_PROFILE) -covermode=count $(CRITICAL_PACKAGES)
	$(GO_ENV) $(GO) tool cover -func=$(CRITICAL_COVER_PROFILE)
	@printf "Critical coverage profile: %s\n" "$(CRITICAL_COVER_PROFILE)"

coverage-threshold:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) test -coverprofile=$(COVER_PROFILE) -covermode=count $(GO_PACKAGES)
	@report="$$( $(GO_ENV) $(GO) tool cover -func=$(COVER_PROFILE) )"; \
	printf "%s\n" "$$report"; \
	below="$$(printf "%s\n" "$$report" | awk -v min=$(COVERAGE_MIN) ' \
	/^[^[:space:]].*:[0-9]+:/ { \
		pct=$$NF; gsub(/%/, "", pct); \
		if (pct+0 < min+0) print $$0; \
	} \
	')"; \
	if [ -n "$$below" ]; then \
		printf "\nFunctions below %.1f%% coverage:\n%s\n" $(COVERAGE_MIN) "$$below"; \
		exit 1; \
	fi

lint: lint-go lint-ts

lint-go:
	mkdir -p $(TMP_DIR)
	$(GO_ENV) $(GO) vet $(GO_PACKAGES)
	$(GOLINT_ENV) $(GOLINT) run

lint-ts:
	mkdir -p $(TMP_DIR)
	$(BUN_ENV) $(BUN) run biome check doc/design-meta/examples

format: format-go format-ts

review: format test test-fixtures lint

format-go:
	mkdir -p $(TMP_DIR)
	find . -type f -name '*.go' \
		-not -path './.git/*' \
		-not -path './.gocache/*' \
		-not -path './.gomodcache/*' \
		-not -path './.e2e-bin/*' \
		-not -path './node_modules/*' \
		-print0 | xargs -0 -r gofmt -w

format-ts:
	mkdir -p $(TMP_DIR)
	$(BUN_ENV) $(BUN) run biome format --write doc/design-meta/examples

typecheck-ts:
	mkdir -p $(TMP_DIR)
	$(BUN_ENV) $(BUN) x tsc -p tsconfig.design-meta.json

doc-design:
	mkdir -p doc/design
	$(MAKE) cue-input
	@if [ -z "$(DESIGN_META_CONFIGS)" ]; then \
		printf "No flyb config files found in doc/design-meta/*.cue\n"; \
		printf "Add a config (for example doc/design-meta/app.cue) then rerun make doc-design.\n"; \
		exit 1; \
	fi
	@for cfg in $(DESIGN_META_CONFIGS); do \
		printf "Running flyb for %s\n" "$$cfg"; \
		$(FLYB) validate --config "$$cfg"; \
		$(FLYB) generate markdown --config "$$cfg"; \
	done

cue-input:
	cue vet doc/design-meta/examples/input/config-key.cue doc/design-meta/examples/input/config-key.schema.cue

doc-decision:
	mkdir -p doc/decision-meta
	mkdir -p doc/decision
	sh ./script/generate-decision-docs.sh

dup:
	npx jscpd --format go --min-lines 10 --ignore "**/.gomodcache/**,**/.gocache/**,**/.e2e-bin/**,**/node_modules/**,**/dist/**" --gitignore .
	npx jscpd --format typescript --min-lines 10 --gitignore .

complexity:
	scc --sort complexity --by-file -i go . | head -n 15
	scc --sort complexity --by-file -i ts . | head -n 15

release:
	$(BUN_ENV) $(BUN) run release-go.ts

sec:
	semgrep scan --config auto

thoth-meta: thoth-meta-go thoth-meta-go-test

thoth-meta-go:
	$(THOTH) run --config ./pipeline-go-maat.thoth.cue

thoth-meta-go-test:
	$(THOTH) run --config ./pipeline-go-test-maat.thoth.cue

check-tools:
	@printf "go=%s\n" "$$(command -v $(GO) >/dev/null 2>&1 && printf true || printf false)"
	@printf "bun=%s\n" "$$(command -v $(BUN) >/dev/null 2>&1 && printf true || printf false)"
	@printf "golangci-lint=%s\n" "$$(command -v $(GOLINT) >/dev/null 2>&1 && printf true || printf false)"
	@printf "flyb=%s\n" "$$(command -v $(FLYB) >/dev/null 2>&1 && printf true || printf false)"
	@printf "thoth=%s\n" "$$(command -v $(THOTH) >/dev/null 2>&1 && printf true || printf false)"
	@printf "semgrep=%s\n" "$$(command -v semgrep >/dev/null 2>&1 && printf true || printf false)"
	@printf "scc=%s\n" "$$(command -v scc >/dev/null 2>&1 && printf true || printf false)"
	@printf "jscpd=%s\n" "$$(command -v npx >/dev/null 2>&1 && printf true || printf false)"

install-tools-help:
	@printf "Install required tools:\n"
	@printf "  go: https://go.dev/doc/install\n"
	@printf "  bun: https://bun.sh/docs/installation\n"
	@printf "  golangci-lint: https://golangci-lint.run/welcome/install/\n"
	@printf "  flyb: install the flyb CLI used for doc generation.\n"
	@printf "  thoth: install the thoth CLI used for metadata pipelines.\n"
	@printf "  semgrep: https://semgrep.dev/docs/getting-started/cli\n"
	@printf "  scc: https://github.com/boyter/scc\n"
	@printf "  jscpd: available via npx jscpd once Node/npm tooling is installed.\n"

help:
	@printf "Targets:\n"
	@printf "  build        Build all Go packages.\n"
	@printf "  test         Run Go tests.\n"
	@printf "  test-go      Run Go test targets.\n"
	@printf "  test-unit    Run verbose Go tests and print coverage summary.\n"
	@printf "  test-race    Run Go tests with the race detector.\n"
	@printf "  test-fixtures Run fixture-backed regression tests.\n"
	@printf "  coverage     Generate the coverage HTML report.\n"
	@printf "  coverage-go  Generate the Go coverage HTML report from test-unit output.\n"
	@printf "  coverage-critical  Report coverage for parser/compiler/argv/validators paths.\n"
	@printf "  coverage-threshold Run tests and fail if any function is below COVERAGE_MIN (default 98.0).\n"
	@printf "  lint         Run Go linting and TypeScript design-meta lint checks.\n"
	@printf "  lint-go      Run go vet and golangci-lint.\n"
	@printf "  lint-ts      Run Biome checks for doc/design-meta TypeScript examples.\n"
	@printf "  format       Format Go and TypeScript design-meta files.\n"
	@printf "  format-go    Format Go files with gofmt.\n"
	@printf "  format-ts    Format doc/design-meta TypeScript examples with Biome.\n"
	@printf "  typecheck-ts Type-check doc/design-meta TypeScript examples with tsc.\n"
	@printf "  review       Run format, test, and lint using existing targets.\n"
	@printf "  doc-design   Regenerate design docs from flyb configs.\n"
	@printf "  cue-input    Validate input CUE examples against schema contracts.\n"
	@printf "  doc-decision Validate decision configs and regenerate markdown decision reports.\n"
	@printf "  dup          Run duplicate code detection.\n"
	@printf "  complexity   Show top Go and TypeScript files by complexity.\n"
	@printf "  release      Run the local release helper script.\n"
	@printf "  sec          Run Semgrep security scan.\n"
	@printf "  thoth-meta   Refresh thoth metadata for Go and Go tests.\n"
	@printf "  check-tools  Report required tool availability as key=value pairs.\n"
	@printf "  install-tools-help  Show how to install required tools.\n"
	@printf "  help         Show this help message.\n"
