# peasant-labs/schema Auto-land Runbook

Operator guide for the **CI auto-land** that reproduces GitLab's
**squash + merge-commit triangle** on a GitHub PR landing into `develop`
(SLICE-G of epoch [#114](https://github.com/peasant-labs/peasant/issues/114)).

It automates what the manual `~/dotfiles/scripts/git-prmerge` flow does by hand:
squash a PR's commits into ONE commit **S** on `develop`'s tip **T**, then have
GitHub author a `--no-ff` merge commit **M** (parents `[T, S]`) so the PR shows
**MERGED** (not Closed).

```
*   M  Merge branch '<feat>' into 'develop'   (parents [T, S]; --no-ff; Closes #N / See PR #X)
|\
| * S  <PR title>                              (the PR's commits squashed onto T)
|/
*   T  (previous develop tip)
```

> **Status:** built and unit-tested via the git/gh-api subprocess seam +
> `actionlint`. The items in **§5 (Live-verify checklist)** are NOT unit-testable
> and must be proven on the **first real auto-land**.

---

## 1. The mechanism

The triangle is produced by `go run ./cmd/release-guard land` (the YAML in
`.github/workflows/auto-land.yml` is intentionally thin):

1. **Reshape → S.** Fetch `develop` (tip `T`), rebase the PR head onto `T`,
   soft-reset to `T`, create ONE commit `S` whose message is built by
   `release.SquashMessage` (mirrors git-prmerge: PR title + blank + body, then
   `---` / `Squashed commits:` / oldest-first headlines). Force-push `S` to the
   PR branch with `--force-with-lease`.
2. **F1 — poll-to-terminal clean gate.** Forcing `S` makes GitHub recompute
   mergeability asynchronously and re-run required checks on `S`. `land` polls
   the PR's `mergeable_state` until it settles, and merges ONLY on `"clean"`
   **with `S`'s own checks having completed** (it confirms the PR's reported head
   is `S` and that `S`'s check-runs are done — defeating the brief stale
   pre-push `"clean"`). It REFUSES on `blocked`/`dirty`/etc. and FAILS on a
   bounded timeout. It gates on `mergeable_state`, never the conflict-only
   `mergeable` boolean (which is `true` even when `mergeable_state=="blocked"`).
3. **F2 — base-tip recheck.** Immediately before merging, re-confirm
   `develop`'s tip is still `T`. The merge API `sha` param locks only the HEAD
   (`S`), not the base, so on drift (`T → T'`) `land` re-reshapes onto `T'` and
   re-enters the poll (bounded retry).
4. **Merge → M.** `gh api PUT .../pulls/{N}/merge` with `merge_method=merge`,
   `sha=S`, `commit_title="Merge branch '<feat>' into '<target>'"`,
   `commit_message="Closes #N\n\nSee PR #X"`. GitHub authors the `--no-ff` M and
   marks the PR MERGED.

A **known, irreducible micro-TOCTOU** remains between the F2 recheck and the
merge call (GitHub's merge API has no base-SHA compare-and-swap). It is
documented at the merge call site, minimized by the recheck + the
`auto-land-develop` concurrency group + the re-reshape loop, and is **not
fixable** in the client — do not attempt to.

---

## 2. Triggering an auto-land

A **maintainer** triggers it on an **approved, green** PR by either:

- adding the **`auto-merge`** label to the PR, or
- commenting **`/merge`** on the PR.

The trigger is **not** the gate. `land` re-verifies at run time:

- the **triggering actor** is a maintainer (`--actor`, via `check-maintainer`),
- there is a **standing maintainer approval** (via `check-approval`), and
- the PR settles to **`mergeable_state=="clean"`** with `S`'s checks complete.

If any check fails, `land` refuses with an actionable message; on a rebase
conflict it names the conflicting files and instructs running `git prmerge`
locally.

---

## 3. GitHub App configuration (the `peasant-labs-releaser` App)

The merge and the force-push run under the **`peasant-labs-releaser` GitHub App
token** (minted in-workflow via `actions/create-github-app-token`, secrets
`PEASANT_RELEASER_APP_ID` / `PEASANT_RELEASER_APP_PRIVATE_KEY`). This is the same
App `release-pr.yml` already uses; install it additionally on `schema` with:

| Permission | Why |
|---|---|
| **Contents: write** | force-push the squashed commit `S` to the PR branch |
| **Pull requests: write** | call the merge API to author `M` and mark the PR MERGED |

- **NO branch-protection bypass.** Minimal capability is deliberate: an
  approved+green PR already satisfies `develop`'s protection, so the merge API
  merges it without a bypass, while a red/unapproved PR stays blocked by the API.
  The App-token re-trigger property is **independent** of any bypass.
- **Why the App token, not `GITHUB_TOKEN` (LOAD-BEARING):** only an App/PAT-
  initiated event RE-TRIGGERS workflows. A `release(vX.Y.Z-rcN)` PR auto-landed
  this way must still fire `release-pr.yml`'s Trigger B (which mints the tag); a
  `GITHUB_TOKEN` merge would silently skip the release ceremony.

---

## 4. `develop` branch-protection configuration

Required settings on `develop` for the triangle to land:

- **"Require linear history" — MUST be OFF.** It would reject the merge commit
  `M` (a non-linear `--no-ff` merge).
- **"Require signed commits" — MUST be OFF**, *unless* the bot-GPG path is used:
  `S` is machine-produced and unsigned by default (`M` is GitHub web-flow-signed
  by the merge API). If signed commits are ever required on `develop`, pass
  `release-guard land --sign-key <bot-gpg-id>` and provision the bot key; this is
  the one place CI diverges from manual git-prmerge (which signs `S` with the
  user's key — CI cannot hold it).
- Keep **required status checks** and **required reviews** ON — that is what the
  no-bypass App relies on for safety.

---

## 5. Live-verify checklist (prove on the first real auto-land)

These are **not** unit-testable (they need a real GitHub run); verify them on the
first auto-land:

- [ ] The App-token-initiated merge **RE-TRIGGERS** workflows.
- [ ] The PR ends **MERGED** (not Closed), and `M`'s parents are `[T, S]`.
- [ ] A `release(vX.Y.Z-rcN)` PR auto-landed this way **fires `release-pr.yml`
      Trigger B** → the tag is minted.
- [ ] The App permissions (Contents:write + Pull-requests:write, **no bypass**)
      and the `develop` config (**no** require-linear-history; **no**
      require-signed-commits unless the `--sign-key` path is enabled) are set.
