package release

// CollaboratorPermission is a GitHub repository collaborator permission level —
// the `.permission` field returned by the collaborators API. Typed per the
// repo's no-stringly-typed rule so the maintainer predicate compares against
// named constants, not bare strings.
type CollaboratorPermission string

const (
	PermAdmin    CollaboratorPermission = "admin"
	PermMaintain CollaboratorPermission = "maintain"
	PermWrite    CollaboratorPermission = "write"
	PermTriage   CollaboratorPermission = "triage"
	PermRead     CollaboratorPermission = "read"
	PermNone     CollaboratorPermission = "none"
)

// String renders the permission for CLI output and error messages.
func (p CollaboratorPermission) String() string { return string(p) }

// IsMaintainer reports whether a collaborator permission grants release-
// maintainer authority (may open OR approve a release PR).
//
// This is the SINGLE source of the maintainer predicate. Both release-pr.yml
// gates — the PR-author check and the approver check — consume it via
// `release-guard check-maintainer`, so the {admin, maintain} set is defined
// exactly once (no duplicated `admin|maintain` shell case across jobs).
func IsMaintainer(perm CollaboratorPermission) bool {
	return perm == PermAdmin || perm == PermMaintain
}
