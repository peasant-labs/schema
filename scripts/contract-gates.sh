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
#              fail if any incompatible change is reported.
#
# Usage:
#   contract-gates.sh vacuum
#   contract-gates.sh oasdiff   <BASE_REF>
#   contract-gates.sh go-apidiff <BASE_REF>
#   contract-gates.sh all       <BASE_REF>
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
  if printf '%s\n' "${out}" | grep -qiE 'Incompatible changes'; then
    echo "::error::go-apidiff reports incompatible exported-Go-API changes vs ${base_ref}. A consumer pinning this module would break; treat as a breaking release (version bump) or revert the API break." >&2
    return 1
  fi
  echo "go-apidiff: no incompatible exported-API changes vs ${base_ref}."
  return 0
}

main() {
  local cmd="${1:-}"
  case "${cmd}" in
    vacuum)     gate_vacuum ;;
    oasdiff)    gate_oasdiff "${2:-}" ;;
    go-apidiff) gate_go_apidiff "${2:-}" ;;
    all)
      local base_ref="${2:-}" rc=0
      gate_vacuum || rc=1
      gate_oasdiff "${base_ref}" || rc=1
      gate_go_apidiff "${base_ref}" || rc=1
      return "${rc}"
      ;;
    *)
      echo "usage: contract-gates.sh <vacuum|oasdiff <BASE_REF>|go-apidiff <BASE_REF>|all <BASE_REF>>" >&2
      exit 2
      ;;
  esac
}

main "$@"
