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
#   go-apidiff exported-Go-API breaking-change diff of the module vs the base ref;
#              fail on any incompatible change EXCEPT an intentional spec-version
#              stamp bump (a VillageAPIVersion / PeasantLocalAPIVersion / TypesVersion
#              value change). Those consts are version MARKERS, not API surface (see
#              gate_go_apidiff / stamp_exempt_regex for the full rationale).
#
# Usage:
#   contract-gates.sh vacuum
#   contract-gates.sh oasdiff        <BASE_REF>
#   contract-gates.sh go-apidiff     <BASE_REF>
#   contract-gates.sh evaluate-apidiff   # reads go-apidiff output on stdin; the
#                                        # pure pass/fail decision (for tests)
#   contract-gates.sh all            <BASE_REF>
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
# match and still fails the gate; and a bare "Incompatible changes" header with no
# parseable bullets fails closed. Each of those cases is pinned as a regression guard
# in TestGoAPIDiffStampExemptionFilter (exhaustive, canned go-apidiff v0.8.3 strings
# through this same decision), with TestGoAPIDiffStampExemption anchoring those
# strings to go-apidiff's real output format — together the proof that this exemption
# cannot mask a real break.
stamp_exempt_regex='^[[:space:]]*-[[:space:]]+(VillageAPIVersion|PeasantLocalAPIVersion|TypesVersion): value changed from "[^"]*" to "[^"]*"$'

# evaluate_go_apidiff reads raw go-apidiff output on stdin and is the PURE pass/fail
# decision of the go-apidiff gate (no process invocation), so it can be exercised
# directly by a synthetic-break test. It returns:
#   0 — PASS: no "Incompatible changes" at all, OR the only incompatible-change
#       bullets are exempt spec-version stamp value-changes; and
#   1 — FAIL: at least one NON-exempt incompatible change remains.
# It is FAIL-CLOSED: if an "Incompatible changes" section is present but no bullet
# items parse out of it, it returns 1 (the output format may have changed — do not
# silently pass an unparsed break).
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

  # Drop the exempt spec-version stamp value-changes; anything left is a real break.
  local remaining
  remaining="$(printf '%s\n' "${incompatible}" | grep -vE "${stamp_exempt_regex}" || true)"
  if [[ -n "$(printf '%s' "${remaining}" | tr -d '[:space:]')" ]]; then
    echo "::error::go-apidiff reports incompatible exported-Go-API change(s) beyond the exempt spec-version stamps. A consumer pinning this module would break; treat as a breaking release (version bump) or revert the API break:" >&2
    printf '%s\n' "${remaining}" >&2
    return 1
  fi
  echo "go-apidiff: the only incompatible change(s) are intentional spec-version stamp bump(s) (VillageAPIVersion/PeasantLocalAPIVersion/TypesVersion); exempted."
  return 0
}

gate_go_apidiff() {
  require_bin go-apidiff
  local base_ref="$1"
  if [[ -z "${base_ref}" ]]; then
    echo ">> go-apidiff: no BASE_REF given; skipping the exported-API breaking diff (nothing prior to compare)."
    return 0
  fi
  echo ">> go-apidiff ${base_ref} (exported Go API of the module vs base)"
  local out
  out="$(go-apidiff "${base_ref}" --repo-path "${repo_root}" 2>&1)" || true
  printf '%s\n' "${out}"
  if printf '%s\n' "${out}" | evaluate_go_apidiff; then
    echo "go-apidiff: no disqualifying incompatible exported-API changes vs ${base_ref}."
    return 0
  fi
  echo "::error::go-apidiff gate failed vs ${base_ref} (see the non-exempt incompatible change(s) above)." >&2
  return 1
}

main() {
  local cmd="${1:-}"
  case "${cmd}" in
    vacuum)     gate_vacuum ;;
    oasdiff)    gate_oasdiff "${2:-}" ;;
    go-apidiff) gate_go_apidiff "${2:-}" ;;
    # evaluate-apidiff reads go-apidiff output on stdin and applies ONLY the pure
    # pass/fail decision (the stamp-exemption filter) — the seam the synthetic-break
    # test drives with real go-apidiff output from a throwaway repo.
    evaluate-apidiff) evaluate_go_apidiff ;;
    all)
      local base_ref="${2:-}" rc=0
      gate_vacuum || rc=1
      gate_oasdiff "${base_ref}" || rc=1
      gate_go_apidiff "${base_ref}" || rc=1
      return "${rc}"
      ;;
    *)
      echo "usage: contract-gates.sh <vacuum|oasdiff <BASE_REF>|go-apidiff <BASE_REF>|evaluate-apidiff|all <BASE_REF>>" >&2
      exit 2
      ;;
  esac
}

main "$@"
