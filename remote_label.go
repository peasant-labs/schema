package schema

import "strings"

// RemoteLabel formats a git remote as the canonical "host:owner/repo" display
// label, using the FULL host for every host so self-hosted and enterprise
// forges are distinguishable (github.com:peasant-labs/village,
// gitlab.example.corp:team/repo). There is no host allowlist and no
// short-prefix table: a host that is not a well-known forge keeps its full
// hostname rather than failing closed.
//
// RemoteLabel accepts the normalized bare form
// ("host/owner/repo") and, defensively, an HTTPS or SSH remote URL that was
// not pre-normalized ("https://host/owner/repo[.git]",
// "git@host:owner/repo[.git]", "ssh://git@host/owner/repo[.git]"). Any
// userinfo embedded in the remote (user:pass@ or user:TOKEN@ — e.g. a
// personal access token baked into a CI/CD or git-credential-store remote) is
// stripped before the host is extracted, so a credential can never reach a
// rendered label. A trailing ".git" is dropped.
//
// ok is false when remote is empty or does not contain a recognizable
// host + path pair; the caller MUST fall back to another display tier in
// that case.
//
// The signature is deliberately two-valued: there is no RemoteLabelSource
// discriminator, because no consumer of this rule needs one.
func RemoteLabel(remote string) (label string, ok bool) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "", false
	}
	remote = strings.TrimSuffix(remote, ".git")
	remote = strings.TrimRight(remote, "/")

	var hostAndPath string
	switch {
	case strings.Contains(remote, "://"):
		// HTTPS/SSH/git URL scheme: keep everything after "://", then strip
		// embedded userinfo (user:pass@ or user:token@ — e.g. a PAT baked
		// into a CI/CD or git-credential-store remote URL) from the host so
		// a credential is never rendered into a display label.
		after := remote[strings.Index(remote, "://")+3:]
		if at := strings.LastIndex(after, "@"); at >= 0 {
			if slash := strings.Index(after, "/"); slash < 0 || at < slash {
				after = after[at+1:]
			}
		}
		hostAndPath = after
	case strings.Contains(remote, "@") && strings.Contains(remote, ":"):
		// SCP-like SSH form: git@host:owner/repo.
		at := strings.Index(remote, "@")
		colon := strings.Index(remote, ":")
		if colon <= at {
			return "", false
		}
		hostAndPath = remote[at+1:colon] + "/" + remote[colon+1:]
	default:
		// Bare normalized form: host/owner/repo.
		hostAndPath = remote
	}

	host, rest, found := strings.Cut(hostAndPath, "/")
	host = strings.ToLower(strings.TrimSpace(host))
	rest = strings.Trim(rest, "/")
	if !found || host == "" || rest == "" {
		return "", false
	}
	return host + ":" + rest, true
}
