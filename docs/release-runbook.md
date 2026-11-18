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
   `automation/release` PR with a generated changelog section and the npm
   version. The release PR is queued by the bot; do not edit or merge it by
   hand. Before making a tag, the plan job uses its read-only token to require
   exactly one merged PR for the candidate, with the candidate as its merge
   commit, the configured release bot branch as its head, and the releaser App
   as its author. A failed lookup or provenance check makes the shell exit
   nonzero; neither the step nor the plan job sets `continue-on-error`, and
   the release job requires a successful plan. The tag names that release
   PR's merge commit; later commits belong to the next release. The independent
   verifier checks the prepared tag against the recorded candidate and branch
   ancestry without replanning against the current tip.
4. The tag starts `release.yml`: tag verification, Nix vendor-hash freshness,
   and the contract check precede the independent GitHub Release and npm
   publishers. The GitHub Release is created as a draft, the generated JSON
   and YAML specs are uploaded, and only then is it published and protected by
   release immutability. The npm job checks that the committed package version
   matches the tag and removes `private` only in its working copy.

On a rerun, a published release with assets is left unchanged. A failed upload
or publish can leave a draft; a rerun fails if the release is still a draft or
has no assets, and does not resume publication or upload missing assets. Fix
forward through the release process with a new tag. Do not attempt to repair a
published immutable release or reuse its tag.

Tags are append-only. Never push, move, or delete a public tag by hand. Old
`pkg/schema/v*` tags remain intact. The old tag parser still recognizes
historical release candidates, but the automated planner creates final
releases.

Until the spec-freeze migration, content PRs continue to bump `versions.go`
and generate numbered documents themselves. `protected_paths: []` deliberately
does not block those edits. That migration must add protection for
`versions.go` and `generated/*-[0-9]*` when the release bot takes ownership.
The changelog and npm version field are already reserved to the release PR.
The generator currently emits plain `## X.Y.Z` headings; changing that style
is a separate follow-up.

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
  on incomplete releases. Fix forward with a new tag rather than attempting
  to repair or reuse an immutable release.
- The npm version guard reads `typescript/package.json` with `jq`. The dev-shell
  banner polluted the old `nix develop -c node -p` capture; do not reintroduce
  a stdout-captured dev-shell command.
