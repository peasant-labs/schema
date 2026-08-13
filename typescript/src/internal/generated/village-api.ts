import type * as Schema from "../../index.js";
export interface paths {
    "/api/v1/annotations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Push annotations for the authenticated owner. Every item selects exactly one target arm: targetKind association requires only targetAssociationId, while every other target kind rejects targetAssociationId. Entry targets require a non-empty sessionId; endIndex must be greater than entryIndex and is enforced by the request validation boundary. The server resolves association targets owner-scoped before writing the all-or-nothing batch. */
        post: operations["pushAnnotations"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/annotations/manifest": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Return the set of annotation content-hashes the village holds for the authenticated owner (the server-authoritative skip-gate manifest), plus a deterministic digest over the sorted hash set for a no-op short-circuit. */
        get: operations["getAnnotationManifest"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/cli/exchange": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Exchange an OAuth authorization code and state for API credentials. Called by the CLI's local callback server after receiving the browser redirect. */
        post: operations["cliExchangeCode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/cli/login": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Initiate CLI login via browser OAuth flow. The CLI opens this URL in the user's browser with a local callback port and CSRF state. After authentication, the village redirects the browser to the CLI's local callback server. */
        get: operations["cliLogin"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pull/transcripts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List transcripts the authenticated user may pull (own + group-shared; public excluded by the canPullTranscript policy). Offset pagination. */
        get: operations["listPullableTranscripts"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pull/transcripts/skip-gate": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Batch currency check for a pulling client: per transcript it holds, send the id, the content-hash it holds, and its own annotation-hash set; receive per pullable id {contentCurrent, annotationsCurrent}. Non-pullable ids are withheld by omission from results. */
        post: operations["pullSkipGate"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pull/transcripts/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get a pullable transcript's metadata (owner, visibility, server-computed content_hash, push contract version). 404 (not 403) when not pullable. */
        get: operations["getPullTranscript"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pull/transcripts/{id}/annotations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List a transcript's annotations as PullAnnotation rows (AnnotationSummary plus author identity), so pulled annotations can be foreign-marked and own-authored rows excluded. */
        get: operations["getPullTranscriptAnnotations"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pull/transcripts/{id}/content": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Stream the transcript blob as served. Sends ETag: "<content_hash>"; a matching If-None-Match request yields 304 Not Modified. */
        get: operations["getPullTranscriptContent"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/schema/version": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Advertise the annotation schema version plus the push and pull contract acceptance/negotiation windows ([Min, Current]) so the CLI can preflight and version-negotiate before publishing or pulling. */
        get: operations["getSchemaVersion"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts/publish": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Publish exact transcript bytes with typed JSON metadata. A server advertising observed_model version 1.0.0 MUST reject observedModel evidence on non-assistant turns before persistence; assistant includes nested subagent output represented by role assistant. Creation returns 201 and replacement returns 200; both carry the complete authoritative receipt. */
        post: operations["publishTranscript"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** @description Update an owned transcript's metadata and governance axes. Every field is optional: omission preserves stored state; empty title, description, or tags clear those fields; license null requests clear; and a canonical license replaces. Explicit null for title, description, tags, or visibility is rejected, as are unknown fields and invalid or duplicate tags. Clearing an already granted Creative Commons license remains subject to the server's irrevocability rule. Only the owner may call this operation. A successful update returns the complete typed authoritative editable state, including the canonical transcript URL and positive update timestamp. Visibility accepts private and public; organization-scoped visibility remains deferred. */
        patch: operations["updateTranscript"];
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        /**
         * Harness
         * @description AI coding tool or development environment
         * @example claude-code
         * @example opencode
         * @enum {string}
         */
        BestiaryHarness: Schema.Harness;
        FormDataOpenapiTranscriptPublishMultipartRequest: {
            /** @description PublishRequest JSON encoded with Content-Type application/json. */
            metadata: components["schemas"]["OpenapiAuthoritativeTranscriptPublishRequest"];
            /** @description Exact transcript bytes whose SHA3-256 digest equals metadata.contentHash. */
            transcript_file: components["schemas"]["MultipartFileHeader"];
        };
        /** Format: binary */
        MultipartFileHeader: string;
        OpenapiAuthoritativeTranscriptPublishRequest: {
            contentHash: components["schemas"]["SchemaTranscriptContentHash"];
            diagnostics?: components["schemas"]["SchemaAuthoritativeDiagnosticsInfo"];
            entries?: components["schemas"]["SchemaAuthoritativeSessionEntry"][];
            git?: components["schemas"]["SchemaAuthoritativeGitContext"];
            identity?: components["schemas"]["SchemaAuthoritativeSessionIdentity"];
            license?: components["schemas"]["SchemaLicense"];
            model: components["schemas"]["SchemaAuthoritativeModelInfo"];
            project?: components["schemas"]["SchemaAuthoritativeProjectContext"];
            quality?: components["schemas"]["SchemaAuthoritativeQualityMetrics"];
            source?: components["schemas"]["SchemaAuthoritativeSourceInfo"];
            stats?: components["schemas"]["SchemaAuthoritativeSessionStats"];
            subagents?: components["schemas"]["SchemaAuthoritativeSubagentRef"][];
            timestamp?: components["schemas"]["SchemaAuthoritativeTimestampInfo"];
            visibilityIntent?: components["schemas"]["SchemaVisibilityIntent"];
        };
        OpenapiTranscriptUpdateErrorResponse: {
            error: string;
        };
        SchemaAnnotationEntryTarget: Schema.AnnotationEntryTarget;
        SchemaAnnotationManifestResponse: Schema.AnnotationManifestResponse;
        SchemaAnnotationPushItem: Schema.AnnotationPushItem;
        SchemaAnnotationPushRequest: Schema.AnnotationPushRequest;
        SchemaAnnotationPushResponse: Schema.AnnotationPushResponse;
        SchemaAnnotationPushResult: Schema.AnnotationPushResult;
        /**
         * Annotation Push Status
         * @description Per-item annotation push outcome
         * @example created
         * @example updated
         * @example skipped
         * @example error
         * @enum {string}
         */
        SchemaAnnotationPushStatus: Schema.AnnotationPushStatus;
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
         * Association ID
         * @description Opaque durable Peasant identifier for one session-to-commit association
         * @example assoc-20260726:session-a:commit-1
         */
        SchemaAssociationID: Schema.AssociationID;
        SchemaAuthoritativeCommitInfo: Schema.AuthoritativeCommitInfo;
        SchemaAuthoritativeDiagnosticEntry: Schema.AuthoritativeDiagnosticEntry;
        SchemaAuthoritativeDiagnosticsInfo: Schema.AuthoritativeDiagnosticsInfo;
        SchemaAuthoritativeGitContext: Schema.AuthoritativeGitContext;
        SchemaAuthoritativeModelInfo: Schema.AuthoritativeModelInfo;
        SchemaAuthoritativeProjectContext: Schema.AuthoritativeProjectContext;
        SchemaAuthoritativePublishResponse: Schema.AuthoritativePublishResponse;
        SchemaAuthoritativeQualityMetrics: Schema.AuthoritativeQualityMetrics;
        SchemaAuthoritativeSessionEntry: Schema.AuthoritativeSessionEntry;
        SchemaAuthoritativeSessionIdentity: Schema.AuthoritativeSessionIdentity;
        SchemaAuthoritativeSessionStats: Schema.AuthoritativeSessionStats;
        SchemaAuthoritativeSourceInfo: Schema.AuthoritativeSourceInfo;
        SchemaAuthoritativeSubagentRef: Schema.AuthoritativeSubagentRef;
        SchemaAuthoritativeTimestampInfo: Schema.AuthoritativeTimestampInfo;
        /**
         * Content Capability
         * @description Optional enriched transcript-content behavior that a client must negotiate before emission
         * @example observed_model
         * @enum {string}
         */
        SchemaContentCapability: Schema.ContentCapability;
        SchemaContentCapabilityAdvertisement: Schema.ContentCapabilityAdvertisement;
        /**
         * Content Capability Version
         * @description Semantic version of one optional enriched transcript-content behavior
         * @example 1.0.0
         * @enum {string}
         */
        SchemaContentCapabilityVersion: Schema.ContentCapabilityVersion;
        /**
         * Entry Type
         * @description Classification of a single entry within an agent session transcript
         * @example text
         * @example tool_use
         * @example tool_result
         * @enum {string}
         */
        SchemaEntryType: Schema.EntryType;
        SchemaExchangeCodeRequest: Schema.ExchangeCodeRequest;
        SchemaExchangeCodeResponse: Schema.ExchangeCodeResponse;
        /**
         * Host Slug
         * @description Sanitized, filesystem-safe identifier derived from git remote; contains only [a-zA-Z0-9._<>-]
         * @example github.com--user--repo
         * @example local--home-user-projects-myapp
         */
        SchemaHostSlug: Schema.HostSlug;
        /**
         * License
         * @description Content license for a published transcript
         * @example CC0-1.0
         * @enum {string}
         */
        SchemaLicense: Schema.License;
        /**
         * Model ID
         * @description Model identifier (e.g. 'claude-opus-4-6', 'gemini-2.0-flash')
         * @example claude-opus-4-6
         * @example gemini-2.0-flash
         * @example codex-mini-latest
         */
        SchemaModelID: Schema.ModelID;
        SchemaOwnerTranscriptUpdateRequest: Schema.OwnerTranscriptUpdateRequest;
        SchemaOwnerTranscriptUpdateResponse: Schema.OwnerTranscriptUpdateResponse;
        /** @description Omitted preserves, null requests clear, and a canonical license replaces */
        SchemaOwnerUpdateLicenseIntent: Schema.OwnerUpdateLicenseIntent;
        /**
         * Project Hash
         * @description SHA-256 hex digest of the project's origin URL or local path
         * @example a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
         */
        SchemaProjectHash: Schema.ProjectHash;
        SchemaProvenance: Schema.Provenance;
        SchemaPublishAppliedState: Schema.PublishAppliedState;
        SchemaPublishNormalizedValues: Schema.PublishNormalizedValues;
        /**
         * Publish Request Fingerprint
         * @description SHA3-256 digest of the canonical domain-separated publish operation
         */
        SchemaPublishRequestFingerprint: Schema.PublishRequestFingerprint;
        SchemaPublishedAssociation: Schema.PublishedAssociation;
        SchemaPullAnnotation: Schema.PullAnnotation;
        SchemaPullListResponse: Schema.PullListResponse;
        SchemaPullSkipGateItem: Schema.PullSkipGateItem;
        SchemaPullSkipGateRequest: Schema.PullSkipGateRequest;
        SchemaPullSkipGateResponse: Schema.PullSkipGateResponse;
        SchemaPullSkipGateResult: Schema.PullSkipGateResult;
        SchemaPullTranscriptInfo: Schema.PullTranscriptInfo;
        /**
         * Role
         * @description Sender role of a message turn
         * @example user
         * @example assistant
         * @enum {string}
         */
        SchemaRole: Schema.Role;
        SchemaSchemaVersionResponse: Schema.SchemaVersionResponse;
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
         * Source Format
         * @description Transcript file format
         * @example jsonl
         * @example json
         * @enum {string}
         */
        SchemaSourceFormat: Schema.SourceFormat;
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
         * Transcript Content Hash
         * @description SHA3-256 digest of the exact transcript file bytes
         */
        SchemaTranscriptContentHash: Schema.TranscriptContentHash;
        /**
         * Transcript ID
         * Format: uuid
         * @description Village-side transcript identifier (canonical lowercase-hex UUID)
         * @example 99d59925-36bc-424c-a789-8be54d9702ba
         */
        SchemaTranscriptID: Schema.TranscriptID;
        /**
         * TranscriptUpdateVisibility
         * @description Visibility values accepted by the owner transcript update operation. Organization-scoped visibility is deferred and is deliberately not offered here.
         * @example public
         * @example private
         * @enum {string}
         */
        SchemaTranscriptUpdateVisibility: Schema.TranscriptUpdateVisibility;
        /**
         * Visibility
         * @description Access control level for a published transcript
         * @example public
         * @example private
         * @enum {string}
         */
        SchemaVisibility: Schema.Visibility;
        /**
         * Visibility Intent
         * @description Optional desired final access for legacy compatibility; content replacement remains private and widening occurs separately
         * @example private
         * @example public
         * @enum {string}
         */
        SchemaVisibilityIntent: Schema.VisibilityIntent;
        /**
         * Visibility
         * @description Access control level for a published transcript
         * @example public
         * @example private
         * @enum {string}
         */
        Visibility: Schema.Visibility;
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    pushAnnotations: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["SchemaAnnotationPushRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaAnnotationPushResponse"];
                };
            };
        };
    };
    getAnnotationManifest: {
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
                    "application/json": components["schemas"]["SchemaAnnotationManifestResponse"];
                };
            };
        };
    };
    cliExchangeCode: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["SchemaExchangeCodeRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaExchangeCodeResponse"];
                };
            };
        };
    };
    cliLogin: {
        parameters: {
            query?: {
                /** @description Local callback server port */
                port?: number;
                /** @description OAuth state parameter for CSRF protection */
                state?: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description No Content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
        };
    };
    listPullableTranscripts: {
        parameters: {
            query?: {
                page?: number;
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
                    "application/json": components["schemas"]["SchemaPullListResponse"];
                };
            };
        };
    };
    pullSkipGate: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["SchemaPullSkipGateRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaPullSkipGateResponse"];
                };
            };
        };
    };
    getPullTranscript: {
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
                    "application/json": components["schemas"]["SchemaPullTranscriptInfo"];
                };
            };
        };
    };
    getPullTranscriptAnnotations: {
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
                    "application/json": null | components["schemas"]["SchemaPullAnnotation"][];
                };
            };
        };
    };
    getPullTranscriptContent: {
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
            /** @description No Content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
        };
    };
    getSchemaVersion: {
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
                    "application/json": components["schemas"]["SchemaSchemaVersionResponse"];
                };
            };
        };
    };
    publishTranscript: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "multipart/form-data": components["schemas"]["FormDataOpenapiTranscriptPublishMultipartRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaAuthoritativePublishResponse"];
                };
            };
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaAuthoritativePublishResponse"];
                };
            };
        };
    };
    updateTranscript: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Transcript identifier */
                id: components["schemas"]["SchemaTranscriptID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaOwnerTranscriptUpdateRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaOwnerTranscriptUpdateResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpenapiTranscriptUpdateErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpenapiTranscriptUpdateErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpenapiTranscriptUpdateErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpenapiTranscriptUpdateErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpenapiTranscriptUpdateErrorResponse"];
                };
            };
        };
    };
}
