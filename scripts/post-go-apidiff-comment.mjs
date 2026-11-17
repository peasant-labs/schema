// This script runs from the reviewed default branch. The downloaded API-diff
// files are untrusted data, never code, and the PR number comes from the event.
import { lstat, readFile } from "node:fs/promises";
import { join } from "node:path";

const marker = "<!-- go-apidiff-gate -->";
const maxSignalBytes = 256 * 1024;
const maxCommentChars = 24_000;

async function readSignal(name) {
  const file = join(process.env.EVIDENCE_DIR, name);
  try {
    const stat = await lstat(file);
    if (!stat.isFile() || stat.size > maxSignalBytes) {
      throw new Error(`${name} is not a bounded regular file`);
    }
    return (await readFile(file, "utf8")).trim();
  } catch (error) {
    if (error.code === "ENOENT") return null;
    throw error;
  }
}

function fenced(text) {
  const limited = text.length > maxCommentChars
    ? `${text.slice(0, maxCommentChars)}\n... output truncated`
    : text;
  const longest = Math.max(2, ...Array.from(limited.matchAll(/`+/g), (m) => m[0].length));
  const fence = "`".repeat(longest + 1);
  return `${fence}\n${limited}\n${fence}\n\n`;
}

async function main() {
  const repository = process.env.GITHUB_REPOSITORY ?? "";
  const number = process.env.PR_NUMBER ?? "";
  const token = process.env.GH_TOKEN ?? "";
  if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(repository) ||
      !/^[1-9][0-9]{0,8}$/.test(number) || !token || !process.env.EVIDENCE_DIR) {
    throw new Error("missing or invalid repository, triggering PR number, token, or evidence directory");
  }

  const changes = await readSignal("apidiff-changes.txt");
  if (changes === null) {
    console.log("go-apidiff did not establish an outcome; leaving any prior comment untouched");
    return;
  }
  const compatible = (await readSignal("apidiff-compatible.txt")) ?? "";
  const api = (process.env.GITHUB_API_URL ?? "https://api.github.com").replace(/\/$/, "");
  const commentsURL = `${api}/repos/${repository}/issues/${number}/comments`;
  const request = async (url, options = {}) => {
    const response = await fetch(url, {
      ...options,
      headers: {
        accept: "application/vnd.github+json",
        authorization: `Bearer ${token}`,
        "content-type": "application/json",
        "x-github-api-version": "2022-11-28",
      },
    });
    if (!response.ok) throw new Error(`GitHub comments API returned ${response.status}`);
    return response.json();
  };

  let existing;
  for (let page = 1; page <= 100; page++) {
    const comments = await request(`${commentsURL}?per_page=100&page=${page}`);
    if (!Array.isArray(comments)) throw new Error("GitHub returned an invalid comments list");
    existing = comments.find((comment) => typeof comment.body === "string" && comment.body.startsWith(marker));
    if (existing || comments.length < 100) break;
    if (page === 100) throw new Error("too many comment pages to safely find the sticky comment");
  }

  let body = `${marker}\n`;
  if (changes) {
    body += "### go-apidiff - incompatible exported-Go-API change(s) vs base\n\n";
    body += "The PR title must declare this break with `![minor]` below 1.0.\n\n";
    body += fenced(changes);
  }
  if (compatible) {
    body += "### go-apidiff - compatible (non-breaking) changes vs base\n\n";
    body += "Informational (additive) - no action required.\n\n";
    body += fenced(compatible);
  }
  if (!changes && !compatible) {
    if (!existing) {
      console.log("go-apidiff is clean and no prior comment exists");
      return;
    }
    body += "### go-apidiff - no incompatible or compatible exported-Go-API changes vs base\n\n";
    body += "The previously-listed change(s) no longer apply to the latest commit on this pull request.";
  }
  const url = existing
    ? `${api}/repos/${repository}/issues/comments/${existing.id}`
    : commentsURL;
  if (existing && (!Number.isSafeInteger(existing.id) || existing.id <= 0)) {
    throw new Error("GitHub returned an invalid sticky comment ID");
  }
  await request(url, { method: existing ? "PATCH" : "POST", body: JSON.stringify({ body: body.trimEnd() }) });
  console.log(existing ? "Updated the go-apidiff comment" : "Created the go-apidiff comment");
}

main().catch((error) => {
  console.error(`go-apidiff comment failed: ${error.message}`);
  process.exitCode = 1;
});
