package schema

import "strings"

// RemoteLabel formats a git remote as the canonical "host:owner/repo" display
// label, using the FULL host for every host so self-hosted and enterprise
// forges are distinguishable (github.com:peasant-labs/village,
// gitlab.example.corp:team/repo). There is no host allowlist and no
// short-prefix table: a host that is not a well-known forge keeps its full
// hostname rather than failing closed.
//
// The port, when a remote carries one, is DROPPED: "gitlab.example.corp:2222"
// and "gitlab.example.corp:2200" both render as "gitlab.example.corp". This
// is a deliberate collision, not an oversight — two forges reachable at the
// same host on different ports are indistinguishable in the rendered label
// by design. An IPv6 literal host keeps its brackets with the port removed
// from inside them ("[::1]:2222" -> "[::1]"); the bracket delimiters are
// never mistaken for a port separator. Port stripping applies only to a
// syntactic host component (the URL-scheme and bare-normalized forms); the
// SCP-like form's colon is a path separator, never a port, so nothing there
// is stripped — see the port-handling note below.
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
		// SCP-like SSH form: git@host:owner/repo. The colon here is the
		// mandatory path separator this syntax uses instead of a slash, not
		// a port marker — SCP-like remotes have no port syntax at all. So
		// "git@host:1234/repo" names owner "1234", never port 1234,
		// and the host segment extracted below is never handed to the port
		// stripper: it was already isolated by this split, and a numeric
		// first path segment must survive untouched.
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
	host = stripRemoteLabelPort(host)
	return host + ":" + rest, true
}

// stripRemoteLabelPort drops a trailing ":port" from a syntactic host
// component. host here is always the piece already isolated ahead of the
// first "/" in a URL-scheme or bare-normalized remote — never the SCP-like
// form's owner/repo path, which is split on a different colon before this
// function is reached and so never contains a port to strip.
//
// An IPv6 literal keeps its brackets: "[::1]:2222" -> "[::1]" and the
// already-portless "[::1]" is returned unchanged. Everything else strips at
// the last colon, so "gitlab.example.corp:2222" -> "gitlab.example.corp" and
// a host with no colon at all passes through untouched.
func stripRemoteLabelPort(host string) string {
	if strings.HasPrefix(host, "[") {
		if end := strings.Index(host, "]"); end >= 0 {
			return host[:end+1]
		}
		return host
	}
	if colon := strings.LastIndex(host, ":"); colon >= 0 {
		return host[:colon]
	}
	return host
}
