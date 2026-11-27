.PHONY: capture-differential capture-bench schema check fmt vet test race-selective freshness freeze-check gates lint-specs nix-vendor-hash docs clean typescript-deps typescript-check check-unstring unstring-update

GO ?= go
UNSTRING ?= go run github.com/peasant-labs/infra/unstring/cmd/unstring@v0.0.4

# BASE_REF: the ref the breaking-change gates (oasdiff / go-apidiff) diff against.
# CI sets it to the PR base (origin/develop). Locally, default to origin/develop;
# override on the CLI: `make gates BASE_REF=HEAD~1`.
BASE_REF ?= origin/develop
export TITLE

# RACE=1 runs the Go suites under the race detector. The default is no race:
# the detector roughly doubles to triples the suite's wall time, and the
# module's only production concurrency is the lazy schema-compilation caches,
# which `make race-selective` covers under -race in seconds. Opt in for a
# full race pass with `make check RACE=1` or `make test RACE=1`.
RACE ?= 0
GORACE_FLAG := $(if $(filter 1,$(RACE)),-race,)

# The tests race-selective runs, by exact name, space-separated. Each one
# must run and pass, so a renamed or deleted test fails the target instead of
# passing with zero tests. New production concurrency adds its test here.
RACE_SELECTIVE_TESTS ?= TestValidatorCachesConcurrently

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

# check is the schema repo's quality gate - the make-check-equivalent of peasant's
# `make check`, minus the peasant-specific web build. The unstring static guard
# runs here and needs ast-grep 0.51.2 from the dev shell. The Go suite
# includes the leaf-audit test and the oasdiff/go-apidiff synthetic-break tests,
# which t.Skip with an actionable message when a gate binary is absent (e.g.
# outside `nix develop`) and run for real inside the flake dev shell. The full
# gate also requires the locked TypeScript dependencies.
#
# The Go suite is the authoritative test run, without
# the race detector unless RACE=1 (see RACE above).
# The root package needs headroom beyond Go's default 10-minute timeout in CI.
check: fmt vet freshness typescript-check
	$(UNSTRING) check
	$(GO) test $(GORACE_FLAG) -timeout 20m ./...
	cd tools/redaction-categories && GOWORK=off $(GO) test $(GORACE_FLAG) ./...

check-unstring:
	$(UNSTRING) check

unstring-update:
	$(UNSTRING) update

test:
	# Match check's CI headroom for the root package suite.
	$(GO) test $(GORACE_FLAG) -timeout 20m ./...
	cd tools/redaction-categories && GOWORK=off $(GO) test $(GORACE_FLAG) ./...

# race-selective runs the race detector over the module's only production
# concurrency: the sync.Once and sync.Map schema-compilation caches behind the
# publish-request, annotation-push and described-value validators. The merge
# queue runs it beside the no-race `make check`.
race-selective:
	@set -eu; \
	log=$$(mktemp "$${TMPDIR:-/tmp}/schema-race-selective.XXXXXXXX"); \
	trap 'rm -f "$$log"' EXIT; \
	pattern="^($$(printf '%s' '$(RACE_SELECTIVE_TESTS)' | tr -s ' ' '|'))$$"; \
	$(GO) test -race -count=1 -v -run "$$pattern" . > "$$log" 2>&1 || { cat "$$log"; exit 1; }; \
	grep -E '^(--- |ok|PASS|FAIL)' "$$log"; \
	for name in $(RACE_SELECTIVE_TESTS); do \
		grep -q -- "^--- PASS: $$name " "$$log" || { echo "race-selective: $$name did not run and pass under -race; a renamed or deleted test must be updated in RACE_SELECTIVE_TESTS"; exit 1; }; \
	done

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

# capture-differential is local evidence for a change to schema.NewTranscriptCookTurns,
# not a CI gate: it builds tools/capture-differential twice, against the schema
# release pinned in its go.mod (GOWORK=off, tag capture_baseline) and against
# this working tree (through a temporary go.work, never a committed replace),
# runs both over the root capture fixture, the seeded mutations, the benchmark
# subjects and the fuzz cache, and fails on the first differing verdict line.
# RECORD=1 rewrites the root fixture's expectations from the baseline build only.
CAPTURE_DIFF_DIR := tools/capture-differential
FUZZ_DIR ?= $(shell $(GO) env GOCACHE)/fuzz/github.com/peasant-labs/schema/FuzzTranscriptCapture
RECORD ?= 0
CAPTURE_BENCH_CPUS ?= 0-3
CAPTURE_BENCH_ROUNDS ?= 10
CAPTURE_BENCH_OUT ?= $${TMPDIR:-/tmp}/capture-bench

define capture_builds
	if [ -n "$${GOEXPERIMENT:-}" ]; then echo "capture-differential: unset GOEXPERIMENT; encoding/json defines the depth limit and escape decoding both builds must share"; exit 1; fi; \
	root=$$(pwd); \
	printf 'go 1.26.0\n\nuse (\n\t%s\n\t%s\n)\n' "$$root" "$$root/$(CAPTURE_DIFF_DIR)" > "$$tmp/go.work"
endef

capture-differential:
	@set -eu; \
	tmp=$$(mktemp -d "$${TMPDIR:-/tmp}/capture-differential.XXXXXXXX"); \
	trap 'rm -rf "$$tmp"' EXIT; \
	$(capture_builds); \
	(cd $(CAPTURE_DIFF_DIR) && GOWORK=off $(GO) build -tags capture_baseline -o "$$tmp/baseline" .); \
	if [ "$(RECORD)" = 1 ]; then "$$tmp/baseline" record -root "$$root/testdata/transcript_capture.yaml"; exit 0; fi; \
	(cd $(CAPTURE_DIFF_DIR) && GOWORK="$$tmp/go.work" $(GO) build -o "$$tmp/candidate" .); \
	set -- -root "$$root/testdata/transcript_capture.yaml" \
		-seeds "$$root/$(CAPTURE_DIFF_DIR)/testdata/mutation_seeds.yaml" \
		-subjects "$$root/$(CAPTURE_DIFF_DIR)/testdata/subjects.yaml" \
		-fuzz "$(FUZZ_DIR),$$root/testdata/fuzz/FuzzTranscriptCapture"; \
	"$$tmp/baseline" run "$$@" > "$$tmp/baseline.txt"; \
	"$$tmp/candidate" run "$$@" > "$$tmp/candidate.txt"; \
	"$$tmp/candidate" compare "$$tmp/baseline.txt" "$$tmp/candidate.txt"

# capture-bench runs BenchmarkPartitionPass for both builds, alternating for
# CAPTURE_BENCH_ROUNDS rounds pinned to CAPTURE_BENCH_CPUS, and leaves
# baseline.txt and candidate.txt in CAPTURE_BENCH_OUT. Evidence only.
capture-bench:
	@set -eu; \
	tmp=$$(mktemp -d "$${TMPDIR:-/tmp}/capture-bench.XXXXXXXX"); \
	trap 'rm -rf "$$tmp"' EXIT; \
	$(capture_builds); \
	out=$(CAPTURE_BENCH_OUT); mkdir -p "$$out"; : > "$$out/baseline.txt"; : > "$$out/candidate.txt"; \
	(cd $(CAPTURE_DIFF_DIR) && GOWORK=off $(GO) test -c -tags capture_baseline -o "$$tmp/baseline.test" .); \
	(cd $(CAPTURE_DIFF_DIR) && GOWORK="$$tmp/go.work" $(GO) test -c -o "$$tmp/candidate.test" .); \
	pin=""; if command -v taskset >/dev/null; then pin="taskset -c $(CAPTURE_BENCH_CPUS)"; fi; \
	i=0; while [ $$i -lt $(CAPTURE_BENCH_ROUNDS) ]; do \
		for build in baseline candidate; do \
			(cd $(CAPTURE_DIFF_DIR) && GOMAXPROCS=4 $$pin "$$tmp/$$build.test" -test.run '^$$' -test.bench BenchmarkPartitionPass -test.benchmem -test.cpu 4 -test.benchtime 20x -test.count 1 >> "$$out/$$build.txt"); \
		done; \
		i=$$((i+1)); \
	done; \
	if command -v benchstat >/dev/null; then benchstat "$$out/baseline.txt" "$$out/candidate.txt"; else echo "benchstat is not on PATH; raw results are in $$out"; fi
