# Schema release runbook

The schema module ships the Go contract, generated OpenAPI and JSON Schema
assets, and the `@peasant-labs/schema` npm package. It ships no binary.
Consumers re-pin only after the schema tag exists.

## Release path

1. A content PR targets `develop` with a title such as
   `feat(local-api)[minor]: add an endpoint`. A breaking change below 1.0 must
   use `![minor]`. `docs`, `test`, `ci`, `chore`, and `style` never carry a
   level. The title guard checks the live title on PR and merge-group events.
2. `tests.yml` lints the specs, checks numbered-file immutability, builds the
   Go/OpenAPI contract report, and checks the live title against that report.
   The merge group compares the entry PR's own diff, not the combined group.
3. After landing, `release-batch.yml` plans from the last final tag. When tag
   mode is on, it checks the landed range and maintains the single
   `automation/release` PR with a generated changelog section, the npm
   version, and the freeze of every ahead supported surface. The tool's title
   is passed unchanged to verification and the PR. The release PR is queued by
   the bot; do not edit or merge it by hand. Before making a tag, the plan job
   uses its read-only token to require exactly one merged PR for the
   candidate, with the candidate as its merge commit, the configured release
   bot branch as its head, and the releaser App as its author. A failed lookup or provenance check makes the shell exit
   nonzero; neither the step nor the plan job sets `continue-on-error`, and
   the release job requires a successful plan. The tag names that release
   PR's merge commit; later commits belong to the next release. The independent
   verifier checks the prepared tag against the recorded candidate and branch
   ancestry without replanning against the current tip. If the run's event
   commit predates the candidate, the ancestry check fails closed; a later
   default-branch run can retry.
4. The tag starts `release.yml`: tag verification, Nix vendor-hash freshness,
   and the contract check precede the independent GitHub Release and npm
   publishers. `scripts/collect-release-assets.sh` copies only the numbered
   frozen JSON and YAML specs, including numbered request schemas, into
   `dist/assets`; `*-next` files never ship as release assets. The release notes
   sit beside the assets at `dist/release-notes.md` and are not uploaded.
   The `release` job checks out and builds the pinned `release-intent` tool
   before downloading `release-dist`, then runs `release-intent publish-release
   --require-immutable`. It builds the local manifest first, lists every
   release (30 per page, drafts included, failing closed if truncated or if any
   draft exists for the tag), confirms `refs/tags/<tag>` exists, creates a draft
   release and uploads every asset by release id, verifies the draft's assets by
   name, size, and SHA-256 digest against the manifest, publishes it by id, and
   verifies the published release again (including `immutable`). The npm job
   checks that the committed package version matches the tag and removes
   `private` only in its working copy.

On 2026-11-18, a read-only check of GitHub's commit-to-pulls API returned the
merged bot release PR, its merge commit, `automation/release` head, and
releaser App author for the merge-queue squash commit `cf9f3b7189d00a1ac841a75c413299d15c6fa257`.
The same check for ordinary merge commit `24323466e020c223cf9750bd0606774c104b4b76`
returned a human-authored PR. This is live evidence, not a guarantee of future
API behavior: the provenance test stubs Git, yq, and gh.

A transient API error (rate limit, 5xx, authentication) before creation fails
the run before anything is created: rerun the workflow; no new tag is needed.
If `draft-exists` is reported, nothing was created by this run; delete stale
drafts by id as below, then rerun.

On a rerun, a published release matching the exact manifest, prerelease flag,
and immutability requirement is left unchanged (`state=already-published`).
Because GitHub's REST API cannot enforce draft-only deletion, real release runs
never send automatic `DELETE` requests: on a pre-publish failure or ambiguous
publish, `publish-release` reads the release back by id and leaves it
untouched, printing either the by-id inspect/delete command (if still a draft)
or fix-forward guidance (if published). Drafts are not immutable. Delete every
draft for the tag **by its id**, then rerun the release workflow for the same
tag. Set `TAG` to the existing tag and list the drafts:

```sh
gh api repos/peasant-labs/schema/releases --paginate \
  --jq ".[] | select(.draft and .tag_name == \"$TAG\") | {id, assets: (.assets | length), created_at}"
gh api -X DELETE repos/peasant-labs/schema/releases/<id>
```

Deleting a draft that holds a few hundred assets can answer HTTP 502 and still
finish a little later; read the release back by id and retry only while it
still exists.

Never use `gh release delete "$TAG"`: when a tag has several drafts, the tag
lookup returns one of them (the oldest), so the command can delete the wrong
one, and `--cleanup-tag` would delete the tag. Do not create a new tag to
recover a draft. If an already-published release has no assets, fix forward
with a new tag; do not attempt to repair a published immutable release or
reuse its tag.

Why `publish-release`: v0.55.0 through v0.57.1 were published with no assets.
In run 44295390680, `gh release create --draft` made one draft at 11:37:06; a
separate upload action, which finds a release only by listing releases, did not
see it, made a second draft at 11:37:13 and uploaded 349 assets there; at
11:37:16 `gh release edit --draft=false` published the first, empty draft,
because GitHub's tag lookup returns the oldest draft. v0.54.0, the last good
release before adopting `publish-release`, used one `gh release create` with
its assets; `publish-release` now creates, uploads, verifies by manifest and
SHA-256 digest, and publishes strictly by release id.

Tags are append-only. Never push, move, or delete a public tag by hand. Old
`pkg/schema/v*` tags remain intact. The old tag parser still recognizes
historical release candidates, but the automated planner creates final
releases.

## Working and frozen specs

Types, Village, Local, and segmented read use committed
`generated/<stem>-next.*` working files. Their numbered files are frozen, and
`versions.go` points the public accessors and exported validators to the
latest frozen bytes. The Village publish-request and annotation-push-request
JSON Schemas follow the Village version. The TypeScript generator, module
tests and contract report read working documents. For segmented read,
`SegmentedReadSpecVersion` names the frozen document, while
`SegmentedReadAPIVersion` names the current wire protocol and
`SegmentedReadMinVersion` the oldest served protocol. A document freeze moves
only the spec version; it does not change the selector table, the minimum
document, or a protocol echo. A breaking protocol change requiring a new `v=`
generation is separate P1 work, not a document freeze.

`schema-gen status --write state.yaml` computes `clean`, `docs-ahead`, or
`ahead`. Description-only changes remain docs-ahead and are copied at the next
functional freeze. `release-intent freeze-plan` chooses each ahead surface's
level from landed titles after the spec-freeze migration boundary, while the
ordinary release floor and the full module changelog range remain independent.
`schema-gen freeze --plan plan.yaml --state state.yaml` checks the handoff,
updates constants and mirrors, rebuilds the generator, and writes numbered
copies and registry rows. The release PR's merge-group tree must pass
`schema-gen freeze --check`. Before tagging, the candidate is also checked
against a plan made from its parent state.

A functional surface change requires a releasing title naming it; an
unchanged or docs-only surface may be named only if it is already ahead on the
base. `spec-change-needs-bump` and `surface-mismatch` identify these errors.
A non-releasing `docs(local-api): clarify ...` is valid for docs-only work.

At cutover, an open content PR that edited `versions.go` or numbered specs
merges the updated base, removes those edits, regenerates working files, and
lists every functionally changed surface in its title. If a wrong stale freeze
bypasses the queue check, pre-tag verification rejects it; the next release
freezes the next number and retains the wrong numbered file in the registry.
That burns one spec number without moving a tag.

The release job takes the migration boundary (the freeze floor) from the
first first-parent `develop` commit whose `.github/release-intent.yaml`
gained the literal line `spec_freeze: true`. Later edits to that line do not
move it. If no commit ever wrote the line literally, the step fails instead
of guessing; `TestFreezeFloorSearch` runs the step against fixture histories.

The changelog and npm version field are reserved to the release PR. The
generator currently emits plain `## X.Y.Z` headings; changing that style is a
separate follow-up.

## Release PR checks

A release PR (or a merge group) whose whole `base..head` change is
`CHANGELOG.md` plus the `typescript/package.json` version does not rerun
`make check`. The path filter alone cannot skip it: a version-only edit under
`typescript/*` classifies as `code=true`, so inheritance is what saves the
run. Only the bot release PR, or a merge group whose entry is that PR, runs
`inherit-verdict`; every other run skips it without building the tool. The
`changes` job runs `release-intent inherit-verdict`, which reads the
`merge-gate` verdict that the exact base commit's merge-queue run recorded.
When that verdict is missing, pending past the 20-minute wait, failed,
cancelled, ambiguous or from another producer, the job fails and the release
candidate is not merged; it never falls back to the suite. Fix it by waiting
for or rerunning the base commit's merge-queue run, then rerun the candidate's
checks. The contract gates, the title guard, the release tree freeze check and
the release group's vendor-hash check still run. Spec-freeze outputs
(`versions.go`, numbered `generated/` documents) are never release-only and
always run the full suite.

`merge-gate` ends with one `verdict-source: full|inherited|unaffected` step
per run, so a later release candidate can tell a verdict that rests on a
`make check` run from an inherited one. `make check` runs without the race
detector on every event; the merge-queue run adds `make race-selective`, the
race pass over the validator schema caches.

## Vendor hash

After a dependency-affecting push to `develop`, `nix-vendor-hash.yml` computes
the hash without credentials, confirms that the only change is one
`vendorHash` line, and hands a Git bundle to a separate credentialed job. That
job pushes only `automation/vendor-hash`, opens or updates its bot PR, and
enables auto-merge. If no change is needed, it closes an open hash PR. It never
pushes `develop`. This workflow is independent of `RELEASE_BATCH_MODE`: even
with tag mode off it can open, update, arm, or close a vendor-hash PR.

On ordinary PRs the hash check is informational. The release PR's merge-group
requires a current hash, and the tag-time check is the final backstop. A stale
release group must wait for the vendor-hash PR; never repair a tagged tree.

## Credentials and cutover

The releaser App must be installed on `peasant-labs/schema` with Contents and
Pull requests write. Existing repository secrets
`PEASANT_RELEASER_APP_ID` and `PEASANT_RELEASER_APP_PRIVATE_KEY` mint its
token. `RELEASE_BATCH_MODE` is unset for plan-only operation; the maintainer
sets it to `tag` only after the plan-only checks below. `RELEASE_BATCH_FLOOR`
is normally unset; it is a recovery boundary, not a release version.
The App must be allowed to push its `automation/*` branches, including the
force-with-lease updates, under the applicable branch rules. That permission
does not bypass the merge queue or permit a push to `develop`.

The reviewed go-apidiff comment job uses the existing organization secret
`JOURNEY_GITHUB_USER_PAT`. Confirm that its owner can read pull requests and
create and update the existing sticky comment, including comments authored by
the earlier bot. The `workflow_run` job downloads the triggering run's artifact
with the read-only workflow token, then runs a default-branch script with the
PAT. It resolves the open PR from the run's head SHA, repository and branch;
the untrusted artifact cannot select a PR or supply Markdown to post. This
trusted-event separation also covers fork and Dependabot PRs.

Before merging the adoption, the maintainer should:

1. Confirm squash merge commits use the PR title, and auto-merge is enabled.
   Retitle or drain open PRs using the old release-title grammar.
2. Ensure the existing credentials and npm Trusted Publisher are available.
   Verify the sticky-comment update on an existing bot-authored comment and
   on a fork PR after the workflow is on the default branch.
3. Configure and compare the `Merge queue` ruleset with
    `.github/release-queue.policy.yml`; protect release tags so only the
    releaser App creates them. Confirm the queue ruleset coexists with the
    existing `develop` and `releases/**/*` rules. If the effective required
    contexts or rules differ, stop: do not enqueue a release PR or enable tag
    mode until the maintainer resolves and rechecks the rulesets.
4. Leave tag mode off for the first group. Inspect the plan and independently
   read back that no release branch, PR, or tag was written. Check the group
   SHA and queue behavior before the maintainer enables tag mode.

The former release-PR maintainer approval assertion was disabled for the
single-maintainer setup. Its old workflow is removed; restoring a mandatory
second approval requires a reviewed replacement compatible with queued bot
PRs and branch rules, not uncommenting a step in a deleted workflow.

## npm Trusted Publishing

The `npm-publish` job uses OIDC, not an npm token. The GitHub environment must
be named `npm-publish`. Register the npm Trusted Publisher for organization
`peasant-labs`, repository `schema`, workflow `release.yml`, and environment
`npm-publish`, scoped to `npm publish`. Optional environment protections
require maintainer review before enabling them.

The job installs the locked TypeScript tooling, checks tag and committed
manifest versions, removes `private` in its temporary checkout, and runs
typecheck, tests, package audit and package smoke before publishing. Historical
release candidates use `next`; final tags use `latest`. A registry replica may
take minutes to show a successful publication.

If OIDC authentication fails, verify all four Trusted Publisher fields and
the environment, then rerun the failed job. If the tag and committed package
version differ, the immutable tag has burned that version: fix forward with a
new release. An npm `E409` means the version was already published; never
attempt to overwrite it. Failure of one publisher does not block its sibling.

## Local checks and recovery

Run `nix develop -c make check` and `nix develop -c make gates
BASE_REF=origin/develop`. The contract report compares committed `HEAD` with
the base, so commit the changes first. With a declared break, put the pinned
`release-intent` on `PATH` and pass
`TITLE='feat(local-api)![minor]: describe the change'` to `make gates`.
Without a title the local gate fails conservatively on a reported break.

For the real workflow policy gate, build the release tool at the SHA in the
workflows and run:

```sh
release-intent check-config --config .github/release-intent.yaml
release-intent check-workflows --policy .github/release-queue.policy.yml --dir .github/workflows
```

`--print-required` is for comparing contexts with the GitHub ruleset, not a
substitute for the workflow check. A stale release candidate is rebuilt by the
bot, never tagged. If a landed subject is unparseable, an operator can set
`RELEASE_BATCH_FLOOR` to that commit after deciding how to account for the
excluded history. A forced release uses the `cut-release` dispatch or a PR
containing only an empty commit titled `release[patch]: cut a release`; do not
push a tag manually. The tag-triggered contract job and the pre-tag check use
the same floor and range.

If the release job fails to push a tag, inspect the job and remote tag first.
After fixing the cause, dispatch `release-batch` on the default branch so it
replans and verifies the current candidate. If the tag already exists, do not
move or recreate it; investigate its target and the tag-triggered run instead.

Release recovery rules:

- The `Protect release tags` ruleset has exactly one bypass actor: the
  `peasant-labs-releaser` GitHub App (Integration id 3988034). A human
  `git push` of a release tag is refused. Fix a failed batch; do not change the
  ruleset to work around it.
- `RELEASE_BATCH_MODE` must be `tag` for the batch to write. When unset, the
  batch runs plan-only and returns `action=none` for pending releases. After any
  runner, permission, or release-config change, confirm the variable is `tag`
  with `gh variable list -R peasant-labs/schema` before treating a plan as
  authoritative.
- In the 0.56.0 incident, the batch returned `action=none` because
  `RELEASE_BATCH_MODE` was unset; that alone prevented a write. A later
  non-releasing fix also moved the tip after the release PR merged. Do not
  treat `none` as success or move a tag; set the mode to `tag`, then cut a new
  release that includes the fix.
- An immutable GitHub Release cannot be updated. Older publisher runs failed
  with `tag_name was used by an immutable release` or `target_commitish cannot
  be changed`. The current publisher leaves complete releases alone and fails
  on incomplete releases. Delete an unpublished draft by its id as described above; for
  an already-published release without assets, fix forward with a new tag.
- The npm version guard reads `typescript/package.json` with `jq`. The dev-shell
  banner polluted the old `nix develop -c node -p` capture; do not reintroduce
  a stdout-captured dev-shell command.
