package schema

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

func TestSessionGraphPrimitiveCorpus(t *testing.T) {
	c, err := LoadSessionGraphFixtures()
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireMin(t, c, 18)
	caseassert.RequireValid(t, c)
	required := []string{"ref-96-ascii", "ref-96-multibyte", "ref-over-byte-boundary", "ref-malformed-not-trimmed", "relationship-known-target", "relationship-unknown-forbids-target", "relationship-unique-kinds", "anchor-general", "anchor-exact", "anchor-partial-rejected", "provenance-all-unknown", "provenance-empty-rejected", "navigation-local-resolved", "navigation-unavailable-no-id", "navigation-unavailable-leaks-id", "helper-group-valid", "helper-group-wrong-purpose", "helper-context-valid", "earlier-empty-valid", "earlier-null-invalid"}
	names := map[string]bool{}
	for _, x := range c.Cases {
		names[x.Name] = true
		err := ValidateSessionGraphFixtureInput(x.Input)
		if x.Classification == testcase.MustPass && err != nil {
			t.Errorf("%s: unexpected error: %v", x.Name, err)
		}
		if x.Classification == testcase.MustFail && (err == nil || !strings.Contains(err.Error(), x.Expected.ErrorContains)) {
			t.Errorf("%s: error=%v want contains %q", x.Name, err, x.Expected.ErrorContains)
		}
	}
	for _, name := range required {
		if !names[name] {
			t.Errorf("required fixture %q missing", name)
		}
	}
}

func TestSessionGraphClosedSetsExact(t *testing.T) {
	checks := map[string][]string{
		"relationship": stringSlice(AllSessionRelationshipKinds), "target": stringSlice(AllRelationshipTargetStates), "evidence": stringSlice(AllEvidenceKinds), "purpose": stringSlice(AllSessionPurposes), "origin": stringSlice(AllContentOrigins), "actor": stringSlice(AllActorOrigins), "delivery": stringSlice(AllDeliveryOrigins), "ownership": stringSlice(AllContentOwnerships), "modality": stringSlice(AllInputModalities), "anchor": stringSlice(AllPublicSourceAnchorKinds), "earlier": stringSlice(AllEarlierHistoryStates), "navigation": stringSlice(AllRelationshipNavigationStatuses), "list-item": stringSlice(AllSessionListItemKinds)}
	expected := map[string]string{"relationship": "started_by,context_from", "target": "target_known,target_known_retained,explicit_none,unknown,conflicting_current_native_evidence", "evidence": "native_typed,lifecycle_typed,existing_adapter,retained_last_good,unknown,conflict", "purpose": "interaction,delegated_work,helper_review,unknown", "origin": "submitted_input,harness_context,agent_output,agent_communication,tool_activity,system_control,generated_summary,unknown", "actor": "operator,agent_delegate,harness,unknown", "delivery": "session_admission,guardian_review,subagent_delivery,inherited_context,tool_delivery,system_lifecycle,unknown", "ownership": "local,inherited,uncertain", "modality": "none,text,media,user_action,mixed,unknown", "anchor": "general_source_session,before_redacted_entry,through_redacted_entry", "earlier": "uncertain_migrated,uncertain_unresolved", "navigation": "resolved,general_link_only,known_unavailable,inaccessible,unknown,conflicting", "list-item": "transcript,context_container"}
	for name, values := range checks {
		if strings.Join(values, ",") != expected[name] {
			t.Errorf("%s inventory=%v", name, values)
		}
	}
}
func stringSlice[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return out
}
