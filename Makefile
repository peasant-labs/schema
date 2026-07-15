.PHONY: schema check fmt vet test freshness gates lint-specs nix-vendor-hash docs clean typescript-deps typescript-check

GO ?= go

# BASE_REF: the ref the breaking-change gates (oasdiff / go-apidiff) diff against.
# CI sets it to the PR base (origin/develop). Locally, default to origin/develop;
# override on the CLI: `make gates BASE_REF=HEAD~1`.
BASE_REF ?= origin/develop

# Regenerate the OpenAPI specs + Redoc HTML from the Go source of truth.
schema: typescript-deps
	$(GO) run ./cmd/schema-gen

typescript-deps:
	@if [ ! -x typescript/node_modules/.bin/openapi-typescript ]; then \
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

# check is the schema repo's quality gate — the make-check-equivalent of peasant's
# `make check`, minus the peasant-specific web build and ast-grep step (no
# ast-grep ruleset is configured in this contract-only leaf). The Go race suite
# includes the leaf-audit test and the oasdiff/go-apidiff synthetic-break tests,
# which t.Skip with an actionable message when a gate binary is absent (e.g.
# outside `nix develop`) and run for real inside the flake dev shell. The full
# gate also requires the locked TypeScript dependencies.
#
# The release-workflow guard (release-guard check-workflow) keeps publication
# behind the required gates; the race suite is the authoritative test run.
check: fmt vet freshness typescript-check
	$(GO) run ./cmd/release-guard check-workflow --policy .github/release-guard.policy.yml --release .github/workflows/release.yml
	$(GO) test -race ./...

test:
	$(GO) test -race ./...

# freshness is the git-diff backstop: regenerate the specs and fail
# if the committed generated/ artifacts drift from the Go source. It is largely
# REDUNDANT with TestCodegenFreshness (which byte-diffs the SAME
# GenerateSpecArtifacts() map main() writes); its marginal value is catching a
# FUTURE main() write-path that emits a *tracked* artifact OUTSIDE that shared
# map (the Go test, iterating the map, could not see such a file). `git add -N`
# surfaces any newly-emitted untracked file to `git diff --exit-code`.
freshness: typescript-deps
	$(GO) run ./cmd/schema-gen
	@git add -N -- generated/ testdata/session-detail/redactions.yaml typescript/src/index.ts typescript/src/local-api.ts typescript/src/village-api.ts typescript/src/types.ts typescript/src/fixtures/quality.ts typescript/src/internal/generated/
	@if ! git diff --quiet -- generated/ testdata/session-detail/redactions.yaml typescript/src/index.ts typescript/src/local-api.ts typescript/src/village-api.ts typescript/src/types.ts typescript/src/fixtures/quality.ts typescript/src/internal/generated/; then \
		echo "generated artifacts drifted from the Go source — run 'make schema' and commit the result."; \
		git --no-pager diff --stat -- generated/ testdata/session-detail/redactions.yaml typescript/src/index.ts typescript/src/local-api.ts typescript/src/village-api.ts typescript/src/types.ts typescript/src/fixtures/quality.ts typescript/src/internal/generated/; \
		exit 1; \
	fi

# gates runs the breaking-change contract gates that need git history: vacuum
# lint of the committed specs, oasdiff vs the base ref, and go-apidiff vs the
# base ref. Requires the flake tools (oasdiff/go-apidiff/vacuum) — run inside
# `nix develop`. CI runs this with fetch-depth:0 so BASE_REF is resolvable.
gates:
	./scripts/contract-gates.sh all $(BASE_REF)

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
