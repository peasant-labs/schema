package intent

// ClassifyPR reports which bot PR, if any, a head branch belongs to. Only a
// PR by the automation author on that kind's listed branch counts: any other
// author, any unlisted branch, and any empty input is an ordinary PR. The
// protected-path exemption applies to the release PR only because only it
// classifies as one.
func ClassifyPR(cfg Config, author, headBranch string) PRKind {
	if author == "" || author != cfg.AutomationAuthor || headBranch == "" {
		return PRKindOrdinary
	}
	for _, kind := range botKinds {
		if branch, ok := cfg.BotBranches[kind]; ok && branch != "" && headBranch == branch {
			return kind
		}
	}
	return PRKindOrdinary
}
