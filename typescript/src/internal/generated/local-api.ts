import type * as Schema from "../../index.js";
export interface paths {
    "/api/v1/annotation-types": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List all registered annotation types. */
        get: operations["listAnnotationTypes"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/annotations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List annotations for a session. */
        get: operations["listAnnotations"];
        put?: never;
        /** @description Create a new annotation. */
        post: operations["createAnnotation"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/config/mock": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get mock data configuration. */
        get: operations["getMockConfig"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/health": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Health check endpoint. */
        get: operations["getHealth"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/map/{projectHash}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the full map graph for a project (optionally at a commit). */
        get: operations["getMapGraph"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/map/{projectHash}/node": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the rail panel detail for one map node. */
        get: operations["getMapNodeDetail"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/map/{projectHash}/tasks": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List a project's tasks (reverse-chronological, cap 500). */
        get: operations["listProjectTasks"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/resolve": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Resolve one explicit project display identity without enumerating sibling projects. */
        get: operations["resolveProject"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/projects/summary": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List per-project summary rows for the home picker (sessions, recorded coverage, last work, open changes). */
        get: operations["listProjectSummaries"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/review/{projectHash}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List a project's changes (open branches, then merged). */
        get: operations["listReviewChanges"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/review/{projectHash}/change": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the Review detail payload for one branch. */
        get: operations["getChangeDetail"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/review/{projectHash}/diff": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the rendered per-file unified diff for one changed file of a branch. */
        get: operations["getChangeDiff"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/search": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Full-text search across recorded (redacted) message entries; matches the first ~2000 chars of each turn plus truncated tool input/output. */
        get: operations["searchMessages"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/sessions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List session summaries. */
        get: operations["listSessions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/sessions/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get session detail by ID. */
        get: operations["getSession"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/shutdown": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Gracefully shutdown the server (localhost only). */
        post: operations["postShutdown"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        AnnotationsPayload: Schema.AnnotationsPayload;
        /**
         * Harness
         * @description AI coding tool or development environment
         * @example claude-code
         * @example opencode
         * @enum {string}
         */
        BestiaryHarness: Schema.Harness;
        ClientMessage: Schema.ClientMessage;
        DashboardPayload: Schema.DashboardPayload;
        QualityPayload: Schema.QualityPayload;
        SchemaActivityEdge: Schema.ActivityEdge;
        /**
         * Annotation Axis
         * @description Subscription dimension for annotation channels
         * @example type
         * @example session
         * @example project
         * @enum {string}
         */
        SchemaAnnotationAxis: Schema.AnnotationAxis;
        /**
         * Annotation Datatype
         * @description Storage type for annotation values (maps to SQLite STRICT column type)
         * @example text
         * @example integer
         * @enum {string}
         */
        SchemaAnnotationDatatype: Schema.AnnotationDatatype;
        /**
         * Annotation Status
         * @description ISO 11179 lifecycle state of an annotation type
         * @example active
         * @example deprecated
         * @enum {string}
         */
        SchemaAnnotationStatus: Schema.AnnotationStatus;
        SchemaAnnotationSummary: Schema.AnnotationSummary;
        SchemaAnnotationTypeSummary: Schema.AnnotationTypeSummary;
        /**
         * Annotator Kind
         * @description Type of entity that produced an annotation: human, agent (AI model), or rule (automated classifier)
         * @example human
         * @example agent
         * @example rule
         * @enum {string}
         */
        SchemaAnnotatorKind: Schema.AnnotatorKind;
        /**
         * Association Conclusion
         * @description Producer-supplied conclusion for a session-to-commit association: confirmed or candidate
         * @example confirmed
         * @example candidate
         * @enum {string}
         */
        SchemaAssociationConclusion: Schema.AssociationConclusion;
        /**
         * Association Evidence Kind
         * @description Atomic observation supporting a session-to-commit association
         * @example recorded_commit
         * @example touched_file
         * @example branch_membership
         * @example time_window
         * @enum {string}
         */
        SchemaAssociationEvidenceKind: Schema.AssociationEvidenceKind;
        SchemaAssociationEvidenceObservation: Schema.AssociationEvidenceObservation;
        /**
         * Association ID
         * @description Opaque durable Peasant identifier for one session-to-commit association
         * @example assoc-20260726:session-a:commit-1
         */
        SchemaAssociationID: Schema.AssociationID;
        /**
         * Change Binding
         * @description Strength of the evidence connecting a recorded session to a code change
         * @example bound
         * @example candidate
         * @enum {string}
         */
        SchemaChangeBinding: Schema.ChangeBinding;
        SchemaChangeDetailPayload: Schema.ChangeDetailPayload;
        SchemaChangeDiffPayload: Schema.ChangeDiffPayload;
        SchemaChangeSession: Schema.ChangeSession;
        SchemaChangeSummary: Schema.ChangeSummary;
        SchemaChannelSubscription: Schema.ChannelSubscription;
        /**
         * Channel Topic
         * @description Subscribable WebSocket data stream
         * @example dashboard
         * @example sessions
         * @example session_detail
         * @example trends
         * @example quality
         * @example annotations
         * @example project_familiarity
         * @enum {string}
         */
        SchemaChannelTopic: Schema.ChannelTopic;
        SchemaChildSessionRef: Schema.ChildSessionRef;
        SchemaCommitRef: Schema.CommitRef;
        /**
         * Confidence
         * @description Strength of evidence behind a derived relationship: high, medium, or low
         * @example high
         * @example medium
         * @example low
         * @enum {string}
         */
        SchemaConfidence: Schema.Confidence;
        SchemaCreateAnnotationRequest: Schema.CreateAnnotationRequest;
        SchemaCreateAnnotationResponse: Schema.CreateAnnotationResponse;
        SchemaDayStats: Schema.DayStats;
        SchemaDiffHunk: Schema.DiffHunk;
        SchemaDiffLine: Schema.DiffLine;
        /**
         * Diff Line Kind
         * @description Unified-diff line kind: context, addition, or deletion
         * @example context
         * @example add
         * @example del
         * @enum {string}
         */
        SchemaDiffLineKind: Schema.DiffLineKind;
        SchemaEdgeViolation: Schema.EdgeViolation;
        /**
         * Edge Violation Kind
         * @description Structural violation detected on a map edge
         * @example cycle
         * @example wrong_way
         * @enum {string}
         */
        SchemaEdgeViolationKind: Schema.EdgeViolationKind;
        /**
         * Entry Type
         * @description Classification of a single entry within an agent session transcript
         * @example text
         * @example tool_use
         * @example tool_result
         * @enum {string}
         */
        SchemaEntryType: Schema.EntryType;
        SchemaFileChange: Schema.FileChange;
        /**
         * File Change Status
         * @description Git file delta status: modified, added, deleted, or renamed
         * @example M
         * @example A
         * @example D
         * @example R
         * @enum {string}
         */
        SchemaFileChangeStatus: Schema.FileChangeStatus;
        SchemaFrictionCluster: Schema.FrictionCluster;
        SchemaHealthResponse: Schema.HealthResponse;
        SchemaInsightClassification: Schema.InsightClassification;
        SchemaInsightEvidence: Schema.InsightEvidence;
        /**
         * Insight Kind
         * @description What a SessionInsight observed: a decision, friction, an unusual rate elevation, or a retry loop
         * @example decision
         * @example friction
         * @example unusual
         * @example retry_loop
         * @enum {string}
         */
        SchemaInsightKind: Schema.InsightKind;
        /**
         * Insight Provenance
         * @description How a SessionInsight was produced: mechanical (rule-derived) or mined
         * @example mechanical
         * @example mined
         * @enum {string}
         */
        SchemaInsightProvenance: Schema.InsightProvenance;
        SchemaMapEdge: Schema.MapEdge;
        SchemaMapGraphPayload: Schema.MapGraphPayload;
        SchemaMapNode: Schema.MapNode;
        SchemaMapNodeDetailPayload: Schema.MapNodeDetailPayload;
        /**
         * Map Node Kind
         * @description Path-derived map node classification
         * @example module
         * @example package
         * @example file
         * @enum {string}
         */
        SchemaMapNodeKind: Schema.MapNodeKind;
        SchemaMapSlice: Schema.MapSlice;
        /**
         * Message Type
         * @description WebSocket message discriminator
         * @example subscribe
         * @example unsubscribe
         * @example dashboard
         * @example sessions
         * @example session_detail
         * @example trends
         * @example quality
         * @example annotations
         * @example project_familiarity
         * @example connected
         * @example error
         * @enum {string}
         */
        SchemaMessageType: Schema.MessageType;
        SchemaMockConfigResponse: Schema.MockConfigResponse;
        /**
         * Project Hash
         * @description SHA-256 hex digest of the project's origin URL or local path
         * @example a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
         */
        SchemaProjectHash: Schema.ProjectHash;
        SchemaProjectResolutionPayload: Schema.ProjectResolutionPayload;
        SchemaProjectSummariesPayload: Schema.ProjectSummariesPayload;
        SchemaProjectSummary: Schema.ProjectSummary;
        SchemaProjectTasksPayload: Schema.ProjectTasksPayload;
        SchemaProvenance: Schema.Provenance;
        SchemaQualitySession: Schema.QualitySession;
        /**
         * Read Attribution State
         * @description Whether per-file read attribution is recoverable for a node's editing sessions: complete, partial, or unavailable
         * @example complete
         * @example partial
         * @example unavailable
         * @enum {string}
         */
        SchemaReadAttributionState: Schema.ReadAttributionState;
        /**
         * Read State Grade
         * @description Ordinal explicit read-state act: none, viewed, reviewed, or reviewed_in_detail
         * @example none
         * @example viewed
         * @example reviewed
         * @example reviewed_in_detail
         * @enum {string}
         */
        SchemaReadStateGrade: Schema.ReadStateGrade;
        SchemaReviewListPayload: Schema.ReviewListPayload;
        /**
         * Rewrite Method
         * @description Mechanism the resolver used to map a ghost commit to its successor
         * @example hash
         * @example patch_id
         * @example author_identity
         * @example message_embedded
         * @example temporal
         * @example none
         * @enum {string}
         */
        SchemaRewriteMethod: Schema.RewriteMethod;
        /**
         * Rewrite Resolution
         * @description Whether a ledger-observed commit hash is live, was rewritten, or could not be resolved
         * @example live
         * @example rewritten
         * @example unresolved
         * @enum {string}
         */
        SchemaRewriteResolution: Schema.RewriteResolution;
        SchemaRewrittenCommit: Schema.RewrittenCommit;
        /**
         * Role
         * @description Sender role of a message turn
         * @example user
         * @example assistant
         * @enum {string}
         */
        SchemaRole: Schema.Role;
        /**
         * Scale Kind
         * @description Stevens measurement level: nominal (categories without order), ordinal (ordered categories), continuous (numeric range)
         * @example nominal
         * @example ordinal
         * @example continuous
         * @enum {string}
         */
        SchemaScaleKind: Schema.ScaleKind;
        SchemaSearchPayload: Schema.SearchPayload;
        SchemaSearchResult: Schema.SearchResult;
        SchemaSessionAssociation: Schema.SessionAssociation;
        SchemaSessionDetailPayload: Schema.SessionDetailPayload;
        /**
         * Session ID
         * Format: session-id
         * @description Unique session identifier (UUID, agent-{hex}, ses_{id}, sess_{id} (ACP), or msg_{id})
         * @example 99d59925-36bc-424c-a789-8be54d9702ba
         * @example agent-a3aee4f
         * @example ses_3cd91f52effeXd3QAJ54jOyzv5
         * @example sess_3cd91f52effeXd3QAJ54jOyzv5
         */
        SchemaSessionID: Schema.SessionID;
        SchemaSessionInsight: Schema.SessionInsight;
        /**
         * Session Outcome
         * @description Resolution status of the session
         * @example resolved
         * @example partial
         * @example failed
         * @enum {string}
         */
        SchemaSessionOutcome: Schema.SessionOutcome;
        SchemaSessionScorecard: Schema.SessionScorecard;
        SchemaSessionSummary: Schema.SessionSummary;
        SchemaSessionsPayload: Schema.SessionsPayload;
        SchemaShutdownResponse: Schema.ShutdownResponse;
        /**
         * Stop Reason
         * @description Reason why a session or turn ended (ACP-aligned)
         * @example end_turn
         * @example max_tokens
         * @enum {string}
         */
        SchemaStopReason: Schema.StopReason;
        /**
         * Target Kind
         * @description What is being annotated: session-level, entry-level (turn/tool call), meta-annotation, project-level, a specific file version (content-hash keyed read-state receipt), or a durable session-to-commit association
         * @example session
         * @example entry
         * @example file_version
         * @example association
         * @enum {string}
         */
        SchemaTargetKind: Schema.TargetKind;
        SchemaTaskSummary: Schema.TaskSummary;
        SchemaTimelineSessionRef: Schema.TimelineSessionRef;
        SchemaToolCallDetail: Schema.ToolCallDetail;
        /**
         * Tool Call Kind
         * @description Classification of a tool call, aligned with ACP ToolCallUpdate.kind
         * @example read
         * @example edit
         * @example execute
         * @enum {string}
         */
        SchemaToolCallKind: Schema.ToolCallKind;
        SchemaTurnDetail: Schema.TurnDetail;
        /**
         * Type Origin
         * @description Who created an annotation type: system (built-in), user (individual), or group (shared)
         * @example system
         * @example user
         * @enum {string}
         */
        SchemaTypeOrigin: Schema.TypeOrigin;
        SchemaUnusualSignal: Schema.UnusualSignal;
        SchemaValueDomain: Schema.ValueDomain;
        /**
         * Value Domain Kind
         * @description ISO 11179 value domain: enumerated (finite allowed set) or described (range/pattern constraint)
         * @example enumerated
         * @example described
         * @enum {string}
         */
        SchemaValueDomainKind: Schema.ValueDomainKind;
        ServerMessage: Schema.ServerMessage;
        SessionDetailPayload: Schema.SessionDetailPayload;
        TrendsPayload: Schema.TrendsPayload;
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    listAnnotationTypes: {
        parameters: {
            query?: {
                /** @description Filter by lifecycle status (active, deprecated, etc.) */
                status?: string;
                /** @description Filter by origin (system, user, group) */
                origin?: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": null | components["schemas"]["SchemaAnnotationTypeSummary"][];
                };
            };
        };
    };
    listAnnotations: {
        parameters: {
            query: {
                /** @description Session ID to filter annotations */
                session_id: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": null | components["schemas"]["SchemaAnnotationSummary"][];
                };
            };
        };
    };
    createAnnotation: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["SchemaCreateAnnotationRequest"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaCreateAnnotationResponse"];
                };
            };
        };
    };
    getMockConfig: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaMockConfigResponse"];
                };
            };
        };
    };
    getHealth: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaHealthResponse"];
                };
            };
        };
    };
    getMapGraph: {
        parameters: {
            query?: {
                /** @description Optional commit SHA to build the graph at (default HEAD) */
                commit?: string;
            };
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaMapGraphPayload"];
                };
            };
        };
    };
    getMapNodeDetail: {
        parameters: {
            query: {
                /** @description Repo-relative node ID */
                path: string;
            };
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaMapNodeDetailPayload"];
                };
            };
        };
    };
    listProjectTasks: {
        parameters: {
            query?: {
                /** @description Optional file or directory filter */
                file?: string;
            };
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaProjectTasksPayload"];
                };
            };
        };
    };
    resolveProject: {
        parameters: {
            query: {
                /** @description Exact project display identity from a saved route */
                name: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaProjectResolutionPayload"];
                };
            };
        };
    };
    listProjectSummaries: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaProjectSummariesPayload"];
                };
            };
        };
    };
    listReviewChanges: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaReviewListPayload"];
                };
            };
        };
    };
    getChangeDetail: {
        parameters: {
            query: {
                /** @description Branch name (may contain slashes) */
                branch: string;
            };
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaChangeDetailPayload"];
                };
            };
        };
    };
    getChangeDiff: {
        parameters: {
            query: {
                /** @description Branch name (may contain slashes) */
                branch: string;
                /** @description Repo-relative file path */
                file: string;
            };
            header?: never;
            path: {
                /** @description Opaque project hash */
                projectHash: components["schemas"]["SchemaProjectHash"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaChangeDiffPayload"];
                };
            };
        };
    };
    searchMessages: {
        parameters: {
            query: {
                /** @description Search query (min 2 chars; whitespace tokens ANDed) */
                q: string;
                /** @description Max results (default 20, capped at 50) */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaSearchPayload"];
                };
            };
        };
    };
    listSessions: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaSessionsPayload"];
                };
            };
        };
    };
    getSession: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaSessionDetailPayload"];
                };
            };
        };
    };
    postShutdown: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaShutdownResponse"];
                };
            };
        };
    };
}
