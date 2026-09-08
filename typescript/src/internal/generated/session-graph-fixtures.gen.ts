// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
export const canonicalSessionGraphFixtures = {
  "refs": {
    "cases": [
      {
        "name": "source-empty",
        "input": {
          "alias": "source",
          "bytesBase64": ""
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "required-ref"
        },
        "mutation": {
          "description": "empty required source ref"
        }
      },
      {
        "name": "submission-invalid-utf8",
        "input": {
          "alias": "submission",
          "bytesBase64": "/w=="
        },
        "expected": {
          "errorContains": "UTF-8"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "invalid-utf8"
        },
        "mutation": {
          "description": "invalid UTF-8 byte vector"
        }
      },
      {
        "name": "revision-one-byte",
        "input": {
          "alias": "revision",
          "bytesBase64": "YQ=="
        },
        "expected": {
          "valueBase64": "YQ=="
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "one-byte"
        },
        "mutation": {
          "description": "minimum valid ref"
        }
      },
      {
        "name": "source-96-multibyte",
        "input": {
          "alias": "source",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "expected": {
          "valueBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "utf8-96"
        },
        "mutation": {
          "description": "exact multibyte limit"
        }
      },
      {
        "name": "submission-97-multibyte",
        "input": {
          "alias": "submission",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WMYQ=="
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "utf8-97"
        },
        "mutation": {
          "description": "one byte above multibyte limit"
        }
      },
      {
        "name": "revision-preserves-space",
        "input": {
          "alias": "revision",
          "bytesBase64": "IHIg"
        },
        "expected": {
          "valueBase64": "IHIg"
        },
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "exact-bytes"
        },
        "mutation": {
          "description": "no trimming"
        }
      },
      {
        "name": "source-invalid-bytes",
        "input": {
          "alias": "source",
          "bytesBase64": "/w=="
        },
        "expected": {
          "errorContains": "UTF-8"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "source-invalid"
        },
        "mutation": {
          "description": "invalid byte boundary for source"
        }
      },
      {
        "name": "source-one-bytes",
        "input": {
          "alias": "source",
          "bytesBase64": "YQ=="
        },
        "expected": {
          "valueBase64": "YQ=="
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "source-one"
        },
        "mutation": {
          "description": "one byte boundary for source"
        }
      },
      {
        "name": "source-97-bytes",
        "input": {
          "alias": "source",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WMYQ=="
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "source-97"
        },
        "mutation": {
          "description": "97 byte boundary for source"
        }
      },
      {
        "name": "submission-empty-bytes",
        "input": {
          "alias": "submission",
          "bytesBase64": ""
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "submission-empty"
        },
        "mutation": {
          "description": "empty byte boundary for submission"
        }
      },
      {
        "name": "submission-one-bytes",
        "input": {
          "alias": "submission",
          "bytesBase64": "YQ=="
        },
        "expected": {
          "valueBase64": "YQ=="
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "submission-one"
        },
        "mutation": {
          "description": "one byte boundary for submission"
        }
      },
      {
        "name": "submission-96-bytes",
        "input": {
          "alias": "submission",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "expected": {
          "valueBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "submission-96"
        },
        "mutation": {
          "description": "96 byte boundary for submission"
        }
      },
      {
        "name": "revision-empty-bytes",
        "input": {
          "alias": "revision",
          "bytesBase64": ""
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "revision-empty"
        },
        "mutation": {
          "description": "empty byte boundary for revision"
        }
      },
      {
        "name": "revision-invalid-bytes",
        "input": {
          "alias": "revision",
          "bytesBase64": "/w=="
        },
        "expected": {
          "errorContains": "UTF-8"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "revision-invalid"
        },
        "mutation": {
          "description": "invalid byte boundary for revision"
        }
      },
      {
        "name": "revision-96-bytes",
        "input": {
          "alias": "revision",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "expected": {
          "valueBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM"
        },
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "revision-96"
        },
        "mutation": {
          "description": "96 byte boundary for revision"
        }
      },
      {
        "name": "revision-97-bytes",
        "input": {
          "alias": "revision",
          "bytesBase64": "55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WM55WMYQ=="
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "revision-97"
        },
        "mutation": {
          "description": "97 byte boundary for revision"
        }
      }
    ]
  },
  "enums": {
    "cases": [
      {
        "name": "enum-relationship",
        "input": {
          "enum": "relationship",
          "members": [
            "started_by",
            "context_from"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "relationship-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-target",
        "input": {
          "enum": "target",
          "members": [
            "target_known",
            "target_known_retained",
            "explicit_none",
            "unknown",
            "conflicting_current_native_evidence"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "target-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-evidence",
        "input": {
          "enum": "evidence",
          "members": [
            "native_typed",
            "lifecycle_typed",
            "existing_adapter",
            "retained_last_good",
            "unknown",
            "conflict"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "evidence-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-purpose",
        "input": {
          "enum": "purpose",
          "members": [
            "interaction",
            "delegated_work",
            "helper_review",
            "unknown"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "purpose-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-origin",
        "input": {
          "enum": "origin",
          "members": [
            "submitted_input",
            "harness_context",
            "agent_output",
            "agent_communication",
            "tool_activity",
            "system_control",
            "generated_summary",
            "unknown"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "origin-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-actor",
        "input": {
          "enum": "actor",
          "members": [
            "operator",
            "agent_delegate",
            "harness",
            "unknown"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "actor-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-delivery",
        "input": {
          "enum": "delivery",
          "members": [
            "session_admission",
            "guardian_review",
            "subagent_delivery",
            "inherited_context",
            "tool_delivery",
            "system_lifecycle",
            "unknown"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "delivery-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-ownership",
        "input": {
          "enum": "ownership",
          "members": [
            "local",
            "inherited",
            "uncertain"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "ownership-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-modality",
        "input": {
          "enum": "modality",
          "members": [
            "none",
            "text",
            "media",
            "user_action",
            "mixed",
            "unknown"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "modality-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-anchor",
        "input": {
          "enum": "anchor",
          "members": [
            "general_source_session",
            "before_redacted_entry",
            "through_redacted_entry"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "anchor-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-earlier",
        "input": {
          "enum": "earlier",
          "members": [
            "uncertain_migrated",
            "uncertain_unresolved"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "earlier-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-navigation",
        "input": {
          "enum": "navigation",
          "members": [
            "resolved",
            "general_link_only",
            "known_unavailable",
            "inaccessible",
            "unknown",
            "conflicting"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "navigation-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      },
      {
        "name": "enum-list_item",
        "input": {
          "enum": "list_item",
          "members": [
            "transcript",
            "context_container"
          ],
          "empty": "",
          "unknown": "future_value"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "enum",
          "ref": "list_item-closed-set"
        },
        "mutation": {
          "description": "exact members and unknown rejection"
        }
      }
    ]
  },
  "relationships": {
    "cases": [
      {
        "name": "relationship-known-target",
        "input": {
          "kind": "started_by",
          "targetState": "target_known",
          "targetLocalId": "ses_parent1",
          "evidence": "native_typed"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "known target has ID"
        }
      },
      {
        "name": "relationship-unknown-forbids-target",
        "input": {
          "kind": "started_by",
          "targetState": "unknown",
          "targetLocalId": "ses_parent1",
          "evidence": "unknown"
        },
        "expected": {
          "errorContains": "disagree"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "unknown state carries forbidden ID"
        }
      },
      {
        "name": "anchor-general",
        "input": {
          "kind": "context_from",
          "targetState": "target_known",
          "targetLocalId": "ses_parent1",
          "evidence": "native_typed",
          "anchor": {
            "kind": "general_source_session"
          }
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "general context anchor"
        }
      },
      {
        "name": "anchor-exact",
        "input": {
          "kind": "context_from",
          "targetState": "target_known_retained",
          "targetLocalId": "ses_parent1",
          "evidence": "retained_last_good",
          "anchor": {
            "kind": "before_redacted_entry",
            "sourceEntryRef": "e1",
            "sourceRevisionRef": "r1"
          }
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "exact context anchor"
        }
      },
      {
        "name": "anchor-partial-rejected",
        "input": {
          "kind": "context_from",
          "targetState": "target_known",
          "targetLocalId": "ses_parent1",
          "evidence": "native_typed",
          "anchor": {
            "kind": "through_redacted_entry",
            "sourceEntryRef": "e1"
          }
        },
        "expected": {
          "errorContains": "requires both"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "exact anchor missing revision"
        }
      },
      {
        "name": "relationship-known-missing-target",
        "input": {
          "kind": "started_by",
          "targetState": "target_known",
          "evidence": "native_typed"
        },
        "expected": {
          "errorContains": "disagree"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "known state missing ID"
        }
      },
      {
        "name": "relationship-known-invalid-target",
        "input": {
          "kind": "started_by",
          "targetState": "target_known",
          "targetLocalId": "bad/id",
          "evidence": "native_typed"
        },
        "expected": {
          "errorContains": "malformed"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "malformed known ID"
        }
      },
      {
        "name": "relationship-explicit-none-forbids-target",
        "input": {
          "kind": "started_by",
          "targetState": "explicit_none",
          "targetLocalId": "ses_parent1",
          "evidence": "lifecycle_typed"
        },
        "expected": {
          "errorContains": "disagree"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "explicit none with ID"
        }
      },
      {
        "name": "relationship-conflict-forbids-target",
        "input": {
          "kind": "context_from",
          "targetState": "conflicting_current_native_evidence",
          "targetLocalId": "ses_parent1",
          "evidence": "conflict"
        },
        "expected": {
          "errorContains": "disagree"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "relationship-matrix"
        },
        "mutation": {
          "description": "conflict with ID"
        }
      },
      {
        "name": "anchor-forbidden-started-by",
        "input": {
          "kind": "started_by",
          "targetState": "target_known",
          "targetLocalId": "ses_parent1",
          "evidence": "native_typed",
          "anchor": {
            "kind": "general_source_session"
          }
        },
        "expected": {
          "errorContains": "only on a known context_from"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "started-by anchor"
        }
      },
      {
        "name": "anchor-forbidden-nonknown",
        "input": {
          "kind": "context_from",
          "targetState": "unknown",
          "evidence": "unknown",
          "anchor": {
            "kind": "general_source_session"
          }
        },
        "expected": {
          "errorContains": "only on a known context_from"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "unknown relation anchor"
        }
      },
      {
        "name": "anchor-general-forbids-refs",
        "input": {
          "kind": "context_from",
          "targetState": "target_known",
          "targetLocalId": "ses_parent1",
          "evidence": "native_typed",
          "anchor": {
            "kind": "general_source_session",
            "sourceEntryRef": "e1"
          }
        },
        "expected": {
          "errorContains": "carries an entry"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "anchor-matrix"
        },
        "mutation": {
          "description": "general anchor with exact ref"
        }
      }
    ]
  },
  "relationship_sets": {
    "cases": [
      {
        "name": "relationship-unique-kinds",
        "input": {
          "relationships": [
            {
              "kind": "started_by",
              "targetState": "explicit_none",
              "evidence": "lifecycle_typed"
            },
            {
              "kind": "started_by",
              "targetState": "unknown",
              "evidence": "unknown"
            }
          ]
        },
        "expected": {
          "errorContains": "more than once"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "unique-kind"
        },
        "mutation": {
          "description": "duplicate relationship kind"
        }
      },
      {
        "name": "relationship-legal-unique-pair",
        "input": {
          "relationships": [
            {
              "kind": "started_by",
              "targetState": "target_known",
              "targetLocalId": "ses_parent1",
              "evidence": "native_typed"
            },
            {
              "kind": "context_from",
              "targetState": "unknown",
              "evidence": "unknown"
            }
          ]
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "unique-kind"
        },
        "mutation": {
          "description": "one relation of each kind"
        }
      }
    ]
  },
  "provenance": {
    "cases": [
      {
        "name": "provenance-all-unknown",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "unknown"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "unknown-valid"
        },
        "mutation": {
          "description": "complete unknown evidence"
        }
      },
      {
        "name": "provenance-empty-rejected",
        "input": {},
        "expected": {
          "errorContains": "missing"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-required"
        },
        "mutation": {
          "description": "empty object"
        }
      },
      {
        "name": "provenance-invalid-origin",
        "input": {
          "origin": "future",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "unknown"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid origin dimension"
        }
      },
      {
        "name": "provenance-invalid-actor",
        "input": {
          "origin": "unknown",
          "actor": "future",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "unknown"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid actor dimension"
        }
      },
      {
        "name": "provenance-invalid-delivery",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "future",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "unknown"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid delivery dimension"
        }
      },
      {
        "name": "provenance-invalid-ownership",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "future",
          "evidence": "unknown",
          "inputModality": "unknown"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid ownership dimension"
        }
      },
      {
        "name": "provenance-invalid-evidence",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "future",
          "inputModality": "unknown"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid evidence dimension"
        }
      },
      {
        "name": "provenance-invalid-modality",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "future"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "provenance-dimensions"
        },
        "mutation": {
          "description": "invalid modality dimension"
        }
      },
      {
        "name": "provenance-invalid-submission",
        "input": {
          "origin": "unknown",
          "actor": "unknown",
          "delivery": "unknown",
          "ownership": "uncertain",
          "evidence": "unknown",
          "inputModality": "unknown",
          "submissionRef": ""
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "optional-submission"
        },
        "mutation": {
          "description": "absent optional submission"
        }
      },
      {
        "name": "provenance-overlong-submission",
        "input": {
          "origin": "submitted_input",
          "actor": "unknown",
          "delivery": "session_admission",
          "ownership": "local",
          "evidence": "native_typed",
          "inputModality": "text",
          "submissionRef": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        },
        "expected": {
          "errorContains": "1..96"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "submission-bound"
        },
        "mutation": {
          "description": "overlong submission ref"
        }
      }
    ]
  },
  "navigation": {
    "cases": [
      {
        "name": "navigation-local-resolved",
        "input": {
          "kind": "started_by",
          "status": "resolved",
          "localId": "ses_parent1"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-matrix"
        },
        "mutation": {
          "description": "resolved local link"
        }
      },
      {
        "name": "navigation-unavailable-no-id",
        "input": {
          "kind": "started_by",
          "status": "known_unavailable"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-matrix"
        },
        "mutation": {
          "description": "unavailable target hides ID"
        }
      },
      {
        "name": "navigation-unavailable-leaks-id",
        "input": {
          "kind": "started_by",
          "status": "inaccessible",
          "localId": "ses_parent1"
        },
        "expected": {
          "errorContains": "exactly one"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-matrix"
        },
        "mutation": {
          "description": "inaccessible target leaks ID"
        }
      },
      {
        "name": "navigation-public-resolved",
        "input": {
          "kind": "context_from",
          "status": "resolved",
          "transcriptId": "99d59925-36bc-424c-a789-8be54d9702ba",
          "anchor": {
            "kind": "through_redacted_entry",
            "sourceEntryRef": "e1",
            "sourceRevisionRef": "r1"
          }
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-targets"
        },
        "mutation": {
          "description": "public exact resolved target"
        }
      },
      {
        "name": "navigation-resolved-missing-target",
        "input": {
          "kind": "started_by",
          "status": "resolved"
        },
        "expected": {
          "errorContains": "exactly one"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-targets"
        },
        "mutation": {
          "description": "resolved without target"
        }
      },
      {
        "name": "navigation-resolved-dual-target",
        "input": {
          "kind": "started_by",
          "status": "resolved",
          "localId": "ses_parent1",
          "transcriptId": "99d59925-36bc-424c-a789-8be54d9702ba"
        },
        "expected": {
          "errorContains": "exactly one"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "navigation-targets"
        },
        "mutation": {
          "description": "local and public targets conflict"
        }
      },
      {
        "name": "navigation-general-link-absent-anchor",
        "input": {
          "kind": "context_from",
          "status": "general_link_only",
          "localId": "ses_parent1"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "general-link"
        },
        "mutation": {
          "description": "general link without anchor"
        }
      },
      {
        "name": "navigation-general-link-general-anchor",
        "input": {
          "kind": "context_from",
          "status": "general_link_only",
          "localId": "ses_parent1",
          "anchor": {
            "kind": "general_source_session"
          }
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "general-link"
        },
        "mutation": {
          "description": "explicit general anchor"
        }
      },
      {
        "name": "navigation-general-link-exact-anchor",
        "input": {
          "kind": "context_from",
          "status": "general_link_only",
          "localId": "ses_parent1",
          "anchor": {
            "kind": "before_redacted_entry",
            "sourceEntryRef": "e1",
            "sourceRevisionRef": "r1"
          }
        },
        "expected": {
          "errorContains": "promises only a general source link"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "general-link"
        },
        "mutation": {
          "description": "contradictory exact anchor"
        }
      }
    ]
  },
  "helper_groups": {
    "cases": [
      {
        "name": "helper-group-valid",
        "input": {
          "groupId": "hg_public",
          "purpose": "helper_review",
          "helperThreadCount": 2,
          "memberScope": "scope_public"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "helper-summary"
        },
        "mutation": {
          "description": "saved helper identities"
        }
      },
      {
        "name": "helper-group-wrong-purpose",
        "input": {
          "groupId": "hg_public",
          "purpose": "interaction",
          "helperThreadCount": 2,
          "memberScope": "scope_public"
        },
        "expected": {
          "errorContains": "helper_review"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "helper-summary"
        },
        "mutation": {
          "description": "non-helper purpose"
        }
      },
      {
        "name": "helper-negative-count",
        "input": {
          "groupId": "hg_public",
          "purpose": "helper_review",
          "helperThreadCount": -1,
          "memberScope": "scope_public"
        },
        "expected": {
          "errorContains": "negative"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "helper-summary"
        },
        "mutation": {
          "description": "negative saved identity count"
        }
      },
      {
        "name": "helper-empty-group",
        "input": {
          "groupId": "",
          "purpose": "helper_review",
          "helperThreadCount": 0,
          "memberScope": "scope_public"
        },
        "expected": {
          "errorContains": "non-empty"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "helper-summary"
        },
        "mutation": {
          "description": "empty stable group ID"
        }
      },
      {
        "name": "helper-empty-scope",
        "input": {
          "groupId": "hg_public",
          "purpose": "helper_review",
          "helperThreadCount": 0,
          "memberScope": ""
        },
        "expected": {
          "errorContains": "non-empty"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "helper-summary"
        },
        "mutation": {
          "description": "empty route scope"
        }
      }
    ]
  },
  "helper_contexts": {
    "cases": [
      {
        "name": "helper-context-valid",
        "input": {
          "groupId": "hg_public",
          "ownerStatus": "known_unavailable"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "helper-context"
        },
        "mutation": {
          "description": "unavailable owner context"
        }
      },
      {
        "name": "helper-context-invalid-status",
        "input": {
          "groupId": "hg_public",
          "ownerStatus": "future"
        },
        "expected": {
          "errorContains": "outside"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "helper-context"
        },
        "mutation": {
          "description": "unknown owner status token"
        }
      }
    ]
  },
  "earlier_history": {
    "cases": [
      {
        "name": "earlier-empty-valid",
        "input": {
          "state": "uncertain_migrated",
          "turns": []
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "earlier-history"
        },
        "mutation": {
          "description": "non-null empty retained section"
        }
      },
      {
        "name": "earlier-null-invalid",
        "input": {
          "state": "uncertain_unresolved",
          "turns": null
        },
        "expected": {
          "errorContains": "turns is null"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "earlier-history"
        },
        "mutation": {
          "description": "null turn array"
        }
      }
    ]
  },
  "raw_relationships": {
    "cases": [
      {
        "name": "raw-relationship-wrong-target-type",
        "input": {
          "rawJSON": "{\"kind\":\"started_by\",\"targetState\":\"target_known\",\"targetLocalId\":7,\"evidence\":\"native_typed\"}"
        },
        "expected": {
          "errorContains": "cannot unmarshal"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "raw-primitive-shape"
        },
        "mutation": {
          "description": "numeric target ID"
        }
      }
    ]
  },
  "raw_navigation": {
    "cases": [
      {
        "name": "raw-navigation-roundtrip-shape",
        "input": {
          "rawJSON": "{\"kind\":\"context_from\",\"status\":\"general_link_only\",\"localId\":\"ses_parent1\",\"anchor\":{\"kind\":\"general_source_session\"}}"
        },
        "expected": {
          "rawJSON": "{\"kind\":\"context_from\",\"status\":\"general_link_only\",\"localId\":\"ses_parent1\",\"anchor\":{\"kind\":\"general_source_session\"}}"
        },
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "raw-primitive-shape"
        },
        "mutation": {
          "description": "valid raw navigation object"
        }
      }
    ]
  },
  "recursive": {
    "cases": [
      {
        "name": "recursive-valid-pi-earlier-evidence",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_pi\",\"harness\":\"pi\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"observedModel\":\"model/x\",\"sourceEntryRef\":\"e_old\",\"usage\":{\"ownerId\":\"u_old\",\"sourceEntryRef\":\"e_old\",\"scope\":\"assistant\",\"completeness\":\"unknown\"}}],\"nativeMetadata\":[{\"id\":\"m_old\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"native_old\",\"sourceType\":\"pi.custom\"},\"customType\":\"extension\",\"data\":{}}]}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "recursive-valid-pi"
        },
        "mutation": {
          "description": "earlier model usage and native evidence validate at production decode"
        }
      },
      {
        "name": "recursive-duplicate-main-earlier-ref",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_dup\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"user\",\"content\":\"main\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_same\"}],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"user\",\"content\":\"old\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_same\"}]}]}"
        },
        "expected": {
          "errorContains": "duplicate block reference"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "global-ref-uniqueness"
        },
        "mutation": {
          "description": "earlier block reuses a main ref"
        }
      },
      {
        "name": "recursive-duplicate-folded-ref",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_dup\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_carrier\",\"toolCalls\":[{\"id\":\"call_1\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_carrier\",\"resultEntryRef\":\"e_result\"}]}]}"
        },
        "expected": {
          "errorContains": "duplicate block reference"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "folded-ref-uniqueness"
        },
        "mutation": {
          "description": "folded call steals carrier ref"
        }
      },
      {
        "name": "recursive-parent-cross-partition",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_parent\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":2,\"role\":\"assistant\",\"content\":\"child\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":1,\"parentIndex\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_child\"}]}]}"
        },
        "expected": {
          "errorContains": "same partition"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "partition-parent"
        },
        "mutation": {
          "description": "parent index is absent from its earlier partition"
        }
      },
      {
        "name": "recursive-earlier-model-role",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_model\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"user\",\"content\":\"old\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"observedModel\":\"model/x\",\"sourceEntryRef\":\"e_old\"}]}]}"
        },
        "expected": {
          "errorContains": "assistant"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "observed-model-role"
        },
        "mutation": {
          "description": "earlier user turn carries observed model"
        }
      },
      {
        "name": "recursive-folded-only-dangling-parent",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_folded\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":1,\"parentIndex\":99,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call_folded\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\"}]}]}"
        },
        "expected": {
          "errorContains": "same partition"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "folded-carrier-parent"
        },
        "mutation": {
          "description": "folded-only carrier targets absent parent"
        }
      },
      {
        "name": "recursive-folded-local-parent",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_folded\",\"harness\":\"claude-code\",\"turnCount\":2,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"parent\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_parent\"},{\"index\":1,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:01Z\",\"depth\":1,\"parentIndex\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call_folded\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\"}]}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "folded-carrier-parent"
        },
        "mutation": {
          "description": "folded-only carrier targets emitted local parent"
        }
      },
      {
        "name": "recursive-valid-pi-earlier-folded-attachment",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_pi_tool\",\"harness\":\"pi\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"observedModel\":\"model/x\",\"sourceEntryRef\":\"e_carrier\",\"usage\":{\"ownerId\":\"u_assistant\",\"sourceEntryRef\":\"e_carrier\",\"scope\":\"assistant\",\"completeness\":\"unknown\"},\"toolCalls\":[{\"id\":\"call_old\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\",\"usage\":{\"ownerId\":\"u_tool\",\"sourceEntryRef\":\"e_result\",\"scope\":\"tool\",\"completeness\":\"unknown\"}}]}],\"nativeMetadata\":[{\"id\":\"m_result\",\"kind\":\"pi.toolresult.details\",\"source\":{\"entryRef\":\"e_result\",\"sourceType\":\"pi.message\",\"messageRole\":\"toolResult\"},\"attachment\":{\"turnIndex\":0,\"toolCallId\":\"call_old\"},\"data\":{}}]}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "recursive-pi-attachment"
        },
        "mutation": {
          "description": "earlier folded result owns matching usage metadata and observed model"
        }
      },
      {
        "name": "recursive-native-source-cross-partition",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_pi_dup\",\"harness\":\"pi\",\"turnCount\":0,\"turns\":[],\"nativeMetadata\":[{\"id\":\"m_main\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"native_same\",\"sourceType\":\"pi.custom\"},\"customType\":\"x\",\"data\":{}}],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[],\"nativeMetadata\":[{\"id\":\"m_old\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"native_same\",\"sourceType\":\"pi.custom\"},\"customType\":\"x\",\"data\":{}}]}]}"
        },
        "expected": {
          "errorContains": "duplicated"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "native-source-uniqueness"
        },
        "mutation": {
          "description": "earlier native record reuses main source identity"
        }
      },
      {
        "name": "recursive-native-wrong-partition-attachment",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_pi_wrong\",\"harness\":\"pi\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"main\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_main\",\"toolCalls\":[{\"id\":\"call_main\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\"}]}],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[],\"nativeMetadata\":[{\"id\":\"m_wrong\",\"kind\":\"pi.toolresult.details\",\"source\":{\"entryRef\":\"e_result\",\"sourceType\":\"pi.message\",\"messageRole\":\"toolResult\"},\"attachment\":{\"turnIndex\":0,\"toolCallId\":\"call_main\"},\"data\":{}}]}]}"
        },
        "expected": {
          "errorContains": "disagrees"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "partition-attachment"
        },
        "mutation": {
          "description": "earlier metadata targets a main folded result"
        }
      }
    ]
  },
  "recursive_required_names": [
    "recursive-valid-pi-earlier-evidence",
    "recursive-duplicate-main-earlier-ref",
    "recursive-duplicate-folded-ref",
    "recursive-parent-cross-partition",
    "recursive-earlier-model-role",
    "recursive-folded-only-dangling-parent",
    "recursive-folded-local-parent",
    "recursive-valid-pi-earlier-folded-attachment",
    "recursive-native-source-cross-partition",
    "recursive-native-wrong-partition-attachment"
  ],
  "raw_durable": {
    "cases": [
      {
        "name": "raw-empty-optional-main-tool-refs",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_emptyrefs\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"\",\"toolCalls\":[{\"id\":\"call\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"\",\"resultEntryRef\":\"\"}]}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "optional-public-references"
        },
        "mutation": {
          "description": "empty optional main and folded references normalize to absence rather than invalid required aliases"
        }
      },
      {
        "name": "raw-empty-optional-earlier-refs",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_emptyrefs\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"\",\"toolCalls\":[{\"id\":\"call\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"\",\"resultEntryRef\":\"\"}]}]}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "optional-public-references"
        },
        "mutation": {
          "description": "earlier partitions share optional reference normalization without crossing identity domains"
        }
      },
      {
        "name": "raw-empty-general-anchor-refs",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_emptyrefs\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationships\":[{\"kind\":\"context_from\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent\",\"evidence\":\"native_typed\",\"anchor\":{\"kind\":\"general_source_session\",\"sourceEntryRef\":\"\",\"sourceRevisionRef\":\"\"}}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "optional-public-references"
        },
        "mutation": {
          "description": "general link treats empty optional source and revision as absent"
        }
      },
      {
        "name": "raw-empty-exact-anchor-refs-rejected",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_emptyrefs\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationships\":[{\"kind\":\"context_from\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent\",\"evidence\":\"native_typed\",\"anchor\":{\"kind\":\"before_redacted_entry\",\"sourceEntryRef\":\"\",\"sourceRevisionRef\":\"\"}}]}"
        },
        "expected": {
          "errorContains": "requires both"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "optional-public-references"
        },
        "mutation": {
          "description": "normalization cannot make an exact anchor valid without its required reference pair"
        }
      },
      {
        "name": "raw-null-earlier-provenance",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"user\",\"content\":\"old\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"provenance\":null}]}]}"
        },
        "expected": {
          "errorContains": "explicitly null"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "raw-present-evidence"
        },
        "mutation": {
          "description": "null earlier provenance would decode as absent"
        }
      },
      {
        "name": "raw-null-folded-provenance",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call_1\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callProvenance\":null}]}]}"
        },
        "expected": {
          "errorContains": "explicitly null"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "raw-present-evidence"
        },
        "mutation": {
          "description": "null folded provenance would decode as absent"
        }
      },
      {
        "name": "raw-forbidden-relationship-navigation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationshipNavigation\":[]}"
        },
        "expected": {
          "errorContains": "forbidden read-only field"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "authorized navigation is submitted as durable content"
        }
      },
      {
        "name": "raw-forbidden-resolved",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"resolved\":true}"
        },
        "expected": {
          "errorContains": "forbidden read-only field"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "resolved read state is submitted"
        }
      },
      {
        "name": "raw-forbidden-cooked",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"cooked\":{}}"
        },
        "expected": {
          "errorContains": "forbidden read-only field"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "cooked view state is submitted"
        }
      },
      {
        "name": "raw-forbidden-detail-wrapper",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"detail\":{}}"
        },
        "expected": {
          "errorContains": "forbidden read-only field"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "read detail wrapper is submitted"
        }
      },
      {
        "name": "raw-unrelated-future-key-compatible",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"futureExtension\":\"kept-compatible\"}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "additive-forward-compatibility"
        },
        "mutation": {
          "description": "unrelated unknown durable key remains ignorable"
        }
      },
      {
        "name": "raw-relationship-read-fields",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"unknown\",\"evidence\":\"unknown\",\"status\":\"resolved\"}]}"
        },
        "expected": {
          "errorContains": "relationships/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "relationship contains resolved navigation state"
        }
      },
      {
        "name": "raw-anchor-label",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationships\":[{\"kind\":\"context_from\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent\",\"evidence\":\"native_typed\",\"anchor\":{\"kind\":\"general_source_session\",\"label\":\"parent\"}}]}"
        },
        "expected": {
          "errorContains": "anchor"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "durable anchor contains cooked label"
        }
      },
      {
        "name": "raw-earlier-collapse",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[],\"collapsed\":true}]}"
        },
        "expected": {
          "errorContains": "earlierHistory/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "earlier section contains cooked collapse state"
        }
      },
      {
        "name": "raw-earlier-provenance-label",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"user\",\"content\":\"old\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"provenance\":{\"origin\":\"unknown\",\"actor\":\"unknown\",\"delivery\":\"unknown\",\"ownership\":\"uncertain\",\"evidence\":\"unknown\",\"inputModality\":\"unknown\",\"label\":\"cooked\"}}]}]}"
        },
        "expected": {
          "errorContains": "provenance"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "earlier provenance contains cooked label"
        }
      },
      {
        "name": "raw-protected-turn-label",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_turn\",\"label\":\"cooked\"}]}"
        },
        "expected": {
          "errorContains": "turns/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "protected turn contains cooked label"
        }
      },
      {
        "name": "raw-protected-turn-navigation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_turn\",\"relationshipNavigation\":[]}]}"
        },
        "expected": {
          "errorContains": "turns/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "protected turn contains authorized navigation"
        }
      },
      {
        "name": "raw-folded-tool-status",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\",\"status\":\"resolved\"}]}]}"
        },
        "expected": {
          "errorContains": "toolCalls/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "folded tool contains resolved read status"
        }
      },
      {
        "name": "raw-folded-tool-cooked",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\",\"cooked\":{}}]}]}"
        },
        "expected": {
          "errorContains": "toolCalls/0"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "folded tool contains cooked projection"
        }
      },
      {
        "name": "raw-turn-future-compatible",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_turn\",\"futureExtension\":\"ignored\"}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "additive-forward-compatibility"
        },
        "mutation": {
          "description": "unrelated future turn field remains compatible"
        }
      },
      {
        "name": "raw-main-turn-explanation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_turn\",\"explanation\":\"cooked\"}]}"
        },
        "expected": {
          "errorContains": "explanation"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "main protected turn contains cooked explanation"
        }
      },
      {
        "name": "raw-earlier-turn-explanation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"kept\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"sourceEntryRef\":\"e_old\",\"explanation\":\"cooked\"}]}]}"
        },
        "expected": {
          "errorContains": "explanation"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "earlier protected turn contains cooked explanation"
        }
      },
      {
        "name": "raw-main-tool-explanation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":1,\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call\",\"resultEntryRef\":\"e_result\",\"explanation\":\"cooked\"}]}]}"
        },
        "expected": {
          "errorContains": "explanation"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "main folded tool contains cooked explanation"
        }
      },
      {
        "name": "raw-earlier-tool-explanation",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_raw\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"entryType\":\"text\",\"toolCalls\":[{\"id\":\"call_old\",\"name\":\"x\",\"arguments\":\"a\",\"result\":\"r\",\"callEntryRef\":\"e_call_old\",\"resultEntryRef\":\"e_result_old\",\"explanation\":\"cooked\"}]}]}]}"
        },
        "expected": {
          "errorContains": "explanation"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "earlier folded tool contains cooked explanation"
        }
      },
      {
        "name": "legacy-parent-omitted-with-unknown-target",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_parent_case\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"unknown\",\"evidence\":\"unknown\"}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "legacy-parent-agreement"
        },
        "mutation": {
          "description": "optional legacy parent is omitted"
        }
      },
      {
        "name": "legacy-parent-null-with-unknown-target",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_parent_case\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"parentSessionId\":null,\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"unknown\",\"evidence\":\"unknown\"}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "legacy-parent-agreement"
        },
        "mutation": {
          "description": "JSON null decodes as an absent optional legacy parent"
        }
      },
      {
        "name": "legacy-parent-matches-known-target",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_parent_case\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"parentSessionId\":\"ses_parent1\",\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent1\",\"evidence\":\"native_typed\"}]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "legacy-parent-agreement"
        },
        "mutation": {
          "description": "legacy parent agrees with known durable target"
        }
      },
      {
        "name": "legacy-parent-conflicts-known-target",
        "input": {
          "rawJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_parent_case\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[],\"parentSessionId\":\"ses_other\",\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent1\",\"evidence\":\"native_typed\"}]}"
        },
        "expected": {
          "errorContains": "parentSessionId disagrees"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "legacy-parent-agreement"
        },
        "mutation": {
          "description": "legacy parent conflicts with known durable target"
        }
      }
    ]
  },
  "raw_durable_required_names": [
    "raw-empty-optional-main-tool-refs",
    "raw-empty-optional-earlier-refs",
    "raw-empty-general-anchor-refs",
    "raw-empty-exact-anchor-refs-rejected",
    "raw-null-earlier-provenance",
    "raw-null-folded-provenance",
    "raw-forbidden-relationship-navigation",
    "raw-forbidden-resolved",
    "raw-forbidden-cooked",
    "raw-forbidden-detail-wrapper",
    "raw-unrelated-future-key-compatible",
    "raw-relationship-read-fields",
    "raw-anchor-label",
    "raw-earlier-collapse",
    "raw-earlier-provenance-label",
    "raw-protected-turn-label",
    "raw-protected-turn-navigation",
    "raw-folded-tool-status",
    "raw-folded-tool-cooked",
    "raw-turn-future-compatible",
    "raw-main-turn-explanation",
    "raw-earlier-turn-explanation",
    "raw-main-tool-explanation",
    "raw-earlier-tool-explanation",
    "legacy-parent-omitted-with-unknown-target",
    "legacy-parent-null-with-unknown-target",
    "legacy-parent-matches-known-target",
    "legacy-parent-conflicts-known-target"
  ],
  "native_limits": {
    "cases": [
      {
        "name": "native-aggregate-under-limit",
        "input": {
          "main_records": 8,
          "earlier_records": [
            4,
            5
          ],
          "data_bytes": 61000,
          "detailJSON": "{\"id\":\"ses_native_budget\",\"harness\":\"pi\",\"startTime\":\"2026-09-08T00:00:00Z\",\"endTime\":\"2026-09-08T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"turns\":[]}",
          "record_json": "{\"id\":\"m_\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"source_\",\"sourceType\":\"pi.custom\"},\"customType\":\"fixture\",\"data\":[]}",
          "section_json": "{\"state\":\"uncertain_migrated\",\"turns\":[]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "native-aggregate-budget"
        },
        "mutation": {
          "description": "seventeen generated values across main and two earlier partitions remain below one MiB"
        }
      },
      {
        "name": "native-aggregate-exact-limit",
        "input": {
          "main_records": 4,
          "earlier_records": [
            6,
            6
          ],
          "data_bytes": 65536,
          "detailJSON": "{\"id\":\"ses_native_budget\",\"harness\":\"pi\",\"startTime\":\"2026-09-08T00:00:00Z\",\"endTime\":\"2026-09-08T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"turns\":[]}",
          "record_json": "{\"id\":\"m_\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"source_\",\"sourceType\":\"pi.custom\"},\"customType\":\"fixture\",\"data\":[]}",
          "section_json": "{\"state\":\"uncertain_migrated\",\"turns\":[]}"
        },
        "expected": {},
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "native-aggregate-budget"
        },
        "mutation": {
          "description": "sixteen generated values equal exactly one MiB"
        }
      },
      {
        "name": "native-aggregate-over-limit",
        "input": {
          "main_records": 8,
          "earlier_records": [
            5,
            5
          ],
          "data_bytes": 61000,
          "detailJSON": "{\"id\":\"ses_native_budget\",\"harness\":\"pi\",\"startTime\":\"2026-09-08T00:00:00Z\",\"endTime\":\"2026-09-08T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"turns\":[]}",
          "record_json": "{\"id\":\"m_\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"source_\",\"sourceType\":\"pi.custom\"},\"customType\":\"fixture\",\"data\":[]}",
          "section_json": "{\"state\":\"uncertain_migrated\",\"turns\":[]}"
        },
        "expected": {
          "errorContains": "aggregate"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "native-aggregate-budget"
        },
        "mutation": {
          "description": "eighteenth generated value across two earlier partitions crosses one MiB"
        }
      }
    ]
  },
  "native_limit_required_names": [
    "native-aggregate-under-limit",
    "native-aggregate-exact-limit",
    "native-aggregate-over-limit"
  ],
  "authoritative_raw": {
    "cases": [
      {
        "name": "authoritative-entry-label",
        "input": {
          "rawJSON": "{\"label\":\"cooked\"}"
        },
        "expected": {
          "errorContains": "entries"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "authoritative entry contains cooked label"
        }
      },
      {
        "name": "authoritative-entry-navigation",
        "input": {
          "rawJSON": "{\"relationshipNavigation\":[]}"
        },
        "expected": {
          "errorContains": "entries"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "authoritative entry contains authorized navigation"
        }
      },
      {
        "name": "authoritative-entry-explanation",
        "input": {
          "rawJSON": "{\"explanation\":\"cooked\"}"
        },
        "expected": {
          "errorContains": "entries"
        },
        "classification": "must-fail",
        "provenance": {
          "source": "requirement",
          "ref": "durable-read-separation"
        },
        "mutation": {
          "description": "authoritative entry contains cooked explanation"
        }
      }
    ]
  },
  "authoritative_required_names": [
    "authoritative-entry-label",
    "authoritative-entry-navigation",
    "authoritative-entry-explanation"
  ],
  "typescriptConstraints": {
    "countCarriers": [
      "AuthoritativeSessionStats",
      "LocalSyncSummary",
      "SessionStats",
      "SessionSummary",
      "SessionDetailPayload",
      "SessionDetailReadPayload",
      "VillageContributableTranscript",
      "VillageGroupTranscript",
      "VillagePendingShare",
      "VillageTranscript",
      "VillageUserGroupShare"
    ],
    "strictObjects": {
      "PublicSourceAnchor": {
        "kind": "general_source_session"
      },
      "SessionRelationship": {
        "kind": "started_by",
        "targetState": "explicit_none",
        "evidence": "native_typed"
      },
      "ContentProvenance": {
        "origin": "unknown",
        "actor": "unknown",
        "delivery": "unknown",
        "ownership": "uncertain",
        "evidence": "unknown",
        "inputModality": "unknown"
      },
      "EarlierHistorySection": {
        "state": "uncertain_unresolved",
        "turns": []
      },
      "SessionRelationshipNavigation": {
        "kind": "started_by",
        "status": "unknown"
      }
    }
  },
  "durable": {
    "round_trip": {
      "cases": [
        {
          "name": "legacy-durable-decode",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_legacy\",\"harness\":\"claude-code\",\"turnCount\":4,\"turns\":[]}"
          },
          "expected": {
            "turn_count": 4,
            "input_present": false,
            "main_refs": [],
            "earlier_refs": []
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "legacy-durable-decode"
          },
          "mutation": {
            "description": "graph fields omitted by an old producer"
          }
        },
        {
          "name": "three-independent-counts",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_graph\",\"harness\":\"claude-code\",\"turnCount\":5,\"inputSubmissionCount\":1,\"parentSessionId\":\"ses_parent\",\"rootSessionId\":\"ses_root\",\"purpose\":\"delegated_work\",\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"target_known\",\"targetLocalId\":\"ses_parent\",\"evidence\":\"native_typed\"}],\"childSessions\":[{\"id\":\"ses_helper1\",\"startTime\":\"2026-09-08T00:01:00Z\"},{\"id\":\"ses_helper2\",\"startTime\":\"2026-09-08T00:02:00Z\"}],\"turns\":[{\"index\":0,\"role\":\"user\",\"entryType\":\"text\",\"content\":\"fix parser\",\"timestamp\":\"2026-09-08T00:00:00Z\",\"depth\":0,\"sourceEntryRef\":\"e_u1\",\"provenance\":{\"origin\":\"submitted_input\",\"actor\":\"unknown\",\"delivery\":\"session_admission\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"text\",\"submissionRef\":\"s_u1\"}},{\"index\":1,\"role\":\"system\",\"entryType\":\"text\",\"content\":\"injected AGENTS content\",\"timestamp\":\"2026-09-08T00:00:01Z\",\"depth\":0,\"sourceEntryRef\":\"e_ctx1\",\"provenance\":{\"origin\":\"harness_context\",\"actor\":\"harness\",\"delivery\":\"system_lifecycle\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"none\"}},{\"index\":2,\"role\":\"user\",\"entryType\":\"text\",\"content\":\"canonical media display content\",\"timestamp\":\"2026-09-08T00:00:02Z\",\"depth\":0,\"sourceEntryRef\":\"e_media1\",\"provenance\":{\"origin\":\"submitted_input\",\"actor\":\"unknown\",\"delivery\":\"session_admission\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"media\",\"submissionRef\":\"s_u1\"}},{\"index\":3,\"role\":\"assistant\",\"entryType\":\"text\",\"content\":\"I will inspect.\",\"timestamp\":\"2026-09-08T00:00:03Z\",\"depth\":0,\"sourceEntryRef\":\"e_a1\",\"provenance\":{\"origin\":\"agent_output\",\"actor\":\"agent_delegate\",\"delivery\":\"session_admission\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"none\"},\"toolCalls\":[{\"id\":\"call_1\",\"name\":\"rg\",\"arguments\":\"rg parser --glob *.go\",\"result\":\"parser.go:12 retained result bytes\",\"callEntryRef\":\"e_call1\",\"resultEntryRef\":\"e_result1\",\"callProvenance\":{\"origin\":\"tool_activity\",\"actor\":\"agent_delegate\",\"delivery\":\"tool_delivery\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"none\"},\"resultProvenance\":{\"origin\":\"tool_activity\",\"actor\":\"agent_delegate\",\"delivery\":\"tool_delivery\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"none\"}}]},{\"index\":4,\"role\":\"assistant\",\"entryType\":\"thinking\",\"content\":\"I will inspect.\",\"timestamp\":\"2026-09-08T00:00:04Z\",\"depth\":0,\"sourceEntryRef\":\"e_reason1\",\"provenance\":{\"origin\":\"agent_output\",\"actor\":\"agent_delegate\",\"delivery\":\"session_admission\",\"ownership\":\"local\",\"evidence\":\"native_typed\",\"inputModality\":\"none\"}}],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"user\",\"entryType\":\"text\",\"content\":\"unique old child question\",\"timestamp\":\"2026-09-07T00:00:00Z\",\"depth\":0,\"sourceEntryRef\":\"e_old\",\"provenance\":{\"origin\":\"submitted_input\",\"actor\":\"unknown\",\"delivery\":\"inherited_context\",\"ownership\":\"uncertain\",\"evidence\":\"retained_last_good\",\"inputModality\":\"text\",\"submissionRef\":\"s_old\"}}]}]}"
          },
          "expected": {
            "turn_count": 5,
            "input_present": true,
            "input_count": 1,
            "helper_count": 2,
            "main_refs": [
              "e_u1",
              "e_ctx1",
              "e_media1",
              "e_a1",
              "e_call1",
              "e_result1",
              "e_reason1"
            ],
            "earlier_refs": [
              "e_old"
            ]
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "folded-round-trip"
          },
          "mutation": {
            "description": "durable graph evidence is populated"
          }
        }
      ]
    },
    "counts": {
      "cases": [
        {
          "name": "input-count-absent",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_absent\",\"harness\":\"claude-code\",\"turnCount\":0,\"turns\":[]}"
          },
          "expected": {
            "present": false
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "input-count-absent-zero-positive"
          },
          "mutation": {
            "description": "historical count is omitted"
          }
        },
        {
          "name": "input-count-zero",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_zero\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":0,\"turns\":[]}"
          },
          "expected": {
            "present": true,
            "value": 0
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "input-count-absent-zero-positive"
          },
          "mutation": {
            "description": "measured zero is present"
          }
        },
        {
          "name": "input-count-positive",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_one\",\"harness\":\"claude-code\",\"turnCount\":5,\"inputSubmissionCount\":1,\"turns\":[]}"
          },
          "expected": {
            "present": true,
            "value": 1
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "three-independent-counts"
          },
          "mutation": {
            "description": "one input remains independent from five turns"
          }
        },
        {
          "name": "input-count-safe-max",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_max\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":9007199254740991,\"turns\":[]}"
          },
          "expected": {
            "present": true,
            "value": 9007199254740991
          },
          "classification": "must-pass",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "exact JS-safe upper bound"
          }
        }
      ]
    },
    "invalid_counts": {
      "cases": [
        {
          "name": "input-count-null",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":null,\"turns\":[]}"
          },
          "expected": {
            "errorContains": "explicit null"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "null replaces omission"
          }
        },
        {
          "name": "input-count-string",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":\"1\",\"turns\":[]}"
          },
          "expected": {
            "errorContains": "inputSubmissionCount"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is a string"
          }
        },
        {
          "name": "input-count-negative",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":-1,\"turns\":[]}"
          },
          "expected": {
            "errorContains": "outside"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is negative"
          }
        },
        {
          "name": "input-count-fraction",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":1.5,\"turns\":[]}"
          },
          "expected": {
            "errorContains": "inputSubmissionCount"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is fractional"
          }
        },
        {
          "name": "input-count-above-safe-max",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":9007199254740992,\"turns\":[]}"
          },
          "expected": {
            "errorContains": "outside"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count exceeds JS-safe range"
          }
        },
        {
          "name": "input-count-bool",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":true,\"turns\":[]}"
          },
          "expected": {
            "errorContains": "inputSubmissionCount"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is boolean"
          }
        },
        {
          "name": "input-count-array",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":[],\"turns\":[]}"
          },
          "expected": {
            "errorContains": "inputSubmissionCount"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is array"
          }
        },
        {
          "name": "input-count-object",
          "input": {
            "json": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"toolCallCount\":0,\"id\":\"ses_bad\",\"harness\":\"claude-code\",\"turnCount\":0,\"inputSubmissionCount\":{},\"turns\":[]}"
          },
          "expected": {
            "errorContains": "inputSubmissionCount"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "boundary",
            "ref": "input-count-invalid"
          },
          "mutation": {
            "description": "count is object"
          }
        }
      ]
    },
    "mirrors": {
      "cases": [
        {
          "name": "authoritative-mirrors-match",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "detail_root": "ses_root",
            "metadata_root": "ses_root",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-mirror"
          },
          "mutation": {
            "description": "separately allocated equal mirrors share one authority"
          }
        },
        {
          "name": "authoritative-count-disagreement",
          "input": {
            "detail_count": 1,
            "metadata_count": 0,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-mirror"
          },
          "mutation": {
            "description": "metadata count differs"
          }
        },
        {
          "name": "authoritative-count-detail-absent",
          "input": {
            "metadata_count": 0,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-mirror"
          },
          "mutation": {
            "description": "detail count presence differs"
          }
        },
        {
          "name": "authoritative-count-metadata-absent",
          "input": {
            "detail_count": 0,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-mirror"
          },
          "mutation": {
            "description": "metadata count presence differs"
          }
        },
        {
          "name": "authoritative-count-both-absent",
          "input": {
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "detail_turn_count": 4,
            "metadata_turn_count": 4
          },
          "expected": {
            "accept": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-mirror"
          },
          "mutation": {
            "description": "historical count remains unknown"
          }
        },
        {
          "name": "authoritative-root-values-differ",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "detail_root": "ses_root",
            "metadata_root": "ses_other",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "graph mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-root-mirror"
          },
          "mutation": {
            "description": "root values differ"
          }
        },
        {
          "name": "authoritative-root-detail-absent",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "metadata_root": "ses_root",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "graph mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-root-mirror"
          },
          "mutation": {
            "description": "detail root is absent"
          }
        },
        {
          "name": "authoritative-root-metadata-absent",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "detail_root": "ses_root",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "graph mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "authoritative-root-mirror"
          },
          "mutation": {
            "description": "metadata root is absent"
          }
        },
        {
          "name": "legacy-turn-count-local4-export5",
          "input": {
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "detail_turn_count": 5,
            "metadata_turn_count": 4
          },
          "expected": {
            "accept": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "legacy-turn-count-control"
          },
          "mutation": {
            "description": "legacy per-surface totals remain distinct"
          }
        },
        {
          "name": "projected-main-count-disagreement",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_parent",
            "graph_parent": "ses_parent",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 4
          },
          "expected": {
            "accept": false,
            "errorContains": "mirrors disagree"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "legacy-turn-count-control"
          },
          "mutation": {
            "description": "projected main count differs"
          }
        },
        {
          "name": "graph-legacy-parent-mismatch",
          "input": {
            "detail_count": 1,
            "metadata_count": 1,
            "parent": "ses_other",
            "graph_parent": "ses_parent",
            "projected_main_count": true,
            "detail_turn_count": 5,
            "metadata_turn_count": 5
          },
          "expected": {
            "accept": false,
            "errorContains": "legacy parentSessionId disagrees"
          },
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "graph-parent-mismatch"
          },
          "mutation": {
            "description": "legacy parent differs from durable graph"
          }
        }
      ]
    },
    "digest": {
      "cases": [
        {
          "name": "digest-absent-versus-zero",
          "input": {
            "left_present": false,
            "right_present": true,
            "right_value": 0
          },
          "expected": {
            "different": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "digest-presence"
          },
          "mutation": {
            "description": "measured zero is added"
          }
        },
        {
          "name": "digest-same-positive",
          "input": {
            "left_present": true,
            "left_value": 1,
            "right_present": true,
            "right_value": 1
          },
          "expected": {
            "different": false
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "digest-presence"
          },
          "mutation": {
            "description": "equal authoritative count is replayed"
          }
        },
        {
          "name": "digest-root-evidence",
          "input": {
            "right_variant": "root"
          },
          "expected": {
            "different": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "graph-digest"
          },
          "mutation": {
            "description": "root evidence is added"
          }
        },
        {
          "name": "digest-purpose-evidence",
          "input": {
            "right_variant": "purpose"
          },
          "expected": {
            "different": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "graph-digest"
          },
          "mutation": {
            "description": "purpose evidence is added"
          }
        },
        {
          "name": "digest-relationship-evidence",
          "input": {
            "right_variant": "relationship"
          },
          "expected": {
            "different": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "graph-digest"
          },
          "mutation": {
            "description": "relationship evidence is added"
          }
        },
        {
          "name": "digest-entry-provenance",
          "input": {
            "right_variant": "entry_provenance"
          },
          "expected": {
            "different": true
          },
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "graph-digest"
          },
          "mutation": {
            "description": "entry provenance is added"
          }
        }
      ]
    },
    "requiredNames": [
      "legacy-durable-decode",
      "three-independent-counts",
      "input-count-absent",
      "input-count-zero",
      "input-count-positive",
      "input-count-safe-max",
      "input-count-null",
      "input-count-string",
      "input-count-negative",
      "input-count-fraction",
      "input-count-above-safe-max",
      "input-count-bool",
      "input-count-array",
      "input-count-object",
      "authoritative-mirrors-match",
      "authoritative-count-disagreement",
      "authoritative-count-detail-absent",
      "authoritative-count-metadata-absent",
      "authoritative-count-both-absent",
      "authoritative-root-values-differ",
      "authoritative-root-detail-absent",
      "authoritative-root-metadata-absent",
      "legacy-turn-count-local4-export5",
      "projected-main-count-disagreement",
      "graph-legacy-parent-mismatch",
      "digest-absent-versus-zero",
      "digest-same-positive",
      "digest-root-evidence",
      "digest-purpose-evidence",
      "digest-relationship-evidence",
      "digest-entry-provenance"
    ]
  },
  "grouped_read": {
    "requiredNames": [
      "nested-local-owner",
      "nested-village-owner",
      "nested-local-leaf",
      "nested-village-leaf",
      "local-member-context-rejected",
      "village-member-context-rejected",
      "local-member-both-arms-rejected",
      "village-member-both-arms-rejected",
      "local-duplicate-member",
      "village-duplicate-member",
      "local-duplicate-nested-group",
      "village-duplicate-nested-group",
      "local-empty-page-zero-total",
      "village-empty-page-zero-total",
      "local-null-members",
      "village-null-members",
      "local-null-total",
      "village-null-total",
      "local-unsafe-total",
      "village-unsafe-total",
      "local-member-over-limit",
      "village-member-over-limit",
      "local-count-zero-preserved",
      "village-count-zero-preserved",
      "local-count-null-rejected",
      "village-count-null-rejected",
      "local-count-absent-preserved",
      "village-count-absent-preserved",
      "local-group-count-negative",
      "village-group-count-negative",
      "local-member-missing-transcript",
      "village-member-missing-transcript",
      "legacy-group-missing-prompts",
      "group-present-false-prompts",
      "group-present-true-prompts",
      "group-invalid-present-mode",
      "canonical-group-required-prompts-control",
      "canonical-group-present-prompts-control",
      "content-resolved-envelope",
      "content-bare-detail-rejected",
      "nested-local-root-counts",
      "nested-village-root-counts",
      "local-second-member-page",
      "village-second-member-page",
      "flat-detail-legacy-group",
      "grouped-detail-legacy-group",
      "flat-detail-present-prompts",
      "grouped-detail-present-prompts",
      "local-null-nested-groups",
      "village-null-nested-groups",
      "local-null-search-matches",
      "local-group-id-across-members",
      "village-group-id-across-members",
      "local-zero-nested-group-count",
      "village-zero-nested-group-count",
      "local-null-nested-group-count",
      "village-null-nested-group-count",
      "local-sync-member-mirrors",
      "local-sync-member-count-mismatch",
      "local-search-member-mirrors",
      "local-search-member-identity-mismatch",
      "village-collective-member-mirrors",
      "village-collective-member-count-mismatch",
      "village-pending-member-mirrors",
      "village-pending-member-identity-mismatch",
      "village-my-share-member-mirrors",
      "village-my-share-member-count-mismatch",
      "village-contributable-member-mirrors",
      "village-contributable-member-identity-mismatch"
    ],
    "cases": {
      "cases": [
        {
          "name": "nested-local-owner",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "P expansion returns G1 with its own group, direct total one rather than descendant total two"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "nested-village-owner",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public P expansion preserves G1 nested group and all nullable summary facts"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "nested-local-leaf",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "independent nested scope returns only G2 on the next expansion"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g2",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "nested-village-leaf",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public nested scope returns one stable G2 identity"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174004",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-member-context-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "a list-only context container cannot count as a saved helper member"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "context_container",
                  "context": {
                    "groupId": "hg_g1",
                    "ownerStatus": "unknown"
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-member-context-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public member pages reject list-only containers"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "context_container",
                  "context": {
                    "groupId": "hg_g1",
                    "ownerStatus": "unknown"
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-member-both-arms-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "member cannot carry a second context arm"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ],
                  "context": {
                    "groupId": "hg_g1",
                    "ownerStatus": "unknown"
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-member-both-arms-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public member cannot carry a second context arm"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ],
                  "context": {
                    "groupId": "hg_g1",
                    "ownerStatus": "unknown"
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-duplicate-member",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "duplicate saved identity cannot increase direct total"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 2,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-duplicate-member",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "duplicate public identity cannot increase direct total"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 2,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-duplicate-nested-group",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "duplicate group id cannot own independent disclosure state"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    },
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-duplicate-nested-group",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "duplicate public nested group rejected"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    },
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-empty-page-zero-total",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "no matching authorized saved helpers yields an empty page with measured zero total"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [],
              "page": 2,
              "limit": 1,
              "total": 0
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-empty-page-zero-total",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public empty scope preserves measured zero"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [],
              "page": 2,
              "limit": 1,
              "total": 0
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-null-members",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "members null is not an empty page"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": null,
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-null-members",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public members null is not an empty page"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": null,
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-null-total",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "null total cannot coerce to measured zero"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [],
              "page": 1,
              "limit": 1,
              "total": null
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-null-total",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public null total cannot coerce to zero"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [],
              "page": 1,
              "limit": 1,
              "total": null
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-unsafe-total",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "unsafe integer total cannot be preserved in JavaScript"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 9007199254740992
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-unsafe-total",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public unsafe integer total rejected"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 9007199254740992
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-member-over-limit",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "two distinct helper rows cannot fit a one-row page"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g2",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-member-over-limit",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "two public helper identities cannot fit a one-row page"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174004",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-count-zero-preserved",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "member retains measured zero input count independently of five turns and one nested helper"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 0
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-count-zero-preserved",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public member preserves zero input count and independent turn and helper counts"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent",
                      "input_submission_count": 0
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-count-null-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "null input count is invalid rather than unknown"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": null
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-count-null-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public null input count must not disappear"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent",
                      "input_submission_count": null
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-count-absent-preserved",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "historical local summary omits unknown input count"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0
                    }
                  }
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-count-absent-preserved",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "historical public member retains unknown input count as absent"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-group-count-negative",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "nested helper count is never negative"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": -1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-group-count-negative",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public nested helper count is never negative"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": -1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-member-missing-transcript",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "member discriminator requires a saved row"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript"
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-member-missing-transcript",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public discriminator requires a saved row"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript"
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "legacy-group-missing-prompts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "eleven released base fields do not fabricate unavailable prompts settings"
          },
          "input": {
            "schema": "VillageGroupDetailRecord",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice"
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "group-present-false-prompts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "explicit false is a supplied fact and must remain present"
          },
          "input": {
            "schema": "VillageGroupDetailRecord",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice",
              "post_prompts_check": false,
              "prompts_check_mode": "informational"
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "group-present-true-prompts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "source-backed true setting remains present"
          },
          "input": {
            "schema": "VillageGroupDetailRecord",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice",
              "post_prompts_check": true,
              "prompts_check_mode": "required"
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "group-invalid-present-mode",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "optional present mode still uses the closed menu"
          },
          "input": {
            "schema": "VillageGroupDetailRecord",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice",
              "prompts_check_mode": "invented"
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "canonical-group-required-prompts-control",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "independent canonical prompts contract remains required, not weakened by read compatibility"
          },
          "input": {
            "schema": "VillageGroup",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice"
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "canonical-group-present-prompts-control",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "canonical group still accepts its supplied prompts configuration"
          },
          "input": {
            "schema": "VillageGroup",
            "payload": {
              "id": "123e4567-e89b-12d3-a456-426614174001",
              "name": "collective",
              "description": null,
              "created_by": "123e4567-e89b-12d3-a456-426614174001",
              "created_at": "2026-09-08T00:00:00Z",
              "updated_at": "2026-09-08T00:00:00Z",
              "acceptance_mode": "open",
              "data_access": "public",
              "linked_github_org": null,
              "display_members": false,
              "transcript_deletion_policy": "user_choice",
              "post_prompts_check": false,
              "prompts_check_mode": "informational"
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "content-resolved-envelope",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "content response resolves to the durable envelope rather than a bare detail"
          },
          "input": {
            "schema": "TranscriptContent",
            "route": "/api/v1/transcripts/{id}/content",
            "payload": {
              "contractVersion": "1",
              "kind": "session_detail",
              "sessionDetail": {
                "id": "g1",
                "harness": "codex",
                "startTime": "2026-09-08T00:00:00Z",
                "endTime": "2026-09-08T00:00:00Z",
                "durationMins": 0,
                "totalTokens": 0,
                "tokensIn": 0,
                "tokensOut": 0,
                "turnCount": 0,
                "toolCallCount": 0,
                "turns": []
              }
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "content-bare-detail-rejected",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "a root detail cannot satisfy the actual content operation envelope schema"
          },
          "input": {
            "schema": "TranscriptContent",
            "route": "/api/v1/transcripts/{id}/content",
            "payload": {
              "id": "g1",
              "harness": "codex",
              "startTime": "2026-09-08T00:00:00Z",
              "endTime": "2026-09-08T00:00:00Z",
              "durationMins": 0,
              "totalTokens": 0,
              "tokensIn": 0,
              "tokensOut": 0,
              "turnCount": 0,
              "toolCallCount": 0,
              "turns": []
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "nested-local-root-counts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "P has one direct helper; total saved helper identities includes nested G2 without another top container"
          },
          "input": {
            "schema": "LocalSessionListPayload",
            "payload": {
              "items": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "p",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_p",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_p"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "totalItems": 1,
              "ordinarySessionTotal": 1,
              "helperThreadTotal": 2
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "nested-village-root-counts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public P root retains ordinary one, helpers two, top items one and direct group one"
          },
          "input": {
            "schema": "VillageSessionListPayload",
            "payload": {
              "items": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174005",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_p",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_p"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "totalItems": 1,
              "ordinarySessionTotal": 1,
              "helperThreadTotal": 2
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-second-member-page",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "second page retains full direct total and one loaded member with nested group"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 2,
              "limit": 1,
              "total": 2
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-second-member-page",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public second page retains direct total independent of loaded descendants"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 2,
              "limit": 1,
              "total": 2
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "flat-detail-legacy-group",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "omitted-view GET detail preserves the eleven group fields without invented prompts settings"
          },
          "input": {
            "schema": "VillageGroupDetailResponse",
            "route": "/api/v1/groups/{id}",
            "payload": {
              "group": {
                "id": "123e4567-e89b-12d3-a456-426614174001",
                "name": "collective",
                "description": null,
                "created_by": "123e4567-e89b-12d3-a456-426614174001",
                "created_at": "2026-09-08T00:00:00Z",
                "updated_at": "2026-09-08T00:00:00Z",
                "acceptance_mode": "open",
                "data_access": "public",
                "linked_github_org": null,
                "display_members": false,
                "transcript_deletion_policy": "user_choice"
              },
              "members": [],
              "stats": {
                "total_transcripts": 0,
                "contributor_count": 0,
                "total_turns": 0,
                "total_duration_ms": 0,
                "total_tokens": 0
              },
              "models": [],
              "contributors": [],
              "can_read": true,
              "your_role": "owner",
              "transcripts": []
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "grouped-detail-legacy-group",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "grouped GET detail preserves identical base group values and omitted settings"
          },
          "input": {
            "schema": "VillageGroupedGroupDetailResponse",
            "route": "/api/v1/groups/{id}",
            "payload": {
              "group": {
                "id": "123e4567-e89b-12d3-a456-426614174001",
                "name": "collective",
                "description": null,
                "created_by": "123e4567-e89b-12d3-a456-426614174001",
                "created_at": "2026-09-08T00:00:00Z",
                "updated_at": "2026-09-08T00:00:00Z",
                "acceptance_mode": "open",
                "data_access": "public",
                "linked_github_org": null,
                "display_members": false,
                "transcript_deletion_policy": "user_choice"
              },
              "members": [],
              "stats": {
                "total_transcripts": 0,
                "contributor_count": 0,
                "total_turns": 0,
                "total_duration_ms": 0,
                "total_tokens": 0
              },
              "models": [],
              "contributors": [],
              "can_read": true,
              "your_role": "owner",
              "transcriptList": {
                "items": [],
                "page": 1,
                "limit": 1,
                "totalItems": 0,
                "ordinarySessionTotal": 0,
                "helperThreadTotal": 0
              }
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "flat-detail-present-prompts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "flat GET detail retains source-backed false and informational mode"
          },
          "input": {
            "schema": "VillageGroupDetailResponse",
            "route": "/api/v1/groups/{id}",
            "payload": {
              "group": {
                "id": "123e4567-e89b-12d3-a456-426614174001",
                "name": "collective",
                "description": null,
                "created_by": "123e4567-e89b-12d3-a456-426614174001",
                "created_at": "2026-09-08T00:00:00Z",
                "updated_at": "2026-09-08T00:00:00Z",
                "acceptance_mode": "open",
                "data_access": "public",
                "linked_github_org": null,
                "display_members": false,
                "transcript_deletion_policy": "user_choice",
                "post_prompts_check": false,
                "prompts_check_mode": "informational"
              },
              "members": [],
              "stats": {
                "total_transcripts": 0,
                "contributor_count": 0,
                "total_turns": 0,
                "total_duration_ms": 0,
                "total_tokens": 0
              },
              "models": [],
              "contributors": [],
              "can_read": true,
              "your_role": "owner",
              "transcripts": []
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "grouped-detail-present-prompts",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "grouped GET detail retains the same explicit false settings as flat reads"
          },
          "input": {
            "schema": "VillageGroupedGroupDetailResponse",
            "route": "/api/v1/groups/{id}",
            "payload": {
              "group": {
                "id": "123e4567-e89b-12d3-a456-426614174001",
                "name": "collective",
                "description": null,
                "created_by": "123e4567-e89b-12d3-a456-426614174001",
                "created_at": "2026-09-08T00:00:00Z",
                "updated_at": "2026-09-08T00:00:00Z",
                "acceptance_mode": "open",
                "data_access": "public",
                "linked_github_org": null,
                "display_members": false,
                "transcript_deletion_policy": "user_choice",
                "post_prompts_check": false,
                "prompts_check_mode": "informational"
              },
              "members": [],
              "stats": {
                "total_transcripts": 0,
                "contributor_count": 0,
                "total_turns": 0,
                "total_duration_ms": 0,
                "total_tokens": 0
              },
              "models": [],
              "contributors": [],
              "can_read": true,
              "your_role": "owner",
              "transcriptList": {
                "items": [],
                "page": 1,
                "limit": 1,
                "totalItems": 0,
                "ordinarySessionTotal": 0,
                "helperThreadTotal": 0
              }
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-null-nested-groups",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "optional groups must be an array when supplied"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": null
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-null-nested-groups",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public optional groups reject explicit null"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": null
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-null-search-matches",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "optional search evidence must be an array when supplied"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    },
                    "matches": null
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-group-id-across-members",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "distinct saved members cannot both claim the same immediate-owner group"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g2",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 2,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-group-id-across-members",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "distinct public members cannot both claim one group"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                },
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174004",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 2,
              "total": 2
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-zero-nested-group-count",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "measured zero nested matches remains distinct from missing count"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 0,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-zero-nested-group-count",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public measured zero nested count remains zero"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 0,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-null-nested-group-count",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "null cannot become zero nested helper matches"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": null,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-null-nested-group-count",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "public null nested count cannot become zero"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": null,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-sync-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "nested local helper keeps the originating sync status and independent matching counts"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1,
                      "project": "example",
                      "projectHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
                    },
                    "sync": {
                      "id": "g1",
                      "harness": "codex",
                      "projectName": "example",
                      "projectHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "hostSlug": "host",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMs": 60000,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "inputSubmissionCount": 1,
                      "model": "codex",
                      "syncStatus": "synced"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-sync-member-count-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "sync member input count cannot disagree with its saved session"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1,
                      "project": "example",
                      "projectHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
                    },
                    "sync": {
                      "id": "g1",
                      "harness": "codex",
                      "projectName": "example",
                      "projectHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "hostSlug": "host",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMs": 60000,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "inputSubmissionCount": 0,
                      "model": "codex",
                      "syncStatus": "synced"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "local-search-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "helper search expansion retains only its own matched content evidence"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    },
                    "matches": [
                      {
                        "sessionId": "g1",
                        "project": "example",
                        "entryIndex": 0,
                        "role": "user",
                        "snippet": "query",
                        "score": 1
                      }
                    ]
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "local-search-member-identity-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "a search match from another helper cannot be attached to this member"
          },
          "input": {
            "schema": "LocalHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "g1",
                      "harness": "codex",
                      "startTime": "2026-09-08T00:00:00Z",
                      "durationMins": 1,
                      "totalTokens": 13,
                      "turnCount": 5,
                      "toolCallCount": 0,
                      "inputSubmissionCount": 1
                    },
                    "matches": [
                      {
                        "sessionId": "g2",
                        "project": "example",
                        "entryIndex": 0,
                        "role": "user",
                        "snippet": "query",
                        "score": 1
                      }
                    ]
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-collective-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "collective helper expansion preserves its owner usage summary and nested group"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "collective": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent",
                      "owner_username": "owner",
                      "owner_avatar_url": null,
                      "owner_is_discoverable": true
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-collective-member-count-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "collective helper count cannot change unknown into measured zero"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "collective": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent",
                      "owner_username": "owner",
                      "owner_avatar_url": null,
                      "owner_is_discoverable": true,
                      "input_submission_count": 0
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-pending-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "pending route uses transcript_id and branch aliases without losing member context"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "pending": {
                      "transcript_id": "123e4567-e89b-12d3-a456-426614174003",
                      "title": null,
                      "model_provider": "codex",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "parent_session_id": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "branch": null,
                      "owner_username": "owner",
                      "owner_is_discoverable": true,
                      "shared_at": "2026-09-08T00:00:00Z"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-pending-member-identity-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "pending alias cannot target another public transcript"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "pending": {
                      "transcript_id": "123e4567-e89b-12d3-a456-426614174004",
                      "title": null,
                      "model_provider": "codex",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "parent_session_id": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "branch": null,
                      "owner_username": "owner",
                      "owner_is_discoverable": true,
                      "shared_at": "2026-09-08T00:00:00Z"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-my-share-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "my-shares helper keeps individual review status and usage mirrors"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "myShare": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "title": null,
                      "model_provider": "codex",
                      "model_name": null,
                      "visibility": "public",
                      "published_at": "2026-09-08T00:00:00Z",
                      "turn_count": 5,
                      "tokens_in": null,
                      "tokens_out": null,
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "parent_session_id": null,
                      "status": "approved",
                      "shared_at": "2026-09-08T00:00:00Z"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-my-share-member-count-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "my-shares turn count cannot replace the stored member total"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "myShare": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "title": null,
                      "model_provider": "codex",
                      "model_name": null,
                      "visibility": "public",
                      "published_at": "2026-09-08T00:00:00Z",
                      "turn_count": 1,
                      "tokens_in": null,
                      "tokens_out": null,
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "parent_session_id": null,
                      "status": "approved",
                      "shared_at": "2026-09-08T00:00:00Z"
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        },
        {
          "name": "village-contributable-member-mirrors",
          "classification": "must-pass",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "contribution expansion preserves individual eligibility and project context"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "contributable": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "visibility": "public",
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "git_branch": null,
                      "parent_session_id": null,
                      "session_origin": "agent",
                      "model_provider": "codex",
                      "published_at": "2026-09-08T00:00:00Z",
                      "already_shared": false
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": true
          }
        },
        {
          "name": "village-contributable-member-identity-mismatch",
          "classification": "must-fail",
          "provenance": {
            "source": "requirement",
            "ref": "immediate-owner grouped read contract"
          },
          "mutation": {
            "description": "contribution row cannot claim another local helper identity"
          },
          "input": {
            "schema": "VillageHelperMembersPayload",
            "payload": {
              "members": [
                {
                  "kind": "transcript",
                  "transcript": {
                    "session": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "owner_id": "123e4567-e89b-12d3-a456-426614174001",
                      "local_id": "123e4567-e89b-12d3-a456-426614174002",
                      "title": null,
                      "description": null,
                      "visibility": "public",
                      "model_provider": "codex",
                      "model_name": null,
                      "harness_version": null,
                      "session_start": null,
                      "session_end": null,
                      "turn_count": 5,
                      "token_count": 13,
                      "blob_size_bytes": null,
                      "schema_version": "11",
                      "published_at": "2026-09-08T00:00:00Z",
                      "updated_at": "2026-09-08T00:00:00Z",
                      "parent_session_id": null,
                      "ingested_at": null,
                      "source_format": null,
                      "git_branch": null,
                      "git_remote": null,
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_name": null,
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "project_remote_label": "",
                      "tool_call_count": null,
                      "subagent_count": null,
                      "duration_ms": null,
                      "subagents": null,
                      "diagnostics_warnings": null,
                      "diagnostics_partial": null,
                      "tokens_in": null,
                      "tokens_out": null,
                      "title_generated": null,
                      "outcome": null,
                      "files_touched": null,
                      "lines_changed": null,
                      "retry_loops": null,
                      "retry_tokens_wasted": null,
                      "within_session_reverts": null,
                      "signal_density": null,
                      "spec_quality_score": null,
                      "exploration_ratio": null,
                      "scope_breadth": null,
                      "discovery_turns": null,
                      "m2_token_outcome_ratio": null,
                      "m3_unique_tool_count": null,
                      "m4_error_recovery_count": null,
                      "m4_consecutive_error_max": null,
                      "m5_context_utilization_pct": null,
                      "m5_peak_context_tokens": null,
                      "m5_avg_message_tokens": null,
                      "m6_output_survival_pct": null,
                      "m6_lines_survived": null,
                      "m6_lines_total": null,
                      "m7_spec_word_count": null,
                      "m7_spec_has_examples": null,
                      "m7_spec_has_constraints": null,
                      "computed_at": null,
                      "compute_version": null,
                      "content_hash": null,
                      "license_id": null,
                      "session_origin": "agent"
                    },
                    "contributable": {
                      "id": "123e4567-e89b-12d3-a456-426614174003",
                      "local_id": "123e4567-e89b-12d3-a456-426614174004",
                      "title": null,
                      "visibility": "public",
                      "project_hash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                      "project_display_name": "example",
                      "project_name_source": "privacy",
                      "git_branch": null,
                      "parent_session_id": null,
                      "session_origin": "agent",
                      "model_provider": "codex",
                      "published_at": "2026-09-08T00:00:00Z",
                      "already_shared": false
                    }
                  },
                  "helperGroups": [
                    {
                      "groupId": "hg_g1",
                      "purpose": "helper_review",
                      "helperThreadCount": 1,
                      "memberScope": "scope_g1"
                    }
                  ]
                }
              ],
              "page": 1,
              "limit": 1,
              "total": 1
            }
          },
          "expected": {
            "valid": false
          }
        }
      ]
    }
  }
} as const;
