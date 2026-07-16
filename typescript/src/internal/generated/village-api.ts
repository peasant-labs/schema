import type * as Schema from "../../index.js";
export interface paths {
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
        /** @description Publish a transcript with session entries to the village. */
        post: operations["publishTranscript"];
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
        /**
         * Harness
         * @description AI coding tool or development environment
         * @example claude-code
         * @example opencode
         * @enum {string}
         */
        BestiaryHarness: Schema.Harness;
        OpenapiTranscriptPublishRequest: {
            diagnostics?: components["schemas"]["SchemaDiagnosticsInfo"];
            entries?: components["schemas"]["SchemaSessionEntry"][];
            git?: components["schemas"]["SchemaGitContext"];
            identity?: components["schemas"]["SchemaSessionIdentity"];
            license?: components["schemas"]["SchemaLicense"];
            model: components["schemas"]["SchemaModelInfo"];
            project?: components["schemas"]["SchemaProjectContext"];
            quality?: components["schemas"]["SchemaQualityMetrics"];
            source?: components["schemas"]["SchemaSourceInfo"];
            stats?: components["schemas"]["SchemaSessionStats"];
            subagents?: components["schemas"]["SchemaSubagentRef"][];
            timestamp?: components["schemas"]["SchemaTimestampInfo"];
        };
        SchemaAnnotationManifestResponse: Schema.AnnotationManifestResponse;
        /**
         * Annotator Kind
         * @description Type of entity that produced an annotation: human, agent (AI model), or rule (automated classifier)
         * @example human
         * @example agent
         * @example rule
         * @enum {string}
         */
        SchemaAnnotatorKind: Schema.AnnotatorKind;
        SchemaCommitInfo: Schema.CommitInfo;
        SchemaDiagnosticEntry: Schema.DiagnosticEntry;
        SchemaDiagnosticsInfo: Schema.DiagnosticsInfo;
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
        SchemaGitContext: Schema.GitContext;
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
        SchemaModelInfo: Schema.ModelInfo;
        SchemaProjectContext: Schema.ProjectContext;
        /**
         * Project Hash
         * @description SHA-256 hex digest of the project's origin URL or local path
         * @example a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
         */
        SchemaProjectHash: Schema.ProjectHash;
        SchemaProvenance: Schema.Provenance;
        SchemaPublishResponse: Schema.PublishResponse;
        SchemaPullAnnotation: Schema.PullAnnotation;
        SchemaPullListResponse: Schema.PullListResponse;
        SchemaPullSkipGateItem: Schema.PullSkipGateItem;
        SchemaPullSkipGateRequest: Schema.PullSkipGateRequest;
        SchemaPullSkipGateResponse: Schema.PullSkipGateResponse;
        SchemaPullSkipGateResult: Schema.PullSkipGateResult;
        SchemaPullTranscriptInfo: Schema.PullTranscriptInfo;
        SchemaQualityMetrics: Schema.QualityMetrics;
        /**
         * Role
         * @description Sender role of a message turn
         * @example user
         * @example assistant
         * @enum {string}
         */
        SchemaRole: Schema.Role;
        SchemaSchemaVersionResponse: Schema.SchemaVersionResponse;
        SchemaSessionEntry: Schema.SessionEntry;
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
        SchemaSessionIdentity: Schema.SessionIdentity;
        /**
         * Session Outcome
         * @description Resolution status of the session
         * @example resolved
         * @example partial
         * @example failed
         * @enum {string}
         */
        SchemaSessionOutcome: Schema.SessionOutcome;
        SchemaSessionStats: Schema.SessionStats;
        /**
         * Source Format
         * @description Transcript file format
         * @example jsonl
         * @example json
         * @enum {string}
         */
        SchemaSourceFormat: Schema.SourceFormat;
        SchemaSourceInfo: Schema.SourceInfo;
        /**
         * Stop Reason
         * @description Reason why a session or turn ended (ACP-aligned)
         * @example end_turn
         * @example max_tokens
         * @enum {string}
         */
        SchemaStopReason: Schema.StopReason;
        SchemaSubagentRef: Schema.SubagentRef;
        /**
         * Target Kind
         * @description What is being annotated: session-level, entry-level (turn/tool call), meta-annotation, or project-level
         * @example session
         * @example entry
         * @enum {string}
         */
        SchemaTargetKind: Schema.TargetKind;
        SchemaTimestampInfo: Schema.TimestampInfo;
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
        /**
         * Visibility
         * @description Access control level for a published transcript
         * @example public
         * @example private
         * @enum {string}
         */
        SchemaVisibility: Schema.Visibility;
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
        requestBody?: {
            content: {
                "application/json": components["schemas"]["OpenapiTranscriptPublishRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaPublishResponse"];
                };
            };
        };
    };
}
