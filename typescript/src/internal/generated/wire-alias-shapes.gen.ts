// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
export const wireAliasShapes = {
  "ActorOrigin": {},
  "ChildSessionRef": {
    "fields": {
      "id": {},
      "project": {},
      "startTime": {}
    }
  },
  "CommandInvocation": {
    "fields": {
      "args": {},
      "name": {}
    }
  },
  "ContentKind": {},
  "ContentOrigin": {},
  "ContentOwnership": {},
  "ContentProvenance": {
    "fields": {
      "actor": "ActorOrigin",
      "delivery": "DeliveryOrigin",
      "evidence": "EvidenceKind",
      "inputModality": "InputModality",
      "origin": "ContentOrigin",
      "ownership": "ContentOwnership",
      "submissionRef": "SubmissionRef"
    }
  },
  "DeliveryOrigin": {},
  "EarlierHistorySection": {
    "fields": {
      "nativeMetadata": {
        "items": "NativeMetadataRecord"
      },
      "state": "EarlierHistoryState",
      "turns": {
        "items": "TurnDetail"
      }
    }
  },
  "EarlierHistoryState": {},
  "EntryType": {},
  "EvidenceKind": {},
  "Harness": {},
  "InputModality": {},
  "InterpretationDiagnostics": {
    "fields": {
      "partial": {}
    }
  },
  "NativeAttachmentRef": {
    "fields": {
      "toolCallId": {},
      "turnIndex": {}
    }
  },
  "NativeMetadataKind": {},
  "NativeMetadataRecord": {
    "fields": {
      "attachment": "NativeAttachmentRef",
      "customType": {},
      "data": {},
      "id": {},
      "kind": "NativeMetadataKind",
      "source": "NativeSourceRef"
    }
  },
  "NativeMetadataSourceType": {},
  "NativePiMessageRole": {},
  "NativeSourceRef": {
    "fields": {
      "entryRef": "SourceEntryRef",
      "messageRole": "NativePiMessageRole",
      "sourceType": "NativeMetadataSourceType"
    }
  },
  "ObservedModelID": {},
  "PublicRevisionRef": {},
  "PublicSourceAnchor": {
    "fields": {
      "kind": "PublicSourceAnchorKind",
      "sourceEntryRef": "SourceEntryRef",
      "sourceRevisionRef": "PublicRevisionRef"
    }
  },
  "PublicSourceAnchorKind": {},
  "RecordedCostDetail": {
    "fields": {
      "cacheRead": {},
      "cacheWrite": {},
      "input": {},
      "output": {},
      "source": {},
      "total": {}
    }
  },
  "RelationshipTargetState": {},
  "RetainedUnknownRecord": {
    "fields": {
      "kind": {},
      "namespace": {},
      "payload": {},
      "pointer": {},
      "position": {},
      "recordIndex": {},
      "sourceRef": {}
    }
  },
  "Role": {},
  "SessionDetailPayload": {
    "fields": {
      "childSessions": {
        "items": "ChildSessionRef"
      },
      "diagnostics": "InterpretationDiagnostics",
      "durationMins": {},
      "earlierHistory": {
        "items": "EarlierHistorySection"
      },
      "endTime": {},
      "gitBranch": {},
      "gitRemote": {},
      "harness": "Harness",
      "id": {},
      "inputSubmissionCount": {},
      "model": {},
      "nativeMetadata": {
        "items": "NativeMetadataRecord"
      },
      "outcome": "SessionOutcome",
      "parentSessionId": "SessionID",
      "project": {},
      "purpose": "SessionPurpose",
      "relationships": {
        "items": "SessionRelationship"
      },
      "retainedUnknown": {
        "items": "RetainedUnknownRecord"
      },
      "rootSessionId": "SessionID",
      "schemaVersion": {},
      "scorecard": "SessionScorecard",
      "sessionOrigin": "SessionOrigin",
      "source": {},
      "startTime": {},
      "status": {},
      "tokensIn": {},
      "tokensOut": {},
      "toolCallCount": {},
      "totalTokens": {},
      "turnCount": {},
      "turns": {
        "items": "TurnDetail"
      },
      "workingDirectory": {}
    }
  },
  "SessionID": {},
  "SessionOrigin": {},
  "SessionOutcome": {},
  "SessionPurpose": {},
  "SessionRelationship": {
    "fields": {
      "anchor": "PublicSourceAnchor",
      "evidence": "EvidenceKind",
      "kind": "SessionRelationshipKind",
      "targetLocalId": "SessionID",
      "targetState": "RelationshipTargetState"
    }
  },
  "SessionRelationshipKind": {},
  "SessionScorecard": {
    "fields": {
      "costTotalUsd": {},
      "m2TokenOutcomeRatio": {},
      "m4ConsecutiveErrorMax": {},
      "m5ContextUtilizationPct": {},
      "m6OutputSurvivalPct": {},
      "m7SpecHasConstraints": {},
      "m7SpecHasExamples": {},
      "outcome": "SessionOutcome",
      "retryTokensWasted": {},
      "signalDensity": {},
      "specQualityScore": {},
      "totalTokens": {},
      "withinSessionReverts": {}
    }
  },
  "SourceEntryRef": {},
  "StopReason": {},
  "SubmissionRef": {},
  "TokenUsageDetail": {
    "fields": {
      "cacheRead": {},
      "cacheWrite": {},
      "cacheWrite1h": {},
      "input": {},
      "output": {},
      "reasoning": {},
      "totalTokens": {}
    }
  },
  "ToolCallDetail": {
    "fields": {
      "arguments": {},
      "callEntryRef": "SourceEntryRef",
      "callProvenance": "ContentProvenance",
      "durationMs": {},
      "exitCode": {},
      "filePath": {},
      "id": {},
      "isError": {},
      "name": {},
      "namespace": {},
      "result": {},
      "resultEntryRef": "SourceEntryRef",
      "resultProvenance": "ContentProvenance",
      "toolKind": "ToolCallKind",
      "usage": "UsageDetail"
    }
  },
  "ToolCallKind": {},
  "TranscriptContent": {
    "fields": {
      "contractVersion": {},
      "kind": "ContentKind",
      "sessionDetail": "SessionDetailPayload"
    }
  },
  "TurnDetail": {
    "fields": {
      "agentName": {},
      "command": "CommandInvocation",
      "content": {},
      "depth": {},
      "entryType": "EntryType",
      "hasThinking": {},
      "index": {},
      "observedModel": "ObservedModelID",
      "parentIndex": {},
      "provenance": "ContentProvenance",
      "role": "Role",
      "sourceEntryRef": "SourceEntryRef",
      "stopReason": "StopReason",
      "timestamp": {},
      "tokensIn": {},
      "tokensOut": {},
      "toolCalls": {
        "items": "ToolCallDetail"
      },
      "usage": "UsageDetail"
    }
  },
  "UsageCompleteness": {},
  "UsageDetail": {
    "fields": {
      "completeness": "UsageCompleteness",
      "cost": "RecordedCostDetail",
      "ownerId": {},
      "scope": "UsageScope",
      "sourceEntryRef": "SourceEntryRef",
      "tokens": "TokenUsageDetail"
    }
  },
  "UsageScope": {}
} as const;
