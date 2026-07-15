package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	schema "github.com/peasant-labs/schema"
)

type typeScriptEnum struct {
	Name    string
	AllName string
	Members []typeScriptEnumMember
	All     []string
}

type typeScriptEnumMember struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func typeScriptEnums() ([]typeScriptEnum, error) {
	descriptors := []struct {
		name, allName string
		values        []string
		all           []string
	}{
		{"AnnotationAxis", "AllAnnotationAxes", stringsOf(schema.AxisType, schema.AxisSession, schema.AxisProject), nil},
		{"AnnotationDatatype", "AllAnnotationDatatypes", stringsOfSlice(schema.AllAnnotationDatatypes), nil},
		{"AnnotationPushStatus", "AllAnnotationPushStatuses", stringsOf(schema.PushStatusCreated, schema.PushStatusUpdated, schema.PushStatusSkipped, schema.PushStatusError), nil},
		{"AnnotationStatus", "AllAnnotationStatuses", stringsOfSlice(schema.AllAnnotationStatuses), nil},
		{"AnnotatorKind", "AllAnnotatorKinds", stringsOfSlice(schema.AllAnnotatorKinds), nil},
		{"BuiltinCommand", "AllClaudeBuiltinCmds", stringsOfSlice(schema.AllClaudeBuiltinCmds[:]), nil},
		{"ChangeBinding", "AllChangeBindings", stringsOf(schema.ChangeBindingBound, schema.ChangeBindingCandidate), nil},
		{"ChannelTopic", "AllChannelTopics", stringsOf(schema.TopicDashboard, schema.TopicSessions, schema.TopicSessionDetail, schema.TopicTrends, schema.TopicQuality, schema.TopicAnnotations, schema.TopicProjectFamiliarity), nil},
		{"ContentKind", "AllContentKinds", stringsOf(schema.ContentKindSessionDetail), nil},
		{"DecayLevel", "AllDecayLevels", stringsOf(schema.DecayFresh, schema.DecayFading, schema.DecayStale, schema.DecayUnexplored), nil},
		{"EdgeViolationKind", "AllEdgeViolationKinds", stringsOf(schema.EdgeViolationCycle, schema.EdgeViolationWrongWay), nil},
		{"EntryType", "AllEntryTypes", stringsOfSlice(schema.AllEntryTypes), nil},
		{"Harness", "AllHarnesses", stringsOfSlice(schema.Harnesses()), stringsOfSlice(schema.AllHarnesses)},
		{"InteractionType", "AllInteractionTypes", stringsOf(schema.InteractionMentioned, schema.InteractionRead, schema.InteractionDiscussed, schema.InteractionQuestioned), nil},
		{"License", "AllLicenses", stringsOfSlice(schema.AllLicenses), nil},
		{"MapNodeKind", "AllMapNodeKinds", stringsOf(schema.MapNodeKindModule, schema.MapNodeKindPackage, schema.MapNodeKindFile), nil},
		{"MessageType", "AllMessageTypes", stringsOf(schema.MsgSubscribe, schema.MsgUnsubscribe, schema.MsgDashboard, schema.MsgSessions, schema.MsgSessionDetail, schema.MsgTrends, schema.MsgQuality, schema.MsgAnnotations, schema.MsgProjectFamiliarity, schema.MsgConnected, schema.MsgError), nil},
		{"Role", "AllRoles", stringsOfSlice(schema.AllRoles), nil},
		{"ScaleKind", "AllScaleKinds", stringsOfSlice(schema.AllScaleKinds), nil},
		{"SessionOutcome", "AllOutcomes", stringsOfSlice(schema.AllOutcomes), nil},
		{"SourceFormat", "AllSourceFormats", stringsOf(schema.SourceFormatJSONL, schema.SourceFormatJSON), nil},
		{"StopReason", "AllStopReasons", stringsOfSlice(schema.AllStopReasons), nil},
		{"TargetKind", "AllTargetKinds", stringsOfSlice(schema.AllTargetKinds), nil},
		{"ToolCallKind", "AllToolCallKinds", stringsOfSlice(schema.AllToolCallKinds), nil},
		{"TypeOrigin", "AllTypeOrigins", stringsOfSlice(schema.AllTypeOrigins), nil},
		{"ValueDomainKind", "AllValueDomainKinds", stringsOfSlice(schema.AllValueDomainKinds), nil},
		{"Visibility", "AllVisibilities", stringsOfSlice(schema.AllVisibilities), nil},
	}

	out := make([]typeScriptEnum, 0, len(descriptors))
	for _, descriptor := range descriptors {
		members := make([]typeScriptEnumMember, 0, len(descriptor.values))
		seenNames := map[string]string{}
		seenValues := map[string]struct{}{}
		for _, value := range descriptor.values {
			if _, duplicate := seenValues[value]; duplicate {
				return nil, fmt.Errorf("TypeScript enum %s contains duplicate Go value %q", descriptor.name, value)
			}
			seenValues[value] = struct{}{}
			name := enumMemberName(descriptor.name, value)
			if prior, collision := seenNames[name]; collision {
				return nil, fmt.Errorf("TypeScript enum %s member-name collision: %q and %q both normalize to %s", descriptor.name, prior, value, name)
			}
			seenNames[name] = value
			members = append(members, typeScriptEnumMember{Name: name, Value: value})
		}
		all := descriptor.all
		if all == nil {
			all = append([]string(nil), descriptor.values...)
		}
		for _, value := range all {
			if _, known := seenValues[value]; !known {
				return nil, fmt.Errorf("TypeScript enum %s collection %s contains unknown Go value %q", descriptor.name, descriptor.allName, value)
			}
		}
		out = append(out, typeScriptEnum{Name: descriptor.name, AllName: descriptor.allName, Members: members, All: all})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func stringsOf[T ~string](values ...T) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}
	return out
}

func stringsOfSlice[T ~string](values []T) []string { return stringsOf(values...) }

var enumWordBoundary = regexp.MustCompile(`[^A-Za-z0-9]+`)

func enumMemberName(enumName, value string) string {
	overrides := map[string]map[string]string{
		"Harness":      {"gemini-cli": "GeminiCLI", "opencode": "OpenCode"},
		"License":      {"CC0-1.0": "CC0", "CC-BY-4.0": "CCBY", "CC-BY-SA-4.0": "CCBYSA"},
		"SourceFormat": {"jsonl": "JSONL", "json": "JSON"},
	}
	if name := overrides[enumName][value]; name != "" {
		return name
	}
	parts := enumWordBoundary.Split(value, -1)
	var out strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		out.WriteRune(unicode.ToUpper(runes[0]))
		out.WriteString(string(runes[1:]))
	}
	if out.Len() == 0 {
		return "Value"
	}
	return out.String()
}
