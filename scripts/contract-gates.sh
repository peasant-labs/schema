#!/usr/bin/env bash
# Contract gates for the peasant-labs/schema OpenAPI contract.
#
# Three gates, all provisioned from the flake dev shell (oasdiff / go-apidiff /
# vacuum-go — NEVER go.mod requires; the leaf-audit test enforces that). CI
# (.github/workflows/tests.yml `contract-gates` job) invokes these against the
# PR base ref; run locally with `nix develop -c make gates BASE_REF=origin/develop`.
#
#   vacuum     lint every generated OpenAPI spec; fail on `error` severity.
#              Hermetic (no git) — also wired into `make check` indirectly via the
#              synthetic-break tests, but run standalone here for the real specs.
#   oasdiff    breaking-change diff of each generated spec vs the SAME file on the
#              base ref (the "prior released golden"); fail on ERR-level breaks.
#              New version files (absent on base) have nothing to diff and skip.
#   go-apidiff exported-Go-API breaking-change diff of the module vs the base ref.
#              PRE-1.0 POLICY: incompatible changes are ADVISORY — they surface as a
#              ::warning:: (and, in CI, a sticky PR comment) but DO NOT fail the gate
#              (exit 0). An intentional spec-version stamp bump (VillageAPIVersion /
#              PeasantLocalAPIVersion / TypesVersion value change) stays fully exempt
#              (no warning). The diff runs WITH --print-compatible so additive
#              (non-breaking) changes are surfaced to reviewers on a SEPARATE
#              informational channel; the DECISION still consumes only the incompatible
#              section. An unparseable "Incompatible changes" section FAILS CLOSED
#              (exit 1) — a blind gate must stop the line. See gate_go_apidiff /
#              stamp_exempt_regex for the full rationale. go-apidiff's go-git
#              implementation cannot safely inspect a linked Git worktree, so the gate
#              resolves immutable commits in the caller repository and performs the diff
#              in a temporary standalone --no-local clone instead.
#
#              CI hand-off (env-gated; the vars are unset locally, so nothing is written
#              off-CI): after go-apidiff completes its comparison, the runner encodes its
#              outcome as a TRI-STATE in APIDIFF_CHANGES_FILE — a non-empty file = WARN
#              (the incompatible payload), an EMPTY file = cleanliness established, an
#              ABSENT file = fail-closed / gate blind — and writes the compatible bullets
#              to APIDIFF_COMPATIBLE_FILE. An unexpected non-zero go-apidiff exit, including
#              exit 1 without the required incompatible-report header, is fail-closed even
#              when it emitted output that looks otherwise parseable.
#
# Usage:
#   contract-gates.sh vacuum
#   contract-gates.sh oasdiff            <BASE_REF>
#   contract-gates.sh go-apidiff         <BASE_REF>
#   contract-gates.sh evaluate-apidiff   # reads go-apidiff output on stdin; the pure
#                                        # incompatible decision — non-exempt bullets on
#                                        # stdout (empty = clean), exit 0 warn/clean or
#                                        # 1 fail-closed (for tests)
#   contract-gates.sh extract-compatible # reads go-apidiff output on stdin; prints the
#                                        # compatible (additive) bullets on stdout,
#                                        # exit 0 always — informational (for tests)
#   contract-gates.sh all                <BASE_REF>
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

# OpenAPI specs under generated/ (the *-api-*.yaml surfaces + the types catalog).
# The publish-request-*.schema.json file is a JSON Schema, not an OpenAPI doc, so
# it is excluded from the OpenAPI gates.
openapi_specs() {
  find generated -maxdepth 1 -type f \
    \( -name '*-api-*.yaml' -o -name 'types-*.yaml' \) | sort
}

require_bin() {
  local bin="$1"
  if ! command -v "${bin}" >/dev/null 2>&1; then
    cat >&2 <<EOF
contract-gates: required tool '${bin}' is not on PATH.
  why:  the contract gates run through the flake dev shell, which provisions
        oasdiff, go-apidiff, and vacuum (none are go.mod requires).
  fix:  run inside the dev shell, e.g. 'nix develop -c make gates BASE_REF=origin/develop',
        or 'direnv allow' so the shell loads automatically.
EOF
    exit 127
  fi
}

gate_vacuum() {
  require_bin vacuum
  local rc=0 spec
  while IFS= read -r spec; do
    echo ">> vacuum lint ${spec}"
    if ! vacuum lint --no-banner --fail-severity error -r .vacuum.yaml "${spec}"; then
      echo "::error::vacuum found error-severity violations in ${spec}" >&2
      rc=1
    fi
  done < <(openapi_specs)
  return "${rc}"
}

gate_oasdiff() {
  require_bin oasdiff
  local base_ref="$1"
  if [[ -z "${base_ref}" ]]; then
    # No base to diff against (e.g. the very first release tag). The breaking
    # gate has nothing to compare, so it is a no-op — vacuum still lints the
    # published specs.
    echo ">> oasdiff: no BASE_REF given; skipping the breaking-change diff (nothing prior to compare)."
    return 0
  fi
  local rc=0 spec
  while IFS= read -r spec; do
    if git cat-file -e "${base_ref}:${spec}" 2>/dev/null; then
      echo ">> oasdiff breaking ${base_ref}:${spec} -> ${spec}"
      if ! oasdiff breaking "${base_ref}:${spec}" "${spec}" --fail-on ERR; then
        echo "::error::oasdiff found ERR-level breaking changes in ${spec} vs ${base_ref}. If intentional, publish a NEW spec version instead of mutating the released one." >&2
        rc=1
      fi
    else
      echo ">> oasdiff: ${spec} is new on this branch (absent at ${base_ref}); nothing to diff."
    fi
  done < <(openapi_specs)
  return "${rc}"
}

# stamp_exempt_regex matches — and ONLY matches — a go-apidiff "Incompatible
# changes" bullet reporting a VALUE CHANGE of a spec-version STAMP const. These three
# consts (VillageAPIVersion / PeasantLocalAPIVersion / TypesVersion, from versions.go)
# are version MARKERS, not part of the behavioural API surface: their value is the pin
# target a consumer asserts against and the trigger that names the newly-generated
# golden spec files. A bump is therefore an INTENTIONAL, expected "change" that
# go-apidiff (which only sees a const's value moved) mislabels as an incompatible
# break. The bump's real correctness is enforced elsewhere — its DRIFT is caught at
# the CONSUMER's CI (the consumer pins + asserts the version), and its
# new-golden-vs-mutated-golden safety is already covered on the module side by the
# oasdiff breaking gate plus the retired-spec immutability and generated-dir
# completeness guards. So a stamp value-change is exempt from THIS gate.
#
# The match is deliberately tight: the EXACT const names, anchored, and only the
# ` value changed from "…" to "…"` form. A stamp that is REMOVED or RENAMED
# (`- VillageAPIVersion: removed`), a NON-stamp const value change
# (`- OtherVersion: value changed …`), or ANY other incompatible change does NOT
# match and is still reported (a pre-1.0 WARNING — emitted on stdout as the payload);
# and a bare "Incompatible changes" header with no parseable bullets fails closed. Each
# of those cases is pinned as a regression guard
# in TestGoAPIDiffStampExemptionFilter (exhaustive, canned go-apidiff v0.8.3 strings
# through this same decision), with TestGoAPIDiffStampExemption anchoring those
# strings to go-apidiff's real output format — together the proof that this exemption
# cannot mask a real break.
stamp_exempt_regex='^[[:space:]]*-[[:space:]]+(VillageAPIVersion|PeasantLocalAPIVersion|TypesVersion): value changed from "[^"]*" to "[^"]*"$'

# evaluate_go_apidiff reads raw go-apidiff output on stdin and is the PURE incompatible
# decision of the go-apidiff gate (no process invocation), so it can be exercised
# directly by a synthetic-break test. Under the pre-1.0 ADVISORY policy it is a
# two-state exit plus a stdout payload:
#   STDOUT : the non-exempt INCOMPATIBLE bullet lines — the "warn" payload (EMPTY when
#            there is no incompatible section or the only bullets are exempt stamps).
#   EXIT 0 : satisfied — clean OR warned (clean-vs-warned = empty-vs-non-empty stdout);
#            a non-exempt incompatible change is a WARNING (::warning:: on stderr), not
#            a failure.
#   EXIT 1 : FAIL-CLOSED — an "Incompatible changes" section is present but no bullet
#            items parse out of it (the output format may have changed; do NOT silently
#            pass an unparsed break — a blind gate must stop the line).
# The exempt/warn verdict text goes to STDERR so STDOUT stays a pure payload channel.
evaluate_go_apidiff() {
  local out; out="$(cat)"

  # No incompatible-changes section anywhere -> nothing to exempt, pass.
  if ! printf '%s\n' "${out}" | grep -qiE 'Incompatible changes'; then
    return 0
  fi

  # Collect the incompatible-change bullet lines: the "- …" items that appear under
  # an "Incompatible changes:" heading, per package block. A "Compatible changes:"
  # heading or a new (non-indented) package header closes the block.
  local incompatible
  incompatible="$(printf '%s\n' "${out}" | awk '
    /^[[:space:]]*Incompatible changes:/ { inblock=1; next }
    /^[[:space:]]*Compatible changes:/   { inblock=0; next }
    /^[^[:space:]]/                      { inblock=0 }
    inblock && /^[[:space:]]*-[[:space:]]/ { print }
  ')"

  # Fail-closed: header seen but no bullets parsed -> unexpected format.
  if [[ -z "$(printf '%s' "${incompatible}" | tr -d '[:space:]')" ]]; then
    echo "::error::go-apidiff reported an 'Incompatible changes' section but no change bullets could be parsed; refusing to pass (its output format may have changed)." >&2
    return 1
  fi

  # Drop the exempt spec-version stamp value-changes; anything left is a reportable break.
  local remaining
  remaining="$(printf '%s\n' "${incompatible}" | grep -vE "${stamp_exempt_regex}" || true)"
  if [[ -z "$(printf '%s' "${remaining}" | tr -d '[:space:]')" ]]; then
    echo "go-apidiff: the only incompatible change(s) are intentional spec-version stamp bump(s) (VillageAPIVersion/PeasantLocalAPIVersion/TypesVersion); exempted." >&2
    return 0
  fi
  echo "::warning::go-apidiff reports incompatible exported-Go-API change(s) beyond the exempt spec-version stamps. Pre-1.0 policy: WARNING, not a failure — ship a version bump when releasing. Changed API(s):" >&2
  printf '%s\n' "${remaining}" >&2
  printf '%s\n' "${remaining}"
  return 0
}

# extract_go_apidiff_compatible reads raw go-apidiff output (produced WITH
# --print-compatible) on stdin and prints the COMPATIBLE (non-breaking/additive) change
# bullet lines on STDOUT, one per line (empty if none). INFORMATIONAL ONLY — it never
# affects the gate decision, the exit code, or the fail-closed guard (all of which are
# incompatible-scoped). Mirror image of the incompatible collector in
# evaluate_go_apidiff: a "Compatible changes:" heading opens the block; an "Incompatible
# changes:" heading or a new (non-indented) package header closes it. Because the two
# collectors close on each other's heading, neither section's bullets can leak into the
# other's payload.
extract_go_apidiff_compatible() {
  local out; out="$(cat)"
  printf '%s\n' "${out}" | awk '
    /^[[:space:]]*Compatible changes:/   { inblock=1; next }
    /^[[:space:]]*Incompatible changes:/ { inblock=0; next }
    /^[^[:space:]]/                      { inblock=0 }
    inblock && /^[[:space:]]*-[[:space:]]/ { print }
  '
}

# resolve_go_apidiff_commit resolves a caller-repository reference to the immutable commit
# hash that the isolated invocation receives. Resolving before cloning prevents a moving
# branch ref from changing either side of the comparison while setup is in progress.
resolve_go_apidiff_commit() {
  local ref="$1" label="$2" resolved
  if ! resolved="$(git -C "${repo_root}" rev-parse --verify "${ref}^{commit}" 2>&1)"; then
    echo "::error::go-apidiff could not resolve ${label} reference ${ref@Q} in ${repo_root}." >&2
    echo "  why: the comparison needs an immutable commit for each side; the reference is missing or is not a commit." >&2
    echo "  fix: fetch or correct ${label}, then rerun make gates BASE_REF=<commit-or-ref>. git said: ${resolved}" >&2
    return 1
  fi
  printf '%s\n' "${resolved}"
}

# go_apidiff_timeout_seconds validates the bounded timeout used for the isolated tool
# process. The environment knob is deliberately constrained so a malformed or unbounded
# value cannot turn the contract gate into a hanging CI step.
go_apidiff_timeout_seconds() {
  local seconds="${GO_APIDIFF_TIMEOUT_SECONDS:-120}"
  if ! [[ "${seconds}" =~ ^[1-9][0-9]*$ ]] || [[ "${#seconds}" -gt 3 ]] || (( 10#${seconds} > 600 )); then
    echo "::error::GO_APIDIFF_TIMEOUT_SECONDS=${seconds@Q} is invalid for the go-apidiff gate." >&2
    echo "  why: isolated API comparison must have a bounded whole-second timeout from 1 through 600." >&2
    echo "  fix: unset GO_APIDIFF_TIMEOUT_SECONDS or set it to an integer in that range, then rerun the gate." >&2
    return 1
  fi
  printf '%s\n' "${seconds}"
}

# run_go_apidiff_isolated runs go-apidiff only in a temporary standalone clone. go-apidiff
# v0.8.3 uses go-git's worktree status and destructive checkout/reset operations, which
# misread linked-worktree gitdir/index topology. A clone made with --no-local (never
# --shared) gives go-git an ordinary .git directory and avoids alternates that it cannot
# safely traverse. This function preserves go-apidiff's stdout, stderr, and exit status;
# setup or cleanup failures use 125 so callers never mistake them for its advisory exit 1.
run_go_apidiff_isolated() (
  local base_hash="$1" head_hash="$2" timeout_seconds="$3"
  local temp_parent="${TMPDIR:-/tmp}" temp_root="" clone_root="" clone_head

  isolated_go_apidiff_fail() {
    local step="$1" reason="$2"
    echo "::error::go-apidiff isolated ${step} failed." >&2
    echo "  why: ${reason}" >&2
    echo "  where: temporary clone for ${base_hash} -> ${head_hash}." >&2
    echo "  fix: inspect the command output above, repair Git or temporary-directory access, then rerun make gates." >&2
    exit 125
  }

  cleanup_isolated_go_apidiff_clone() {
    local original_status=$?
    trap - EXIT
    if [[ -z "${temp_root}" || ! -d "${temp_root}" || "${temp_root}" != "${temp_parent%/}/schema-go-apidiff."* ]]; then
      echo "::error::go-apidiff isolated cleanup precondition failed." >&2
      echo "  why: the temporary clone directory was missing or did not match the directory created for this comparison." >&2
      echo "  where: expected ${temp_parent%/}/schema-go-apidiff.* after comparing ${base_hash} -> ${head_hash}." >&2
      echo "  fix: inspect temporary-directory permissions and remove the abandoned clone before rerunning the gate." >&2
      exit 125
    fi
    if ! rm -rf -- "${temp_root}"; then
      echo "::error::go-apidiff isolated cleanup failed for ${temp_root}." >&2
      echo "  why: the temporary standalone clone could not be removed after the comparison." >&2
      echo "  fix: remove the directory manually, repair temporary-directory permissions, then rerun the gate." >&2
      exit 125
    fi
    exit "${original_status}"
  }

  if ! temp_root="$(mktemp -d "${temp_parent%/}/schema-go-apidiff.XXXXXXXX")"; then
    isolated_go_apidiff_fail "temporary-directory creation" "mktemp could not create an isolated clone directory under ${temp_parent}."
  fi
  trap cleanup_isolated_go_apidiff_clone EXIT
  clone_root="${temp_root}/repository"

  if ! git clone --no-local --no-checkout "${repo_root}" "${clone_root}"; then
    isolated_go_apidiff_fail "clone" "git clone --no-local --no-checkout could not copy the caller repository."
  fi
  if [[ ! -d "${clone_root}/.git" ]]; then
    isolated_go_apidiff_fail "clone verification" "the clone does not have a standalone .git directory."
  fi
  if [[ -e "${clone_root}/.git/objects/info/alternates" || -L "${clone_root}/.git/objects/info/alternates" ]]; then
    isolated_go_apidiff_fail "clone verification" "the clone contains an objects/info/alternates file; shared object storage is prohibited."
  fi
  if ! git -C "${clone_root}" cat-file -e "${base_hash}^{commit}"; then
    isolated_go_apidiff_fail "base verification" "the standalone clone does not contain resolved base commit ${base_hash}."
  fi
  if ! git -C "${clone_root}" cat-file -e "${head_hash}^{commit}"; then
    isolated_go_apidiff_fail "head verification" "the standalone clone does not contain resolved head commit ${head_hash}."
  fi
  if ! git -C "${clone_root}" checkout --detach --quiet "${head_hash}"; then
    isolated_go_apidiff_fail "detached checkout" "Git could not detach the standalone clone at resolved head ${head_hash}."
  fi
  if ! clone_head="$(git -C "${clone_root}" rev-parse --verify HEAD^{commit})" || [[ "${clone_head}" != "${head_hash}" ]]; then
    isolated_go_apidiff_fail "detached checkout verification" "the standalone clone is not detached at resolved head ${head_hash}."
  fi

  local tool_status
  if (
    cd "${clone_root}"
    timeout --kill-after=5s "${timeout_seconds}s" go-apidiff "${base_hash}" "${head_hash}" --print-compatible --repo-path "${clone_root}"
  ); then
    exit 0
  else
    tool_status=$?
  fi
  if [[ "${tool_status}" -eq 124 ]]; then
    echo "::error::go-apidiff timed out after ${timeout_seconds}s while comparing ${base_hash} to ${head_hash} in the isolated clone." >&2
    echo "  fix: inspect module analysis or increase GO_APIDIFF_TIMEOUT_SECONDS up to 600, then rerun the gate." >&2
  fi
  exit "${tool_status}"
)

# gate_go_apidiff runs the exported-Go-API diff vs the base ref and applies the pre-1.0
# ADVISORY policy: incompatible changes WARN (exit 0), an unparseable incompatible
# section FAILS CLOSED (exit 1), and compatible changes are surfaced on a separate
# informational channel. When the CI env vars are set it also encodes go-apidiff's OWN
# outcome as a tri-state signal for the PR-comment step (see the top-of-file header):
#   APIDIFF_CHANGES_FILE    non-empty = warn payload / empty = clean established /
#                           ABSENT (never written) = fail-closed, gate blind.
#   APIDIFF_COMPATIBLE_FILE the compatible (additive) bullets (possibly empty).
# Both writes are guarded by `mkdir -p "$(dirname …)"`, and both are skipped entirely
# when the env var is unset (the local / non-CI path writes no files).
gate_go_apidiff() {
  require_bin go-apidiff
  require_bin git
  require_bin mktemp
  require_bin rm
  require_bin timeout
  local base_ref="$1"
  if [[ -z "${base_ref}" ]]; then
    echo ">> go-apidiff: no BASE_REF given; skipping the exported-API breaking diff (nothing prior to compare)."
    return 0
  fi
  local base_hash head_hash timeout_seconds
  if ! base_hash="$(resolve_go_apidiff_commit "${base_ref}" "BASE_REF")"; then
    return 1
  fi
  if ! head_hash="$(resolve_go_apidiff_commit "HEAD" "HEAD")"; then
    return 1
  fi
  if ! timeout_seconds="$(go_apidiff_timeout_seconds)"; then
    return 1
  fi

  echo ">> go-apidiff ${base_hash} -> ${head_hash} (isolated standalone clone; +compatible for reviewers)"
  local out tool_rc
  # --print-compatible: emit BOTH sections so reviewers see additive changes; the
  # DECISION below still consumes only the incompatible section. The wrapper returns the
  # real tool status while reserving 125 for clone, checkout, and cleanup failures.
  if out="$(run_go_apidiff_isolated "${base_hash}" "${head_hash}" "${timeout_seconds}" 2>&1)"; then
    tool_rc=0
  else
    tool_rc=$?
  fi
  printf '%s\n' "${out}"

  # go-apidiff uses exit 1 for a completed comparison that found incompatible API
  # changes. Preserve the pre-1.0 advisory policy only for that documented report form.
  # The isolated wrapper reserves 125 for setup/cleanup failures and timeout returns 124;
  # both, an exit 1 without the incompatible-report header, and any other non-zero exit
  # fail before the parser and before either CI signal file can be written.
  if [[ "${tool_rc}" -ne 0 ]]; then
    if [[ "${tool_rc}" -eq 1 ]] && printf '%s\n' "${out}" | grep -qE '^[[:space:]]*Incompatible changes:'; then
      echo "go-apidiff: exited 1 after reporting incompatible exported-Go-API changes; evaluating the advisory report." >&2
    else
      case "${tool_rc}" in
        124) echo "::error::go-apidiff timed out before it established an API report; refusing to write API-diff signals." >&2 ;;
        125) echo "::error::go-apidiff isolated setup or cleanup failed before it established an API report; refusing to write API-diff signals." >&2 ;;
        *)   echo "::error::go-apidiff invocation did not produce the required incompatible-change report header in gate_go_apidiff while comparing ${base_hash} to ${head_hash} (exit ${tool_rc}); refusing to parse its output or write API-diff signals." >&2 ;;
      esac
      echo "  why: exit 1 is advisory only for go-apidiff's parseable incompatible report; all other exits are tooling, timeout, setup, or cleanup failures." >&2
      echo "  fix: inspect the go-apidiff output above, repair the reported failure, then rerun make gates BASE_REF=${base_ref}." >&2
      return 1
    fi
  fi

  # Pure INCOMPATIBLE decision: STDOUT = non-exempt incompatible payload (empty = clean);
  # exit 0 = warn/clean, 1 = fail-closed.
  local payload rc=0
  payload="$(printf '%s\n' "${out}" | evaluate_go_apidiff)" || rc=$?

  # Fail-closed (unparseable incompatible section): return 1 BEFORE any write. An ABSENT
  # signal file tells the comment step "go-apidiff established nothing" (distinct from an
  # EMPTY = clean file).
  if [[ "${rc}" -ne 0 ]]; then
    echo "::error::go-apidiff gate failed vs ${base_ref}: an 'Incompatible changes' section could not be parsed (tooling/format drift, not a policy decision); refusing to pass. Fix the parser in scripts/contract-gates.sh." >&2
    return 1
  fi

  # Compatible (non-breaking) changes: informational reviewer aid on a SEPARATE channel.
  # Never affects the decision/exit; written only when the workflow provides the env.
  if [[ -n "${APIDIFF_COMPATIBLE_FILE:-}" ]]; then
    mkdir -p "$(dirname "${APIDIFF_COMPATIBLE_FILE}")"
    printf '%s\n' "${out}" | extract_go_apidiff_compatible > "${APIDIFF_COMPATIBLE_FILE}"
  fi

  if [[ -n "$(printf '%s' "${payload}" | tr -d '[:space:]')" ]]; then
    # WARN (non-fatal): the incompatible payload is the "warn" signal for the comment step.
    if [[ -n "${APIDIFF_CHANGES_FILE:-}" ]]; then
      mkdir -p "$(dirname "${APIDIFF_CHANGES_FILE}")"
      printf '%s\n' "${payload}" > "${APIDIFF_CHANGES_FILE}"
    fi
    echo "go-apidiff: incompatible exported-API change(s) vs ${base_ref} reported as a WARNING (pre-release policy); gate does not block."
    return 0
  fi

  # CLEAN (no incompatible, or exempt-stamp-only): write an EMPTY file — the
  # go-apidiff-specific "cleanliness established" signal (distinct from an ABSENT file =
  # fail-closed / gate blind).
  if [[ -n "${APIDIFF_CHANGES_FILE:-}" ]]; then
    mkdir -p "$(dirname "${APIDIFF_CHANGES_FILE}")"
    : > "${APIDIFF_CHANGES_FILE}"
  fi
  echo "go-apidiff: no disqualifying incompatible exported-API changes vs ${base_ref}."
  return 0
}

main() {
  local cmd="${1:-}"
  case "${cmd}" in
    vacuum)     gate_vacuum ;;
    oasdiff)    gate_oasdiff "${2:-}" ;;
    go-apidiff) gate_go_apidiff "${2:-}" ;;
    # evaluate-apidiff reads go-apidiff output on stdin and applies ONLY the pure
    # incompatible decision (the stamp-exemption filter) — the seam the synthetic-break
    # test drives with real go-apidiff output from a throwaway repo. STDOUT is the
    # non-exempt incompatible payload; exit 0 warn/clean, 1 fail-closed.
    evaluate-apidiff)   evaluate_go_apidiff ;;
    # extract-compatible reads go-apidiff output on stdin and prints the COMPATIBLE
    # (additive) bullets on stdout — the informational reviewer-aid seam, parallel to
    # evaluate-apidiff and never on the decision path (exit 0 always).
    extract-compatible) extract_go_apidiff_compatible ;;
    all)
      local base_ref="${2:-}" rc=0
      gate_vacuum || rc=1
      gate_oasdiff "${base_ref}" || rc=1
      gate_go_apidiff "${base_ref}" || rc=1
      return "${rc}"
      ;;
    *)
      echo "usage: contract-gates.sh <vacuum|oasdiff <BASE_REF>|go-apidiff <BASE_REF>|evaluate-apidiff|extract-compatible|all <BASE_REF>>" >&2
      exit 2
      ;;
  esac
}

main "$@"
