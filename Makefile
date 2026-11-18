.PHONY: schema check fmt vet test freshness freeze-check gates lint-specs nix-vendor-hash docs clean typescript-deps typescript-check check-unstring unstring-update

GO ?= go
UNSTRING ?= go run github.com/peasant-labs/infra/unstring/cmd/unstring@v0.0.4

# BASE_REF: the ref the breaking-change gates (oasdiff / go-apidiff) diff against.
# CI sets it to the PR base (origin/develop). Locally, default to origin/develop;
# override on the CLI: `make gates BASE_REF=HEAD~1`.
BASE_REF ?= origin/develop
export TITLE

# Regenerate OpenAPI/Redoc from Go, then derive the TypeScript contract package.
schema: typescript-deps
	$(GO) run ./cmd/schema-gen
	pnpm --dir typescript run generate

typescript-deps:
	@if [ ! -x typescript/node_modules/.bin/openapi-ts ] || [ ! -x typescript/node_modules/.bin/openapi-typescript ]; then \
		echo "TypeScript dependencies are missing; run 'pnpm --dir typescript install --frozen-lockfile --ignore-scripts' from the module root."; \
		exit 1; \
	fi

typescript-check: typescript-deps
	pnpm --dir typescript run typecheck
	pnpm --dir typescript test
	pnpm --dir typescript run package:audit
	pnpm --dir typescript run package:smoke

# gofmt gate: fail (non-zero) if any file needs formatting.
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt: the following files need formatting:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...
	cd tools/redaction-categories && GOWORK=off $(GO) vet ./...

# check is the schema repo's quality gate — the make-check-equivalent of peasant's
# `make check`, minus the peasant-specific web build. The unstring static guard
# runs here and needs ast-grep 0.51.2 from the dev shell. The Go race suite
# includes the leaf-audit test and the oasdiff/go-apidiff synthetic-break tests,
# which t.Skip with an actionable message when a gate binary is absent (e.g.
# outside `nix develop`) and run for real inside the flake dev shell. The full
# gate also requires the locked TypeScript dependencies.
#
# The release-workflow guard (release-guard check-workflow) keeps publication
# behind the required gates; the race suite is the authoritative test run.
# The root package needs headroom beyond Go's default 10-minute timeout in CI.
check: fmt vet freshness typescript-check
	$(UNSTRING) check
	$(GO) run ./cmd/release-guard check-workflow --policy .github/release-guard.policy.yml --release .github/workflows/release.yml
	$(GO) test -race -timeout 20m ./...
	cd tools/redaction-categories && GOWORK=off $(GO) test -race ./...

check-unstring:
	$(UNSTRING) check

unstring-update:
	$(UNSTRING) update

test:
	# Match check's CI headroom for the root package race suite.
	$(GO) test -race -timeout 20m ./...

# freshness is the git-diff backstop: regenerate the specs and fail
# if the committed generated/ artifacts drift from the Go source. It is largely
# REDUNDANT with TestCodegenFreshness (which byte-diffs the SAME
# GenerateSpecArtifacts() map main() writes); its marginal value is catching a
# FUTURE main() write-path that emits a *tracked* artifact OUTSIDE that shared
# map (the Go test, iterating the map, could not see such a file). `git add -N`
# surfaces any newly-emitted untracked file to `git diff --exit-code`.
freshness: typescript-deps
	$(GO) run ./cmd/schema-gen
	pnpm --dir typescript run generate
	@git add -N -- generated/ route/routes_gen.go testdata/session-detail/redactions.yaml typescript/src/internal/generated/
	@if ! git diff --quiet -- generated/ route/routes_gen.go testdata/session-detail/redactions.yaml typescript/src/internal/generated/; then \
		echo "generated artifacts drifted from their canonical Go/OpenAPI/YAML sources — run 'make schema' and commit the result."; \
		git --no-pager diff --stat -- generated/ route/routes_gen.go testdata/session-detail/redactions.yaml typescript/src/internal/generated/; \
		exit 1; \
	fi

freeze-check:
	$(GO) run ./cmd/schema-gen freeze --check

# gates conservatively refuses a break without a title. With TITLE set it uses
# release-intent to judge the report just as CI does. The binary must be on PATH.
gates:
	@set -eu; \
	report=$$(mktemp "$${TMPDIR:-/tmp}/schema-contract-report.XXXXXXXX.yaml"); \
	trap 'rm -f "$$report"' EXIT; \
	./scripts/contract-gates.sh all "$(BASE_REF)" --out "$$report"; \
	if [ -n "$${TITLE:-}" ]; then \
		command -v release-intent >/dev/null || { echo "release-intent is required on PATH to judge TITLE; build it from the pinned infra revision."; exit 1; }; \
		printf '%s' "$$TITLE" | release-intent check-contract --report "$$report"; \
	elif grep -Eq '^[[:space:]]*(go_api_breaking|breaking): true[[:space:]]*$$' "$$report"; then \
		echo "Contract report contains a break; rerun with TITLE='feat(scope)![minor]: describe the change' and the pinned release-intent on PATH."; \
		exit 1; \
	fi

# lint-specs runs just the hermetic vacuum lint (no git/base ref needed).
lint-specs:
	./scripts/contract-gates.sh vacuum

# Recompute the flake's buildGoModule vendorHash after a dependency bump.
nix-vendor-hash:
	./scripts/update-nix-vendor-hash.sh

docs: schema
	@echo "Docs generated (OpenAPI specs + Redoc HTML under docs/api/)"

clean:
	rm -rf docs/api/ typescript/dist/ typescript/.test-dist/
