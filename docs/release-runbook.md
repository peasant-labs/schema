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
   the bot; do not edit or merge it by hand. The bot tags only a fresh,
   verified release commit.
4. The tag starts `release.yml`: tag verification, Nix vendor-hash freshness,
   and the contract check precede the independent GitHub Release and npm
   publishers. The GitHub Release carries only numbered frozen JSON and YAML
   specs, including numbered request schemas; `*-next` files never ship as
   release assets.
   The npm job checks that the committed package version matches the tag and
   removes `private` only in its working copy.

Tags are append-only. Never push, move, or delete a public tag by hand. Old
`pkg/schema/v*` tags remain intact. The old tag parser still recognizes
historical release candidates, but the automated planner creates final
releases.

## Working and frozen specs

Types, Village, and Local use committed `generated/<stem>-next.*` working
files. Their numbered files are frozen, and `versions.go` points the public
accessors and exported validators to the latest frozen bytes. The Village
publish-request and annotation-push-request JSON Schemas follow the Village
version. The TypeScript generator, module tests and contract report read
working documents. Segmented read stays on its numbered path until its separate
protocol-version decision.

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
