package intent

// This file holds the integration-point stubs: every exported function
// signature the package promises, each returning a six-part error until its
// owning layer lands the logic. The CLI worker builds against these names, so
// no exported identifier is renamed without recording the change first.

// unimplemented reports a function whose logic has not landed yet. It carries
// all six parts so a caller hitting it before its layer lands still learns
// what is missing, why, where it failed, when, what it means, and the fix.
func unimplemented(name string) error {
	return &Error{
		What:  name + " is not implemented yet",
		Why:   "its owning implementation layer has not landed on this branch",
		Where: "internal/release/intent",
		When:  "calling " + name + " before its layer commit",
		Means: "no decision is available; the caller must wait for the layer",
		Fix:   "merge the branch after the owning layer lands, then retry the call",
	}
}

// ParseTitle parses one intent title against the release config and the
// repo's current version.
func ParseTitle(cfg Config, current SemVer, raw string) (Title, error) {
	return Title{}, unimplemented("ParseTitle")
}

// ParseSemVer validates a bare X.Y.Z value.
func ParseSemVer(raw string) (SemVer, error) {
	return "", unimplemented("ParseSemVer")
}

// NextVersion computes the next version from a base, the highest pending
// level, and whether any pending subject is breaking.
func NextVersion(base SemVer, level Level, breaking bool) (SemVer, error) {
	return "", unimplemented("NextVersion")
}

// LoadConfig reads and validates one repo's release-intent file.
func LoadConfig(path string) (Config, error) {
	return Config{}, unimplemented("LoadConfig")
}

// ClassifyPR reports which bot PR, if any, a head branch belongs to.
func ClassifyPR(cfg Config, author, headBranch string) PRKind {
	return PRKindOrdinary
}

// Plan reduces branch facts to the one release action and its values.
func Plan(cfg Config, facts Facts) (Decision, error) {
	return Decision{}, unimplemented("Plan")
}

// Generate renders the release content for a pending set.
func Generate(cfg Config, in ContentInput) (Content, error) {
	return Content{}, unimplemented("Generate")
}

// DecodeQueueResponse decodes a raw GraphQL merge-queue response body.
func DecodeQueueResponse(data []byte) (QueueSnapshot, error) {
	return QueueSnapshot{}, unimplemented("DecodeQueueResponse")
}

// QueueGate applies the hold rule to a decoded queue read.
func QueueGate(cfg Config, current SemVer, entries []QueueEntry, releasePR int) (Gate, error) {
	return "", unimplemented("QueueGate")
}
