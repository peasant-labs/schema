// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
import type { TimelineFixtureCorpus } from "../../fixtures/timeline.js";

export const canonicalTimelineFixtures: TimelineFixtureCorpus = {
  "cases": [
    {
      "family": "many-to-many-bindings",
      "name": "many_to_many_bindings",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "first task",
            "harness": "claude-code",
            "startMs": 2000,
            "hasCommitBinding": true
          },
          {
            "sessionId": "session-b",
            "title": "second task",
            "harness": "codex",
            "startMs": 1000,
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-2",
            "subject": "follow-up",
            "hasSession": true,
            "sessionIds": [
              "session-a",
              "session-b"
            ],
            "associations": [
              {
                "id": "assoc-commit-2-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-2"
                  }
                ]
              },
              {
                "id": "assoc-commit-2-b",
                "sessionId": "session-b",
                "conclusion": "candidate",
                "confidence": "medium",
                "evidence": [
                  {
                    "kind": "touched_file",
                    "touchedFilePath": "src/main.go"
                  }
                ]
              }
            ]
          },
          {
            "hash": "commit-1",
            "subject": "setup",
            "hasSession": true,
            "sessionIds": [
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-commit-1-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-1"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "authoritative many-to-many commit bindings"
      },
      "mutation": {
        "description": "binds two sessions to one commit and one session to two commits"
      }
    },
    {
      "family": "unattached-session",
      "name": "session_not_linked_to_commit",
      "input": {
        "sessions": [
          {
            "sessionId": "session-unlinked",
            "title": "exploration",
            "harness": "opencode",
            "startMs": 3000,
            "hasCommitBinding": false
          }
        ],
        "commits": [
          {
            "hash": "commit-empty",
            "subject": "manual change",
            "hasSession": false,
            "sessionIds": [],
            "associations": []
          }
        ]
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "unattached sessions remain discoverable"
      },
      "mutation": {
        "description": "includes a visible session with no commit association"
      }
    },
    {
      "family": "binding-outside-window",
      "name": "binding_outside_default_branch_window",
      "input": {
        "sessions": [
          {
            "sessionId": "session-old",
            "title": "older recorded work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-recent",
            "subject": "recent manual change",
            "hasSession": false,
            "sessionIds": [],
            "associations": []
          }
        ]
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "complete binding truth exceeds bounded commit window"
      },
      "mutation": {
        "description": "keeps binding truth when the associated commit is outside the returned window"
      }
    },
    {
      "family": "rewrite-ledger-bound-session",
      "name": "rewrite_ledger_references_bound_session",
      "input": {
        "sessions": [
          {
            "sessionId": "session-rewrite-bound",
            "title": "squash repair",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-bound",
            "subject": "squash successor",
            "hasSession": false,
            "sessionIds": [],
            "associations": []
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-bound",
            "subject": "squash repair",
            "sessionIds": [
              "session-rewrite-bound"
            ],
            "associations": [
              {
                "id": "assoc-ghost-bound",
                "sessionId": "session-rewrite-bound",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-bound"
                  }
                ]
              }
            ],
            "successorHash": "successor-bound",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "rewrite ledger references require authoritative binding truth"
      },
      "mutation": {
        "description": "references a bound timeline session from the rewrite ledger and names a visible successor"
      }
    },
    {
      "family": "rewrite-ledger-binding-truth",
      "name": "rewrite_ledger_reference_requires_binding_truth",
      "input": {
        "sessions": [
          {
            "sessionId": "session-rewrite-unbound",
            "title": "stale ledger",
            "harness": "codex",
            "hasCommitBinding": false
          }
        ],
        "commits": [
          {
            "hash": "successor-unbound",
            "subject": "stale successor",
            "hasSession": false,
            "sessionIds": [],
            "associations": []
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-unbound",
            "subject": "stale ledger",
            "sessionIds": [
              "session-rewrite-unbound"
            ],
            "associations": [
              {
                "id": "assoc-ghost-unbound",
                "sessionId": "session-rewrite-unbound",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-unbound"
                  }
                ]
              }
            ],
            "successorHash": "successor-unbound",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "hasCommitBinding=false",
        "repair": {
          "kind": "set_session_binding_true",
          "sessionId": "session-rewrite-unbound",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "bug",
        "ref": "ReviewListPayload rewrite ledger references must imply HasCommitBinding"
      },
      "mutation": {
        "description": "points a rewrite ledger row at an existing timeline session whose binding truth is false"
      }
    },
    {
      "family": "rewrite-successor-association-identity",
      "name": "rewrite_successor_preserves_association_identity",
      "input": {
        "sessions": [
          {
            "sessionId": "session-rewrite-preserved",
            "title": "preserved relationship",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-preserved",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-rewrite-preserved"
            ],
            "associations": [
              {
                "id": "assoc-rewrite-preserved",
                "sessionId": "session-rewrite-preserved",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-preserved"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-preserved",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-rewrite-preserved"
            ],
            "associations": [
              {
                "id": "assoc-rewrite-preserved",
                "sessionId": "session-rewrite-preserved",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-preserved"
                  }
                ]
              }
            ],
            "successorHash": "successor-preserved",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "displayed rewrite successors preserve the first-class association object when they represent the binding"
      },
      "mutation": {
        "description": "carries the same durable association ID, conclusion, confidence, and observation from the ledger to its displayed successor"
      }
    },
    {
      "family": "rewrite-ledger-missing-session",
      "name": "rewrite_ledger_references_missing_session",
      "input": {
        "sessions": [],
        "commits": [
          {
            "hash": "successor-missing-session",
            "subject": "stale successor",
            "hasSession": false,
            "sessionIds": [],
            "associations": []
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-missing-session",
            "subject": "stale ledger",
            "sessionIds": [
              "session-rewrite-missing"
            ],
            "associations": [
              {
                "id": "assoc-ghost-missing",
                "sessionId": "session-rewrite-missing",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-missing-session"
                  }
                ]
              }
            ],
            "successorHash": "successor-missing-session",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "references sessionId \"session-rewrite-missing\" but that session is absent from sessions"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "bug",
        "ref": "rewrite ledger session references must name timeline sessions"
      },
      "mutation": {
        "description": "points a rewrite ledger row at a session absent from the timeline session catalog"
      }
    },
    {
      "family": "non-default-branch-binding",
      "name": "binding_on_non_default_branch",
      "input": {
        "sessions": [
          {
            "sessionId": "session-branch",
            "title": "branch work",
            "harness": "codex",
            "hasCommitBinding": true
          }
        ],
        "commits": []
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "branch bindings are authoritative without inference"
      },
      "mutation": {
        "description": "keeps binding truth when no default-branch commit is visible"
      }
    },
    {
      "family": "stable-session-order",
      "name": "stable_equal_timestamp_and_missing_timestamp_order",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "first tie",
            "harness": "claude-code",
            "startMs": 4000,
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-b",
            "title": "second tie",
            "harness": "codex",
            "startMs": 4000,
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-c",
            "title": "unknown time",
            "harness": "opencode",
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {},
      "classification": "must-pass",
      "provenance": {
        "source": "requirement",
        "ref": "deterministic session timeline ordering"
      },
      "mutation": {
        "description": "combines timestamp ties and a missing timestamp in canonical order"
      }
    },
    {
      "family": "descending-timestamp-order",
      "name": "newer_timestamp_must_precede_older",
      "input": {
        "sessions": [
          {
            "sessionId": "session-old",
            "title": "older",
            "harness": "claude-code",
            "startMs": 1000,
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-new",
            "title": "newer",
            "harness": "codex",
            "startMs": 2000,
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {
        "errorContains": "violate canonical ordering"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "descending known start timestamp order"
      },
      "mutation": {
        "description": "places a newer timestamp after an older timestamp"
      }
    },
    {
      "family": "equal-timestamp-tiebreak",
      "name": "equal_timestamp_uses_session_identity_tiebreak",
      "input": {
        "sessions": [
          {
            "sessionId": "session-b",
            "title": "second",
            "harness": "codex",
            "startMs": 3000,
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-a",
            "title": "first",
            "harness": "claude-code",
            "startMs": 3000,
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {
        "errorContains": "violate canonical ordering"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "stable equal-timestamp tie break"
      },
      "mutation": {
        "description": "reverses session identity order for equal timestamps"
      }
    },
    {
      "family": "missing-timestamp-order",
      "name": "missing_timestamp_follows_known_timestamp",
      "input": {
        "sessions": [
          {
            "sessionId": "session-unknown",
            "title": "unknown",
            "harness": "opencode",
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-known",
            "title": "known",
            "harness": "claude-code",
            "startMs": 3000,
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {
        "errorContains": "violate canonical ordering"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "missing timestamps sort after known timestamps"
      },
      "mutation": {
        "description": "places a missing timestamp before a known timestamp"
      }
    },
    {
      "family": "compatibility-boolean",
      "name": "boolean_mirror_mismatch",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": false,
            "sessionIds": [
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-boolean-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "hasSession must mirror"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "compatibility boolean mirrors authoritative bindings"
      },
      "mutation": {
        "description": "clears hasSession while retaining a session binding"
      }
    },
    {
      "family": "complete-binding-truth",
      "name": "visible_reference_requires_complete_binding_truth",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": false
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-binding-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "hasCommitBinding=false"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "visible binding proves complete session binding truth"
      },
      "mutation": {
        "description": "marks a visibly bound session as having no authoritative binding"
      }
    },
    {
      "family": "known-session-reference",
      "name": "unknown_session_reference",
      "input": {
        "sessions": [],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-missing"
            ],
            "associations": [
              {
                "id": "assoc-missing",
                "sessionId": "session-missing",
                "conclusion": "candidate",
                "confidence": "medium",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "references unknown sessionId"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "commit bindings reference normalized session identities"
      },
      "mutation": {
        "description": "points a commit binding at a session absent from the session catalog"
      }
    },
    {
      "family": "unique-session-identity",
      "name": "duplicate_session_identity",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "first",
            "harness": "claude-code",
            "hasCommitBinding": false
          },
          {
            "sessionId": "session-a",
            "title": "duplicate",
            "harness": "claude-code",
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {
        "errorContains": "duplicate timeline session"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "normalized session identity uniqueness"
      },
      "mutation": {
        "description": "repeats one session identity in the session catalog"
      }
    },
    {
      "family": "unique-commit-binding",
      "name": "duplicate_binding",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a",
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-duplicate-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              },
              {
                "id": "assoc-duplicate-b",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "repeats sessionId"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "commit binding set uniqueness"
      },
      "mutation": {
        "description": "repeats one session identity in a commit binding list"
      }
    },
    {
      "family": "non-null-commit-bindings",
      "name": "null_commit_bindings",
      "input": {
        "sessions": [],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": false,
            "associations": [],
            "sessionIds": null
          }
        ]
      },
      "expected": {
        "errorContains": "null sessionIds"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "non-null commit binding arrays"
      },
      "mutation": {
        "description": "omits the required sessionIds array from a commit"
      }
    },
    {
      "family": "known-harness",
      "name": "unknown_timeline_harness",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "unknown-harness",
            "hasCommitBinding": false
          }
        ],
        "commits": []
      },
      "expected": {
        "errorContains": "unknown harness"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "timeline sessions use the canonical harness closed set"
      },
      "mutation": {
        "description": "replaces a known harness with an unknown harness token"
      }
    },
    {
      "family": "canonical-binding-order",
      "name": "noncanonical_binding_order",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "first",
            "harness": "claude-code",
            "startMs": 2000,
            "hasCommitBinding": true
          },
          {
            "sessionId": "session-b",
            "title": "second",
            "harness": "codex",
            "startMs": 1000,
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-b",
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-order-b",
                "sessionId": "session-b",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              },
              {
                "id": "assoc-order-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "noncanonical sessionIds order"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "commit bindings follow normalized timeline session rank"
      },
      "mutation": {
        "description": "reverses two valid bindings relative to the canonical session order"
      }
    },
    {
      "family": "null-associations",
      "name": "null_associations",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a"
            ],
            "associations": null
          }
        ]
      },
      "expected": {
        "errorContains": "null associations"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "the non-nil Associations invariant"
      },
      "mutation": {
        "description": "omits the required associations array from a commit that has sessionIds"
      }
    },
    {
      "family": "association-count-mismatch",
      "name": "association_count_does_not_mirror_session_ids",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a"
            ],
            "associations": []
          }
        ]
      },
      "expected": {
        "errorContains": "Associations must mirror SessionIDs"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "the Associations-mirrors-SessionIDs invariant"
      },
      "mutation": {
        "description": "leaves associations empty while sessionIds carries one binding"
      }
    },
    {
      "family": "association-rank-mismatch",
      "name": "association_rank_order_does_not_match_session_ids",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "first",
            "harness": "claude-code",
            "startMs": 2000,
            "hasCommitBinding": true
          },
          {
            "sessionId": "session-b",
            "title": "second",
            "harness": "codex",
            "startMs": 1000,
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a",
              "session-b"
            ],
            "associations": [
              {
                "id": "assoc-rank-b",
                "sessionId": "session-b",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              },
              {
                "id": "assoc-rank-a",
                "sessionId": "session-a",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "Associations must equal SessionIDs in the same rank order"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "requirement",
        "ref": "the Associations rank-order invariant"
      },
      "mutation": {
        "description": "swaps the two associations relative to the canonical sessionIds order"
      }
    },
    {
      "family": "association-invalid-conclusion",
      "name": "association_conclusion_not_in_closed_set",
      "input": {
        "sessions": [
          {
            "sessionId": "session-a",
            "title": "work",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "commit-a",
            "subject": "work",
            "hasSession": true,
            "sessionIds": [
              "session-a"
            ],
            "associations": [
              {
                "id": "assoc-invalid-conclusion",
                "sessionId": "session-a",
                "conclusion": "unknown-conclusion",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "commit-a"
                  }
                ]
              }
            ]
          }
        ]
      },
      "expected": {
        "errorContains": "association conclusion validation failed"
      },
      "classification": "must-fail",
      "provenance": {
        "source": "enum",
        "ref": "schema.AllAssociationConclusions"
      },
      "mutation": {
        "description": "replaces a known association conclusion with a value outside the closed set"
      }
    }
  ],
  "successorAssociationMirrorCases": [
    {
      "family": "rewrite-successor-association-id-mirror",
      "name": "rewrite_successor_association_id_drift_is_rejected",
      "input": {
        "sessions": [
          {
            "sessionId": "session-mirror-id",
            "title": "mirrored identifier",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-mirror-id",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-mirror-id"
            ],
            "associations": [
              {
                "id": "assoc-mirror-id-successor",
                "sessionId": "session-mirror-id",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-id"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-mirror-id",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-mirror-id"
            ],
            "associations": [
              {
                "id": "assoc-mirror-id-ledger",
                "sessionId": "session-mirror-id",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-id"
                  }
                ]
              }
            ],
            "successorHash": "successor-mirror-id",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "does not exactly mirror the rewrite-ledger association",
        "repair": {
          "kind": "replace_successor_association",
          "ghostHash": "ghost-mirror-id",
          "successorHash": "successor-mirror-id",
          "associationId": "assoc-mirror-id-ledger",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "displayed rewrite successors preserve durable association IDs"
      },
      "mutation": {
        "description": "changes only the successor association ID while retaining the same session binding"
      }
    },
    {
      "family": "rewrite-successor-association-session-mirror",
      "name": "rewrite_successor_association_session_drift_is_rejected",
      "input": {
        "sessions": [
          {
            "sessionId": "session-mirror-session-ledger",
            "title": "ledger session",
            "harness": "claude-code",
            "hasCommitBinding": true
          },
          {
            "sessionId": "session-mirror-session-successor",
            "title": "successor session",
            "harness": "codex",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-mirror-session",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-mirror-session-successor"
            ],
            "associations": [
              {
                "id": "assoc-mirror-session",
                "sessionId": "session-mirror-session-successor",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-session"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-mirror-session",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-mirror-session-ledger"
            ],
            "associations": [
              {
                "id": "assoc-mirror-session",
                "sessionId": "session-mirror-session-ledger",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-session"
                  }
                ]
              }
            ],
            "successorHash": "successor-mirror-session",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "does not exactly mirror the rewrite-ledger association",
        "repair": {
          "kind": "replace_successor_association",
          "ghostHash": "ghost-mirror-session",
          "successorHash": "successor-mirror-session",
          "associationId": "assoc-mirror-session",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "displayed rewrite successors preserve the ledger session binding for a durable association ID"
      },
      "mutation": {
        "description": "changes only the successor binding session while retaining the durable association ID"
      }
    },
    {
      "family": "rewrite-successor-association-conclusion-mirror",
      "name": "rewrite_successor_association_conclusion_drift_is_rejected",
      "input": {
        "sessions": [
          {
            "sessionId": "session-mirror-conclusion",
            "title": "mirrored conclusion",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-mirror-conclusion",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-mirror-conclusion"
            ],
            "associations": [
              {
                "id": "assoc-mirror-conclusion",
                "sessionId": "session-mirror-conclusion",
                "conclusion": "candidate",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-conclusion"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-mirror-conclusion",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-mirror-conclusion"
            ],
            "associations": [
              {
                "id": "assoc-mirror-conclusion",
                "sessionId": "session-mirror-conclusion",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-conclusion"
                  }
                ]
              }
            ],
            "successorHash": "successor-mirror-conclusion",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "does not exactly mirror the rewrite-ledger association",
        "repair": {
          "kind": "replace_successor_association",
          "ghostHash": "ghost-mirror-conclusion",
          "successorHash": "successor-mirror-conclusion",
          "associationId": "assoc-mirror-conclusion",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "displayed rewrite successors preserve the ledger conclusion"
      },
      "mutation": {
        "description": "changes only the successor association conclusion"
      }
    },
    {
      "family": "rewrite-successor-association-confidence-mirror",
      "name": "rewrite_successor_association_confidence_drift_is_rejected",
      "input": {
        "sessions": [
          {
            "sessionId": "session-mirror-confidence",
            "title": "mirrored confidence",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-mirror-confidence",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-mirror-confidence"
            ],
            "associations": [
              {
                "id": "assoc-mirror-confidence",
                "sessionId": "session-mirror-confidence",
                "conclusion": "confirmed",
                "confidence": "low",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-confidence"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-mirror-confidence",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-mirror-confidence"
            ],
            "associations": [
              {
                "id": "assoc-mirror-confidence",
                "sessionId": "session-mirror-confidence",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-confidence"
                  }
                ]
              }
            ],
            "successorHash": "successor-mirror-confidence",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "does not exactly mirror the rewrite-ledger association",
        "repair": {
          "kind": "replace_successor_association",
          "ghostHash": "ghost-mirror-confidence",
          "successorHash": "successor-mirror-confidence",
          "associationId": "assoc-mirror-confidence",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "displayed rewrite successors preserve the ledger confidence"
      },
      "mutation": {
        "description": "changes only the successor association confidence"
      }
    },
    {
      "family": "rewrite-successor-association-evidence-mirror",
      "name": "rewrite_successor_association_evidence_drift_is_rejected",
      "input": {
        "sessions": [
          {
            "sessionId": "session-mirror-evidence",
            "title": "mirrored evidence",
            "harness": "claude-code",
            "hasCommitBinding": true
          }
        ],
        "commits": [
          {
            "hash": "successor-mirror-evidence",
            "subject": "squash successor",
            "hasSession": true,
            "sessionIds": [
              "session-mirror-evidence"
            ],
            "associations": [
              {
                "id": "assoc-mirror-evidence",
                "sessionId": "session-mirror-evidence",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "successor-mirror-evidence"
                  },
                  {
                    "kind": "touched_file",
                    "touchedFilePath": "internal/api/review.go"
                  }
                ]
              }
            ]
          }
        ],
        "rewrittenCommits": [
          {
            "ghostHash": "ghost-mirror-evidence",
            "subject": "original recorded commit",
            "sessionIds": [
              "session-mirror-evidence"
            ],
            "associations": [
              {
                "id": "assoc-mirror-evidence",
                "sessionId": "session-mirror-evidence",
                "conclusion": "confirmed",
                "confidence": "high",
                "evidence": [
                  {
                    "kind": "recorded_commit",
                    "recordedCommitHash": "ghost-mirror-evidence"
                  },
                  {
                    "kind": "touched_file",
                    "touchedFilePath": "internal/api/review.go"
                  }
                ]
              }
            ],
            "successorHash": "successor-mirror-evidence",
            "resolution": "rewritten",
            "method": "patch_id",
            "confidence": "high"
          }
        ]
      },
      "expected": {
        "errorContains": "does not exactly mirror the rewrite-ledger association",
        "repair": {
          "kind": "replace_successor_association",
          "ghostHash": "ghost-mirror-evidence",
          "successorHash": "successor-mirror-evidence",
          "associationId": "assoc-mirror-evidence",
          "postMutationValid": true
        }
      },
      "classification": "must-fail",
      "provenance": {
        "source": "boundary",
        "ref": "displayed rewrite successors preserve ordered atomic evidence"
      },
      "mutation": {
        "description": "changes only the recorded commit detail in otherwise canonical ordered successor evidence"
      }
    }
  ]
};
