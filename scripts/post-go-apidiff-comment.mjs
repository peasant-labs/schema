// Runs from the reviewed default branch. The artifact is untrusted data; only
// the triggering run and GitHub's API can identify the target pull request.
import { lstat, readFile } from "node:fs/promises";
import { join } from "node:path";

const marker = "<!-- go-apidiff-gate -->";
const maxSignalBytes = 256 * 1024;
const maxSectionBytes = 16_000;
const maxPages = 20;
const pageSize = 100;
const repositoryPattern = /^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/;
const shaPattern = /^[0-9a-f]{40}$/;

async function readSignal(directory, name) {
  const file = join(directory, name);
  try {
    const stat = await lstat(file);
    if (!stat.isFile() || stat.size > maxSignalBytes) {
      throw new Error(`${name} is not a bounded regular file`);
    }
    return new TextDecoder("utf-8", { fatal: true }).decode(await readFile(file)).trim();
  } catch (error) {
    if (error.code === "ENOENT") return null;
    throw error;
  }
}

// go-apidiff's collected signals contain one "- symbol: change" per line.
// Parse both fields before rendering; arbitrary artifact Markdown is never used.
function parseEntries(signal) {
  if (!signal) return [];
  const lines = signal.split(/\r?\n/);
  if (lines.length > 1_000) throw new Error("API diff contains too many entries");
  return lines.map((line) => {
    const match = /^\s*-\s+([^:\r\n]{1,512}):[ \t]+([^\r\n]{1,2048})\s*$/u.exec(line);
    if (!match || /[\p{Cc}\p{Cf}]/u.test(match[1] + match[2])) {
      throw new Error("API diff contains an invalid change entry");
    }
    const symbol = match[1].trim();
    const change = match[2].trim();
    if (!symbol || !change) throw new Error("API diff contains an empty change field");
    return { symbol, change };
  });
}

function safeText(text) {
  // Break mentions explicitly, then encode Markdown and HTML punctuation.
  return text.replace(/@/g, "@\u200b").replace(/[&<>"'`*_~[\]()#!|\\]/g,
    (character) => `&#${character.codePointAt(0)};`);
}

function renderEntries(entries) {
  let rendered = "";
  for (const entry of entries) {
    const line = `- ${safeText(entry.symbol)}: ${safeText(entry.change)}\n`;
    if (Buffer.byteLength(rendered + line, "utf8") > maxSectionBytes) {
      return `${rendered}- output truncated\n`;
    }
    rendered += line;
  }
  return rendered;
}

function validRun(event) {
  const run = event?.workflow_run;
  const repository = run?.head_repository?.full_name;
  const branch = run?.head_branch;
  const sha = run?.head_sha;
  if (typeof repository !== "string" || !repositoryPattern.test(repository) ||
      typeof branch !== "string" || !branch || branch.length > 255 ||
      /[\p{Cc}\p{Cf}]/u.test(branch) || typeof sha !== "string" || !shaPattern.test(sha)) {
    return null;
  }
  return { repository, branch, sha };
}

function matchesRun(pull, run) {
  return pull?.state === "open" && pull?.head?.repo?.full_name === run.repository &&
    pull?.head?.ref === run.branch;
}

async function resolvePullRequest(request, api, repository, run) {
  let match = null;
  const url = `${api}/repos/${repository}/commits/${run.sha}/pulls`;
  for (let page = 1; page <= maxPages; page++) {
    const pulls = await request(`${url}?per_page=${pageSize}&page=${page}`);
    if (!Array.isArray(pulls) || pulls.length > pageSize) {
      throw new Error("GitHub returned an invalid pull request list");
    }
    for (const pull of pulls) {
      if (!matchesRun(pull, run)) continue;
      if (match) return null;
      match = pull;
    }
    if (pulls.length < pageSize) {
      if (!match || match.head.sha !== run.sha) return null;
      if (!Number.isSafeInteger(match.number) || match.number <= 0) {
        throw new Error("GitHub returned an invalid pull request number");
      }
      return match.number;
    }
  }
  throw new Error("too many pull request pages to establish a unique match");
}

async function main() {
  const repository = process.env.GITHUB_REPOSITORY ?? "";
  const token = process.env.GH_TOKEN ?? "";
  const directory = process.env.EVIDENCE_DIR ?? "";
  const eventPath = process.env.GITHUB_EVENT_PATH ?? "";
  if (!repositoryPattern.test(repository) || !token || !directory || !eventPath) {
    throw new Error("missing or invalid repository, token, evidence directory, or event path");
  }
  const dirStat = await lstat(directory);
  if (!dirStat.isDirectory()) throw new Error("evidence directory is missing or invalid");
  try {
    await lstat(join(directory, "apidiff-changes.txt"));
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
    console.log("go-apidiff did not establish an outcome; leaving any prior comment untouched");
    return;
  }
  const run = validRun(JSON.parse(await readFile(eventPath, "utf8")));
  if (!run) {
    console.log("workflow run has no valid PR head; leaving comments untouched");
    return;
  }
  const api = (process.env.GITHUB_API_URL ?? "https://api.github.com").replace(/\/$/, "");
  const request = async (url, options = {}) => {
    const response = await fetch(url, {
      ...options,
      signal: AbortSignal.timeout(10_000),
      headers: {
        accept: "application/vnd.github+json",
        authorization: `Bearer ${token}`,
        "content-type": "application/json",
        "x-github-api-version": "2022-11-28",
      },
    });
    if (!response.ok) throw new Error(`GitHub API returned ${response.status}`);
    return response.json();
  };

  const number = await resolvePullRequest(request, api, repository, run);
  if (number === null) {
    console.log("workflow run is stale or has no unique open PR; leaving comments untouched");
    return;
  }
  const changesSignal = await readSignal(directory, "apidiff-changes.txt");
  if (changesSignal === null) {
    console.log("go-apidiff did not establish an outcome; leaving any prior comment untouched");
    return;
  }
  const changes = parseEntries(changesSignal);
  const compatible = parseEntries((await readSignal(directory, "apidiff-compatible.txt")) ?? "");
  const commentsURL = `${api}/repos/${repository}/issues/${number}/comments`;
  let existing;
  for (let page = 1; page <= maxPages; page++) {
    const comments = await request(`${commentsURL}?per_page=${pageSize}&page=${page}`);
    if (!Array.isArray(comments) || comments.length > pageSize) {
      throw new Error("GitHub returned an invalid comments list");
    }
    existing = comments.find((comment) => typeof comment.body === "string" && comment.body.startsWith(marker));
    if (existing || comments.length < pageSize) break;
    if (page === maxPages) throw new Error("too many comment pages to safely find the sticky comment");
  }

  let body = `${marker}\n`;
  if (changes.length) {
    body += "### go-apidiff - incompatible exported-Go-API change(s) vs base\n\n";
    body += "The PR title must declare this break with `![minor]` below 1.0.\n\n";
    body += `${renderEntries(changes)}\n`;
  }
  if (compatible.length) {
    body += "### go-apidiff - compatible (non-breaking) changes vs base\n\n";
    body += "Informational (additive) - no action required.\n\n";
    body += `${renderEntries(compatible)}\n`;
  }
  if (!changes.length && !compatible.length) {
    if (!existing) {
      console.log("go-apidiff is clean and no prior comment exists");
      return;
    }
    body += "### go-apidiff - no incompatible or compatible exported-Go-API changes vs base\n\n";
    body += "The previously-listed change(s) no longer apply to the latest commit on this pull request.";
  }
  if (existing && (!Number.isSafeInteger(existing.id) || existing.id <= 0)) {
    throw new Error("GitHub returned an invalid sticky comment ID");
  }
  // Recheck immediately before writing in case the PR advanced during lookup.
  const current = await request(`${api}/repos/${repository}/pulls/${number}`);
  if (!matchesRun(current, run) || current.head.sha !== run.sha) {
    console.log("pull request advanced; leaving comments untouched");
    return;
  }
  const url = existing ? `${api}/repos/${repository}/issues/comments/${existing.id}` : commentsURL;
  await request(url, { method: existing ? "PATCH" : "POST", body: JSON.stringify({ body: body.trimEnd() }) });
  console.log(existing ? "Updated the go-apidiff comment" : "Created the go-apidiff comment");
}

main().catch((error) => {
  console.error(`go-apidiff comment failed: ${error.message}`);
  process.exitCode = 1;
});
