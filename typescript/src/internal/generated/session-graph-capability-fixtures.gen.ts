// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
export const canonicalSessionGraphCapabilityFixtures = {
  "requiredNames": {
    "derivation": [
      "legacy-empty",
      "model-only",
      "source-ref-only",
      "count-only-zero",
      "count-only-positive",
      "root-only",
      "purpose-only-interaction",
      "purpose-only-delegated",
      "purpose-only-helper",
      "purpose-only-unknown",
      "relationship-only",
      "main-deeper-provenance",
      "folded-call-provenance",
      "folded-result-provenance",
      "earlier-only",
      "earlier-nested-tool-native-usage-model",
      "navigation-only-read"
    ],
    "reader": [
      "advertisement-omitted",
      "advertisement-empty",
      "advertisement-null",
      "advertisement-unknown-duplicate-unordered",
      "advertisement-future-suffix"
    ],
    "producer": [
      "producer-canonical-four",
      "producer-unsorted-four",
      "producer-duplicate",
      "producer-unknown"
    ]
  },
  "derivation": {
    "cases": [
      {
        "name": "legacy-empty",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "omits optional evidence"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[]}"
        },
        "expected": {}
      },
      {
        "name": "model-only",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds model seed"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"model\":\"seed/model\",\"turns\":[]}"
        },
        "expected": {}
      },
      {
        "name": "source-ref-only",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds source ref"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":1,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[{\"index\":0,\"role\":\"user\",\"sourceEntryRef\":\"e1\",\"content\":\"\",\"depth\":0,\"timestamp\":\"2020-01-01T00:00:00Z\"}]}"
        },
        "expected": {}
      },
      {
        "name": "count-only-zero",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "measures zero submissions"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"inputSubmissionCount\":0,\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "count-only-positive",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "measures positive submissions"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"inputSubmissionCount\":1,\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "root-only",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds root"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"rootSessionId\":\"00000000-0000-4000-8000-000000000001\",\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "purpose-only-interaction",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds interaction purpose"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"purpose\":\"interaction\",\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "purpose-only-delegated",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds delegated purpose"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"purpose\":\"delegated_work\",\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "purpose-only-helper",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds helper purpose"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"purpose\":\"helper_review\",\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "purpose-only-unknown",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds unknown purpose"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"purpose\":\"unknown\",\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "relationship-only",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds relationship"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"relationships\":[{\"kind\":\"started_by\",\"targetState\":\"unknown\",\"evidence\":\"unknown\"}],\"turns\":[]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "main-deeper-provenance",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-11"
        },
        "mutation": {
          "description": "adds nonroot provenance"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":2,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"depth\":0,\"timestamp\":\"2020-01-01T00:00:00Z\"},{\"index\":1,\"role\":\"assistant\",\"depth\":1,\"parentIndex\":0,\"provenance\":{\"origin\":\"unknown\",\"actor\":\"unknown\",\"delivery\":\"unknown\",\"ownership\":\"uncertain\",\"evidence\":\"unknown\",\"inputModality\":\"unknown\"},\"content\":\"\",\"timestamp\":\"2020-01-01T00:00:00Z\"}]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "folded-call-provenance",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-11"
        },
        "mutation": {
          "description": "adds call provenance"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":1,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"toolCalls\":[{\"id\":\"t1\",\"callProvenance\":{\"origin\":\"unknown\",\"actor\":\"unknown\",\"delivery\":\"unknown\",\"ownership\":\"uncertain\",\"evidence\":\"unknown\",\"inputModality\":\"unknown\"},\"name\":\"\",\"arguments\":\"\",\"result\":\"\"}],\"content\":\"\",\"depth\":0,\"timestamp\":\"2020-01-01T00:00:00Z\"}]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "folded-result-provenance",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-11"
        },
        "mutation": {
          "description": "adds result provenance"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":1,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"toolCalls\":[{\"id\":\"t1\",\"resultProvenance\":{\"origin\":\"unknown\",\"actor\":\"unknown\",\"delivery\":\"unknown\",\"ownership\":\"uncertain\",\"evidence\":\"unknown\",\"inputModality\":\"unknown\"},\"name\":\"\",\"arguments\":\"\",\"result\":\"\"}],\"content\":\"\",\"depth\":0,\"timestamp\":\"2020-01-01T00:00:00Z\"}]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "earlier-only",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds empty retained section"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[]}]}"
        },
        "expected": {
          "capabilities": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "earlier-nested-tool-native-usage-model",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "combines valid Pi earlier evidence"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"pi\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[],\"earlierHistory\":[{\"state\":\"uncertain_migrated\",\"turns\":[{\"index\":0,\"role\":\"assistant\",\"content\":\"\",\"depth\":0,\"timestamp\":\"2020-01-01T00:00:00Z\"},{\"index\":1,\"role\":\"assistant\",\"depth\":1,\"parentIndex\":0,\"observedModel\":\"provider/model\",\"provenance\":{\"origin\":\"unknown\",\"actor\":\"unknown\",\"delivery\":\"unknown\",\"ownership\":\"uncertain\",\"evidence\":\"unknown\",\"inputModality\":\"unknown\"},\"sourceEntryRef\":\"e_assistant\",\"usage\":{\"ownerId\":\"u_assistant\",\"sourceEntryRef\":\"e_assistant\",\"scope\":\"assistant\",\"completeness\":\"unknown\"},\"toolCalls\":[{\"id\":\"t1\",\"resultEntryRef\":\"e_result\",\"usage\":{\"ownerId\":\"u_tool\",\"sourceEntryRef\":\"e_result\",\"scope\":\"tool\",\"completeness\":\"unknown\"},\"name\":\"\",\"arguments\":\"\",\"result\":\"\"}],\"content\":\"\",\"timestamp\":\"2020-01-01T00:00:00Z\"}],\"nativeMetadata\":[{\"id\":\"n1\",\"kind\":\"pi.custom.data\",\"source\":{\"entryRef\":\"e_native\",\"sourceType\":\"pi.custom\"},\"customType\":\"fixture\",\"data\":{}}]}]}"
        },
        "expected": {
          "capabilities": [
            "detailed_usage_v1",
            "native_metadata_v1",
            "observed_model_v1",
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "navigation-only-read",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "separates read metadata from durable detail"
        },
        "input": {
          "detailJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[]}",
          "readJSON": "{\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"id\":\"child\",\"harness\":\"claude-code\",\"turns\":[],\"relationshipNavigation\":[]}",
          "rejectedDurableJSON": "{\"id\":\"fixture-session\",\"startTime\":\"2020-01-01T00:00:00Z\",\"endTime\":\"2020-01-01T00:00:00Z\",\"durationMins\":0,\"totalTokens\":0,\"tokensIn\":0,\"tokensOut\":0,\"turnCount\":0,\"toolCallCount\":0,\"harness\":\"claude-code\",\"outcome\":\"resolved\",\"sessionOrigin\":\"unknown\",\"turns\":[],\"relationshipNavigation\":[]}"
        },
        "expected": {}
      }
    ]
  },
  "reader": {
    "cases": [
      {
        "name": "advertisement-omitted",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "omits list"
        },
        "input": {
          "json": "{}"
        },
        "expected": {
          "missing": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "advertisement-empty",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "supplies empty list"
        },
        "input": {
          "json": "{\"contentCapabilities\":[]}"
        },
        "expected": {
          "missing": [
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "advertisement-null",
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "supplies null list"
        },
        "input": {
          "json": "{\"contentCapabilities\":null}"
        },
        "expected": {
          "errorContains": "contentCapabilities is null"
        }
      },
      {
        "name": "advertisement-unknown-duplicate-unordered",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "combines tolerated reader forms"
        },
        "input": {
          "json": "{\"contentCapabilities\":[\"session_graph_provenance_v1\",\"future_v9\",\"observed_model_v1\",\"session_graph_provenance_v1\"]}"
        },
        "expected": {
          "known": [
            "observed_model_v1",
            "session_graph_provenance_v1"
          ]
        }
      },
      {
        "name": "advertisement-future-suffix",
        "classification": "must-pass",
        "provenance": {
          "source": "boundary",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "substitutes opaque v2 suffix"
        },
        "input": {
          "json": "{\"contentCapabilities\":[\"session_graph_provenance_v2\"]}"
        },
        "expected": {
          "missing": [
            "session_graph_provenance_v1"
          ]
        }
      }
    ]
  },
  "producer": {
    "cases": [
      {
        "name": "producer-canonical-four",
        "classification": "must-pass",
        "provenance": {
          "source": "requirement",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "emits sorted inventory"
        },
        "input": {
          "tokens": [
            "detailed_usage_v1",
            "native_metadata_v1",
            "observed_model_v1",
            "session_graph_provenance_v1"
          ]
        },
        "expected": {
          "accepted": true
        }
      },
      {
        "name": "producer-unsorted-four",
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "swaps sorted tokens"
        },
        "input": {
          "tokens": [
            "native_metadata_v1",
            "detailed_usage_v1",
            "observed_model_v1",
            "session_graph_provenance_v1"
          ]
        },
        "expected": {
          "errorContains": "canonical lexicographic order"
        }
      },
      {
        "name": "producer-duplicate",
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "duplicates graph token"
        },
        "input": {
          "tokens": [
            "session_graph_provenance_v1",
            "session_graph_provenance_v1"
          ]
        },
        "expected": {
          "errorContains": "duplicated"
        }
      },
      {
        "name": "producer-unknown",
        "classification": "must-fail",
        "provenance": {
          "source": "boundary",
          "ref": "section-12.4"
        },
        "mutation": {
          "description": "adds unknown token"
        },
        "input": {
          "tokens": [
            "future_v9"
          ]
        },
        "expected": {
          "errorContains": "unknown token"
        }
      }
    ]
  }
} as const;
