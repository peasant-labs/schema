#!/usr/bin/env bash
# Recompute flake.nix's MAIN buildGoModule vendorHash (the schema module's own
# third-party graph) and rewrite it in place. Mirrors peasant's script.
#
# Only the FIRST `vendorHash = "...";` in flake.nix is touched — that is the
# schema module's hash (let-binding near the top). The contract-gate derivations
# (oasdiff / go-apidiff) further down carry their OWN vendorHashes and are left
# untouched: this script substitutes a single (non-global) occurrence.
#
# Run after a go.mod/go.sum dependency bump. The nix-vendor-hash CI gate
# (.github/workflows/nix-vendor-hash.yml) runs this on develop; the release-PR
# tagging job runs it before minting a tag (tagged source can't be repaired in
# place).
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
flake="${repo_root}/flake.nix"

# The module vendorHash is the FIRST quoted vendorHash in the file.
current_hash="$(
  sed -nE 's/^[[:space:]]*vendorHash = "([^"]+)";[[:space:]]*$/\1/p' "${flake}" | head -n1
)"

if [[ -z "${current_hash}" ]]; then
  echo "could not find a quoted vendorHash in ${flake}" >&2
  exit 1
fi

restore_current_hash() {
  perl -0pi -e "s/vendorHash = nixpkgs\\.lib\\.fakeHash;/vendorHash = \"${current_hash}\";/" "${flake}"
}

# Substitute only the first occurrence (no /g) so the gate-tool hashes are safe.
perl -0pi -e 's/vendorHash = "[^"]+";/vendorHash = nixpkgs.lib.fakeHash;/' "${flake}"
trap restore_current_hash EXIT

set +e
build_output="$(cd "${repo_root}" && nix build .#default --no-link 2>&1)"
build_status=$?
set -e

new_hash="$(
  printf '%s\n' "${build_output}" |
    sed -nE 's/^[[:space:]]*got:[[:space:]]*(sha256-[A-Za-z0-9+/=]+)[[:space:]]*$/\1/p' |
    tail -n1
)"

if [[ -z "${new_hash}" ]]; then
  printf '%s\n' "${build_output}" >&2
  echo "nix did not report a replacement vendorHash" >&2
  exit "${build_status}"
fi

trap - EXIT
perl -0pi -e "s/vendorHash = nixpkgs\\.lib\\.fakeHash;/vendorHash = \"${new_hash}\";/" "${flake}"

echo "updated vendorHash: ${current_hash} -> ${new_hash}"
