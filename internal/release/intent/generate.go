package intent

import (
	"regexp"
	"strings"
)

// This file holds the release-content generator: the pure function from the
// pending releasing subjects, the base changelog text, and the release config
// to the new changelog section and the version-field values.
//
// The section heading follows the repo's existing style, read from the base
// text's latest section: a bracketed heading keeps brackets and the date
// rule, a bare heading keeps the bare version. Nothing in the output depends
// on when it is generated: a dated heading takes the date the caller passes,
// which is the commit date of the newest pending subject.

var (
	bracketHeading = regexp.MustCompile(`^## \[([^\]]+)\]( - (.+))?$`)
	bareHeading    = regexp.MustCompile(`^## (.+)$`)
)

// Generate renders the release content for a pending set. The pending titles
// are the planner's, oldest commit first in the section lines. A release
// commit adds no line because the planner leaves it out of pending.
func Generate(cfg Config, in ContentInput) (Content, error) {
	if len(in.Pending) == 0 {
		return Content{}, inputError(
			"nothing pending for the release content",
			"the generator renders the pending releasing subjects",
			"intent.Generate",
			"rendering the release content",
			"no section and no version values can be rendered from an empty set",
			"call Generate only for an open-pr plan, with the planner's pending set",
		)
	}
	if in.Tag == "" || in.Version == "" {
		return Content{}, inputError(
			"the release content needs a version and a tag",
			"the section heading and the version fields both carry the computed number",
			"intent.Generate",
			"rendering the release content",
			"no section and no version values can be rendered without the number",
			"pass the planner's version and tag for this pending set",
		)
	}

	ordered := append([]Title(nil), in.Pending...)
	for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
		ordered[i], ordered[j] = ordered[j], ordered[i]
	}
	lines := make([]string, 0, len(ordered))
	for _, title := range ordered {
		line := "- " + title.Subject
		if title.PR != 0 {
			line += " (#" + itoa(title.PR) + ")"
		}
		if title.Breaking {
			line += " (BREAKING)"
		}
		lines = append(lines, line)
	}

	heading, err := changelogHeading(in)
	if err != nil {
		return Content{}, err
	}
	section := heading + "\n\n" + strings.Join(lines, "\n") + "\n"
	changelog := placeSection(in.BaseChangelog, heading, lines)

	versions := make(map[string]string, len(cfg.VersionFiles))
	for _, vf := range cfg.VersionFiles {
		versions[vf.Path] = string(in.Version)
	}
	return Content{
		Changelog: changelog,
		Section:   section,
		Versions:  versions,
	}, nil
}

// changelogHeading builds the section heading in the base text's style.
func changelogHeading(in ContentInput) (string, error) {
	bracket, dated := sniffHeadingStyle(in.BaseChangelog)
	if bracket {
		if dated || in.Date != "" {
			if in.Date == "" {
				return "", inputError(
					"the changelog style carries a date but no date was given",
					"a dated heading takes the commit date of the newest pending subject",
					"intent.Generate",
					"rendering the section heading",
					"no section can be rendered without the date its heading carries",
					"pass the newest pending subject's commit date",
				)
			}
			return "## [" + in.Tag + "] - " + in.Date, nil
		}
		return "## [" + in.Tag + "]", nil
	}
	return "## " + string(in.Version), nil
}

// sniffHeadingStyle reads the base text's latest section heading: bracketed
// or bare, and whether it carries a date. A base with no section defaults to
// the bracketed style.
func sniffHeadingStyle(base string) (bracket, dated bool) {
	for _, line := range strings.Split(base, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "## ") {
			continue
		}
		if m := bracketHeading.FindStringSubmatch(trimmed); m != nil {
			return true, m[3] != ""
		}
		return false, false
	}
	return true, false
}

// placeSection merges the new section into the base text. An [Unreleased]
// block keeps its entries above the generated lines under the new heading;
// otherwise the section goes before the first section. A base with no
// section gains the new one after its preamble.
func placeSection(base, heading string, lines []string) string {
	ls := strings.Split(base, "\n")
	first := -1
	for i, line := range ls {
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			first = i
			break
		}
	}
	if first < 0 {
		return strings.TrimRight(base, "\n") + "\n\n" + heading + "\n\n" + strings.Join(lines, "\n") + "\n"
	}
	name := strings.TrimSpace(ls[first][len("## "):])
	name = strings.Trim(name, "[]")
	if strings.EqualFold(name, "unreleased") {
		end := len(ls)
		for i := first + 1; i < len(ls); i++ {
			if strings.HasPrefix(strings.TrimSpace(ls[i]), "## ") {
				end = i
				break
			}
		}
		body := strings.Trim(strings.Join(ls[first+1:end], "\n"), "\n")
		rest := strings.Join(ls[end:], "\n")
		out := strings.Join(ls[:first], "\n")
		out = strings.TrimRight(out, "\n") + "\n\n" + heading + "\n"
		if body != "" {
			out += "\n" + body + "\n"
		}
		out += "\n" + strings.Join(lines, "\n") + "\n"
		if strings.TrimSpace(rest) != "" {
			out += "\n" + strings.TrimLeft(rest, "\n")
		}
		return out
	}
	head := strings.TrimRight(strings.Join(ls[:first], "\n"), "\n")
	tail := strings.Join(ls[first:], "\n")
	return head + "\n\n" + heading + "\n\n" + strings.Join(lines, "\n") + "\n\n" + strings.TrimLeft(tail, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
