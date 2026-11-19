#!/usr/bin/env bash
# Copy only numbered, frozen contract documents to a release distribution.
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: collect-release-assets.sh GENERATED_DIR DIST_DIR" >&2
  exit 2
fi

source_dir="$1"
dist_dir="$2"
[[ -d "${source_dir}" ]] || { echo "generated contract directory is missing: ${source_dir}" >&2; exit 1; }
mkdir -p "${dist_dir}"

shopt -s nullglob
copied=0
for path in "${source_dir}"/*; do
  [[ -f "${path}" && ! -L "${path}" ]] || continue
  name="${path##*/}"
  if [[ "${name}" =~ ^[a-z][a-z0-9-]*-[0-9]+\.[0-9]+\.[0-9]+(\.schema)?\.(json|yaml)$ ]]; then
    cp -- "${path}" "${dist_dir}/${name}"
    copied=$((copied + 1))
  fi
done
if [[ "${copied}" -eq 0 ]]; then
  echo "no numbered frozen contracts found in ${source_dir}; refusing an empty release" >&2
  exit 1
fi
# The destination is checked after the copy, so a working file already there
# or copied by a widened pattern is refused before anything is published.
for path in "${dist_dir}"/*-next.*; do
  echo "working contract must not be published: ${path}" >&2
  exit 1
done
