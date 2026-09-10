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
    "/api/v1/config/capabilities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Discover optional UI behavior enabled for this server process. */
        get: operations["getUICapabilities"];
        put?: never;
        post?: never;
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
    "/api/v1/session-groups/{groupId}/members": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List authorized saved helper sessions within the originating grouped query scope. */
        get: operations["listLocalHelperGroupMembers"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/session-summaries": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get session summaries for an explicit set of session identifiers. This operation resolves links, so it applies NEITHER origin scope NOR selection scope: a session hidden from every discovery list is still returned here when its identifier is named. An identifier that names no session on this machine is omitted from the response rather than failing the batch, so a partially stale link still resolves the sessions that do exist. */
        get: operations["getSessionSummariesByID"];
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
    "/api/v1/sessions/{id}/transcript": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the flat additive session transcript by ID. */
        get: operations["getSessionTranscript"];
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
    "/api/v1/sync/sessions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List sync candidates with optional owner-nested helper groups. */
        get: operations["listSyncSessionsGrouped"];
        put?: never;
        post?: never;
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
        LocalSyncSessionsPayload: Schema.LocalSyncSessionsPayload;
        QualityPayload: Schema.QualityPayload;
        SchemaActivityEdge: Schema.ActivityEdge;
        /**
         * Actor Origin
         * @description Closed session graph value
         * @example operator
         * @example agent_delegate
         * @example harness
         * @example unknown
         * @enum {string}
         */
        SchemaActorOrigin: Schema.ActorOrigin;
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
        SchemaCommandInvocation: Schema.CommandInvocation;
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
        /**
         * Content Origin
         * @description Closed session graph value
         * @example submitted_input
         * @example harness_context
         * @example agent_output
         * @example agent_communication
         * @example tool_activity
         * @example system_control
         * @example generated_summary
         * @example unknown
         * @enum {string}
         */
        SchemaContentOrigin: Schema.ContentOrigin;
        /**
         * Content Ownership
         * @description Closed session graph value
         * @example local
         * @example inherited
         * @example uncertain
         * @enum {string}
         */
        SchemaContentOwnership: Schema.ContentOwnership;
        SchemaContentProvenance: Schema.ContentProvenance;
        SchemaCreateAnnotationRequest: Schema.CreateAnnotationRequest;
        SchemaCreateAnnotationResponse: Schema.CreateAnnotationResponse;
        SchemaDayStats: Schema.DayStats;
        /**
         * Delivery Origin
         * @description Closed session graph value
         * @example session_admission
         * @example guardian_review
         * @example subagent_delivery
         * @example inherited_context
         * @example tool_delivery
         * @example system_lifecycle
         * @example unknown
         * @enum {string}
         */
        SchemaDeliveryOrigin: Schema.DeliveryOrigin;
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
        SchemaEarlierHistorySection: Schema.EarlierHistorySection;
        /**
         * Earlier History State
         * @description Closed session graph value
         * @example uncertain_migrated
         * @example uncertain_unresolved
         * @enum {string}
         */
        SchemaEarlierHistoryState: Schema.EarlierHistoryState;
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
        /**
         * Evidence Kind
         * @description Closed session graph value
         * @example native_typed
         * @example lifecycle_typed
         * @example existing_adapter
         * @example retained_last_good
         * @example unknown
         * @example conflict
         * @enum {string}
         */
        SchemaEvidenceKind: Schema.EvidenceKind;
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
        SchemaHelperContextSummary: Schema.HelperContextSummary;
        SchemaHelperGroupSummary: Schema.HelperGroupSummary;
        /**
         * Input Modality
         * @description Closed session graph value
         * @example none
         * @example text
         * @example media
         * @example user_action
         * @example mixed
         * @example unknown
         * @enum {string}
         */
        SchemaInputModality: Schema.InputModality;
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
        SchemaLocalHelperMembersPayload: Schema.LocalHelperMembersPayload;
        SchemaLocalSessionListItem: Schema.LocalSessionListItem;
        SchemaLocalSessionListPayload: Schema.LocalSessionListPayload;
        SchemaLocalSessionRow: Schema.LocalSessionRow;
        SchemaLocalSyncSummary: Schema.LocalSyncSummary;
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
        SchemaNativeAttachmentRef: Schema.NativeAttachmentRef;
        /**
         * Native Metadata Kind
         * @description Kind of bounded non-conversational native metadata
         * @example pi.custom.data
         * @example pi.custommessage.details
         * @example pi.toolresult.details
         * @example pi.compaction.details
         * @example pi.branchsummary.details
         * @enum {string}
         */
        SchemaNativeMetadataKind: Schema.NativeMetadataKind;
        SchemaNativeMetadataRecord: Schema.NativeMetadataRecord;
        /**
         * Native Metadata Source Type
         * @description Native source category for public metadata
         * @example pi.custom
         * @example pi.custom_message
         * @example pi.message
         * @example pi.compaction
         * @example pi.branch_summary
         * @enum {string}
         */
        SchemaNativeMetadataSourceType: Schema.NativeMetadataSourceType;
        /**
         * Native Pi Message Role
         * @description Pi message role needed for public metadata validation
         * @example toolResult
         * @enum {string}
         */
        SchemaNativePiMessageRole: Schema.NativePiMessageRole;
        SchemaNativeSourceRef: Schema.NativeSourceRef;
        /**
         * Observed Model ID
         * @description Exact UTF-8 model identifier observed on an assistant-generated turn; producer-enforced as assistant or subagent evidence. Values are non-empty and may not have a Unicode White_Space code point at either edge; all accepted bytes, including Unicode, mixed case, slashes, and internal spaces, are preserved.
         * @example anthropic/Claude-Opus-4-8
         * @example provider/Model Family
         */
        SchemaObservedModelID: Schema.ObservedModelID;
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
        /**
         * Public Revision Reference
         * Format: public-ref-utf8-96-bytes
         */
        SchemaPublicRevisionRef: Schema.PublicRevisionRef;
        SchemaPublicSourceAnchor: Schema.PublicSourceAnchor;
        /**
         * Public Source Anchor Kind
         * @description Closed session graph value
         * @example general_source_session
         * @example before_redacted_entry
         * @example through_redacted_entry
         * @enum {string}
         */
        SchemaPublicSourceAnchorKind: Schema.PublicSourceAnchorKind;
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
        SchemaRecordedCostDetail: Schema.RecordedCostDetail;
        /**
         * Relationship Navigation Status
         * @description Closed session graph value
         * @example resolved
         * @example general_link_only
         * @example known_unavailable
         * @example inaccessible
         * @example unknown
         * @example conflicting
         * @enum {string}
         */
        SchemaRelationshipNavigationStatus: Schema.RelationshipNavigationStatus;
        /**
         * Relationship Target State
         * @description Closed session graph value
         * @example target_known
         * @example target_known_retained
         * @example explicit_none
         * @example unknown
         * @example conflicting_current_native_evidence
         * @enum {string}
         */
        SchemaRelationshipTargetState: Schema.RelationshipTargetState;
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
        SchemaSearchResult: Schema.SearchResult;
        SchemaSessionAssociation: Schema.SessionAssociation;
        SchemaSessionDetailReadPayload: Schema.SessionDetailReadPayload;
        /**
         * Session ID
         * Format: session-id
         * @description Unique session identifier (UUID, agent-{hex}, ses_{id}, sess_{id} (ACP), msg_{id}, or a Strike session ID)
         * @example 99d59925-36bc-424c-a789-8be54d9702ba
         * @example agent-a3aee4f
         * @example ses_3cd91f52effeXd3QAJ54jOyzv5
         * @example sess_3cd91f52effeXd3QAJ54jOyzv5
         * @example 20260728T123456.123456789Z-ABCDEFGHIJKLMNOPQRST234567
         * @example ABCDEFGHIJKLMNOPQRST234567
         */
        SchemaSessionID: Schema.SessionID;
        SchemaSessionInsight: Schema.SessionInsight;
        /**
         * Session List Item Kind
         * @description Closed session graph value
         * @example transcript
         * @example context_container
         * @enum {string}
         */
        SchemaSessionListItemKind: Schema.SessionListItemKind;
        /**
         * Session Origin
         * @description Who drove a recorded session, as declared by the producer that recorded it
         * @example user
         * @example agent
         * @example unknown
         * @enum {string}
         */
        SchemaSessionOrigin: Schema.SessionOrigin;
        /**
         * Session Outcome
         * @description Resolution status of the session
         * @example resolved
         * @example partial
         * @example failed
         * @enum {string}
         */
        SchemaSessionOutcome: Schema.SessionOutcome;
        /**
         * Session Purpose
         * @description Closed session graph value
         * @example interaction
         * @example delegated_work
         * @example helper_review
         * @example unknown
         * @enum {string}
         */
        SchemaSessionPurpose: Schema.SessionPurpose;
        SchemaSessionRelationship: Schema.SessionRelationship;
        /**
         * Session Relationship Kind
         * @description Closed session graph value
         * @example started_by
         * @example context_from
         * @enum {string}
         */
        SchemaSessionRelationshipKind: Schema.SessionRelationshipKind;
        SchemaSessionRelationshipNavigation: Schema.SessionRelationshipNavigation;
        SchemaSessionScorecard: Schema.SessionScorecard;
        SchemaSessionSummary: Schema.SessionSummary;
        SchemaSessionsPayload: Schema.SessionsPayload;
        SchemaShutdownResponse: Schema.ShutdownResponse;
        /**
         * Source Entry Reference
         * Format: public-ref-utf8-96-bytes
         */
        SchemaSourceEntryRef: Schema.SourceEntryRef;
        /**
         * Stop Reason
         * @description Reason why a session or turn ended (ACP-aligned)
         * @example end_turn
         * @example max_tokens
         * @enum {string}
         */
        SchemaStopReason: Schema.StopReason;
        /**
         * Submission Reference
         * Format: public-ref-utf8-96-bytes
         */
        SchemaSubmissionRef: Schema.SubmissionRef;
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
        SchemaTokenUsageDetail: Schema.TokenUsageDetail;
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
        /**
         * Transcript ID
         * Format: uuid
         * @description Village-side transcript identifier (canonical lowercase-hex UUID)
         * @example 99d59925-36bc-424c-a789-8be54d9702ba
         */
        SchemaTranscriptID: Schema.TranscriptID;
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
        /**
         * Usage Completeness
         * @description Completeness of the five base token fields
         * @example complete
         * @example partial
         * @example unknown
         * @enum {string}
         */
        SchemaUsageCompleteness: Schema.UsageCompleteness;
        SchemaUsageDetail: Schema.UsageDetail;
        /**
         * Usage Scope
         * @description Native owner scope for detailed token and cost evidence
         * @example assistant
         * @example tool
         * @example summary
         * @enum {string}
         */
        SchemaUsageScope: Schema.UsageScope;
        SchemaValueDomain: Schema.ValueDomain;
        /**
         * Value Domain Kind
         * @description ISO 11179 value domain: enumerated (finite allowed set) or described (range/pattern constraint)
         * @example enumerated
         * @example described
         * @enum {string}
         */
        SchemaValueDomainKind: Schema.ValueDomainKind;
        SearchPayload: Schema.SearchPayload;
        ServerMessage: Schema.ServerMessage;
        SessionDetailPayload: Schema.SessionDetailPayload;
        SessionDetailReadPayload: Schema.SessionDetailReadPayload;
        TrendsPayload: Schema.TrendsPayload;
        UICapabilitiesResponse: Schema.UICapabilitiesResponse;
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
    getUICapabilities: {
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
                    "application/json": components["schemas"]["UICapabilitiesResponse"];
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
                /** @description Set to grouped to return owner-nested helper groups */
                view?: "grouped";
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
                    "application/json": components["schemas"]["SearchPayload"] | components["schemas"]["SchemaLocalSessionListPayload"];
                };
            };
        };
    };
    listLocalHelperGroupMembers: {
        parameters: {
            query: {
                /** @description Opaque member scope from the originating grouped response */
                scope: string;
                /** @description One-based member page */
                page?: number;
                /** @description Maximum members per page */
                limit?: number;
            };
            header?: never;
            path: {
                groupId: string;
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
                    "application/json": components["schemas"]["SchemaLocalHelperMembersPayload"];
                };
            };
        };
    };
    getSessionSummariesByID: {
        parameters: {
            query: {
                /** @description Comma-separated session IDs to resolve */
                ids: string;
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
                    "application/json": components["schemas"]["SchemaSessionsPayload"];
                };
            };
        };
    };
    listSessions: {
        parameters: {
            query?: {
                /** @description Set to grouped to return owner-nested helper groups */
                view?: "grouped";
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
                    "application/json": components["schemas"]["SchemaSessionsPayload"] | components["schemas"]["SchemaLocalSessionListPayload"];
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
                    "application/json": components["schemas"]["SchemaSessionDetailReadPayload"];
                };
            };
        };
    };
    getSessionTranscript: {
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
                    "application/json": components["schemas"]["SchemaSessionDetailReadPayload"];
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
    listSyncSessionsGrouped: {
        parameters: {
            query?: {
                /** @description Set to grouped for grouped rows */
                view?: "grouped";
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
                    "application/json": components["schemas"]["LocalSyncSessionsPayload"] | components["schemas"]["SchemaLocalSessionListPayload"];
                };
            };
        };
    };
}
