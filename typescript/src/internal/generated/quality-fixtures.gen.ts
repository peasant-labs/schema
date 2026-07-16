// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
import type { QualityFixtures } from "../../fixtures/quality.js";

export const canonicalQualityFixtures: QualityFixtures = {
  "sessions": [
    {
      "name": "resolved_typical",
      "id": "sess-000",
      "date": "2025-11-01T00:00:00Z",
      "project": "fortuna",
      "scope": "Personal",
      "title": "Fix authentication middleware",
      "totalTokens": 15000,
      "inputTokens": 13500,
      "outputTokens": 1500,
      "turnCount": 10,
      "toolCalls": 15,
      "outcome": "resolved",
      "filesTouched": 3,
      "linesChanged": 150,
      "durationMinutes": 20,
      "retryLoops": 0,
      "retryTokensWasted": 0,
      "withinSessionReverts": 1,
      "signalDensity": 45.5,
      "specQualityScore": 75,
      "explorationRatio": 25,
      "scopeBreadth": 2,
      "discoveryTurns": 2
    },
    {
      "name": "resolved_high_tokens",
      "id": "sess-001",
      "date": "2025-11-02T00:00:00Z",
      "project": "peasant-api",
      "scope": "Platform Team",
      "title": "Add pagination to API",
      "totalTokens": 85000,
      "inputTokens": 80000,
      "outputTokens": 5000,
      "turnCount": 25,
      "toolCalls": 40,
      "outcome": "resolved",
      "filesTouched": 8,
      "linesChanged": 400,
      "durationMinutes": 50,
      "retryLoops": 1,
      "retryTokensWasted": 5000,
      "withinSessionReverts": 0,
      "signalDensity": 58.2,
      "specQualityScore": 85,
      "explorationRatio": 15,
      "scopeBreadth": 3,
      "discoveryTurns": 3
    },
    {
      "name": "partial_medium",
      "id": "sess-002",
      "date": "2025-11-03T00:00:00Z",
      "project": "data-pipeline",
      "scope": "Growth Team",
      "title": "Refactor database queries",
      "totalTokens": 120000,
      "inputTokens": 110000,
      "outputTokens": 10000,
      "turnCount": 35,
      "toolCalls": 60,
      "outcome": "partial",
      "filesTouched": 10,
      "linesChanged": 350,
      "durationMinutes": 70,
      "retryLoops": 2,
      "retryTokensWasted": 8000,
      "withinSessionReverts": 4,
      "signalDensity": 32.1,
      "specQualityScore": 55,
      "explorationRatio": 45,
      "scopeBreadth": 6,
      "discoveryTurns": 8
    },
    {
      "name": "failed_complex",
      "id": "sess-003",
      "date": "2025-11-04T00:00:00Z",
      "project": "auth-service",
      "scope": "Platform Team",
      "title": "Implement search filters",
      "totalTokens": 200000,
      "inputTokens": 180000,
      "outputTokens": 20000,
      "turnCount": 55,
      "toolCalls": 100,
      "outcome": "failed",
      "filesTouched": 15,
      "linesChanged": 800,
      "durationMinutes": 110,
      "retryLoops": 6,
      "retryTokensWasted": 25000,
      "withinSessionReverts": 5,
      "signalDensity": 18.7,
      "specQualityScore": 25,
      "explorationRatio": 75,
      "scopeBreadth": 8,
      "discoveryTurns": 10
    },
    {
      "name": "resolved_minimal",
      "id": "sess-004",
      "date": "2025-11-05T00:00:00Z",
      "project": "docs-site",
      "scope": "Personal",
      "title": "Update user settings page",
      "totalTokens": 5000,
      "inputTokens": 4500,
      "outputTokens": 500,
      "turnCount": 5,
      "toolCalls": 5,
      "outcome": "resolved",
      "filesTouched": 1,
      "linesChanged": 20,
      "durationMinutes": 10,
      "retryLoops": 0,
      "retryTokensWasted": 0,
      "withinSessionReverts": 0,
      "signalDensity": 62,
      "specQualityScore": 90,
      "explorationRatio": 10,
      "scopeBreadth": 1,
      "discoveryTurns": 1
    }
  ],
  "sets": [
    {
      "name": "project_mix",
      "cases": [
        "resolved_typical",
        "partial_medium"
      ]
    }
  ],
  "variations": {
    "outcomes": [
      {
        "value": "resolved"
      },
      {
        "value": "partial"
      },
      {
        "value": "failed"
      },
      {
        "value": ""
      }
    ],
    "projects": [
      {
        "value": "fortuna"
      },
      {
        "value": "peasant-api"
      },
      {
        "value": "data-pipeline"
      },
      {
        "value": "auth-service"
      },
      {
        "value": "docs-site"
      },
      {
        "value": ""
      },
      {
        "value": "invalid-project"
      }
    ],
    "scopes": [
      {
        "value": "Personal"
      },
      {
        "value": "Platform Team"
      },
      {
        "value": "Growth Team"
      },
      {
        "value": ""
      },
      {
        "value": "Unknown Scope"
      }
    ],
    "taskTitles": [
      {
        "value": "Fix authentication middleware"
      },
      {
        "value": "Add pagination to API"
      },
      {
        "value": "Refactor database queries"
      },
      {
        "value": "Implement search filters"
      },
      {
        "value": "Update user settings page"
      },
      {
        "value": "Fix CORS configuration"
      },
      {
        "value": "Add webhook handlers"
      },
      {
        "value": "Migrate to TypeScript"
      },
      {
        "value": "Optimize image loading"
      },
      {
        "value": "Add rate limiting"
      },
      {
        "value": ""
      }
    ],
    "tokenRatios": [
      {
        "name": "typical_input_heavy",
        "inputRatio": 0.95
      },
      {
        "name": "balanced",
        "inputRatio": 0.5
      },
      {
        "name": "output_heavy",
        "inputRatio": 0.2
      },
      {
        "name": "empty_input",
        "inputRatio": 0
      },
      {
        "name": "empty_output",
        "inputRatio": 1
      }
    ],
    "metrics": {
      "retryLoops": [
        {
          "name": "none",
          "value": 0
        },
        {
          "name": "few",
          "value": 2
        },
        {
          "name": "many",
          "value": 8
        },
        {
          "name": "negative",
          "value": -1
        }
      ],
      "signalDensity": [
        {
          "name": "high",
          "value": 60
        },
        {
          "name": "typical",
          "value": 40
        },
        {
          "name": "low",
          "value": 15
        },
        {
          "name": "zero",
          "value": 0
        },
        {
          "name": "over_100",
          "value": 120
        }
      ],
      "specQualityScore": [
        {
          "name": "perfect",
          "value": 100
        },
        {
          "name": "high",
          "value": 85
        },
        {
          "name": "medium",
          "value": 50
        },
        {
          "name": "low",
          "value": 20
        },
        {
          "name": "zero",
          "value": 0
        },
        {
          "name": "negative",
          "value": -5
        },
        {
          "name": "over_100",
          "value": 150
        }
      ],
      "filesTouched": [
        {
          "name": "none",
          "value": 0
        },
        {
          "name": "few",
          "value": 3
        },
        {
          "name": "typical",
          "value": 8
        },
        {
          "name": "many",
          "value": 15
        },
        {
          "name": "negative",
          "value": -1
        }
      ],
      "linesChanged": [
        {
          "name": "none",
          "value": 0
        },
        {
          "name": "few",
          "value": 20
        },
        {
          "name": "typical",
          "value": 200
        },
        {
          "name": "many",
          "value": 800
        },
        {
          "name": "negative",
          "value": -10
        }
      ]
    }
  }
};
