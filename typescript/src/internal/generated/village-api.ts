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
        /** @description Publish a transcript with session entries to the village. */
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
        /** @description Update an owned transcript's metadata and governance axes. Every field is optional and an omitted field is left unchanged, resolved against the locked stored row so a concurrent edit is not reverted. License is three-valued: omit to preserve, send the empty string to clear, send a menu license to replace. Clearing a license that was actually granted is refused with 400 because a granted Creative Commons license is irrevocable. Only the owner may call this; anyone else receives 403 and neither the transcript nor its governance audit changes. Visibility accepts private and public; organization-scoped visibility is deferred. The village additionally accepts and stores a legacy 'shared' value that is deliberately NOT declared: it is not a member of this contract's Visibility enum at all, whose third member is 'group', and the village refuses 'group'. Declaring 'shared' would mean inventing an enum member to expose the deferred organization-ACL capability, so its absence is a decision. The transcript id is likewise narrower than the server: uuid.Parse also accepts uppercase, brace-wrapped, urn:uuid-prefixed and 32-undashed-hex spellings that the declared pattern rejects. Those stay undeclared because the village only ever emits the canonical lowercase form, so no client holds another unless it manufactures one, and accepting five spellings for one identity is itself a defect surface. Note this document now describes one transcript id two ways: this operation constrains it to the canonical pattern, while the older pull operations still declare a bare string. That difference is not a contradiction about what the village accepts, only about what each operation declares; the pull operations are deliberately untouched here. Omit a field to leave it unchanged; send an empty string to clear a title, a description, or a license. Explicit null is refused on every field, because the server would read it as preserve rather than the clear a caller usually intends, and an unknown field is refused because the server would accept and silently discard it. 400 covers five distinct refusals: an unparseable transcript id, an undecodable body, a visibility outside the accepted set, a license outside the canonical menu, and the attempt to clear a granted license. 401 is returned by the authentication boundary before the handler runs, and is distinct from 403: 401 means the credential is missing or expired and the caller should re-authenticate, while 403 means the caller is authenticated but does not own this transcript. 404 covers both a transcript that does not exist and a lookup that failed, so it must not be read as proof of absence. 500 has two forms and only one carries this body: the handler's own failure returns the envelope, while a panic recovered by the router's middleware returns 500 with an EMPTY body, so a client must tolerate an absent body on 500 rather than assuming the envelope is always present. The refusals are declared while the 200 body is NOT, and that asymmetry is deliberate rather than an oversight. A client must distinguish 403 from 404 from each 400 to tell a user anything useful, so those distinctions are exactly what this contract is for. The success body has no such consumer: the applied state is read back through GET /api/v1/pull/transcripts/{id}. The village does return a 200 body, but it currently serves an untyped object wrapping the stored row's internal columns (owner_id, blob_key, project_hash, source_file_path and others) at backend/internal/handler/transcripts.go:723-727, and those columns serialize through pgtype wrappers, so a consumer would receive {"String":"x","Valid":true} where it expects a string. Declaring a projection the village does not serve would break the property that the served and declared contracts cannot drift, so nothing is declared until the handler serves a shape worth declaring. Tracked at https://github.com/peasant-labs/village/issues/55; adding the response schema later is additive. Do not 'harmonize' this by inventing a success body. */
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
        SchemaPublishedAssociation: Schema.PublishedAssociation;
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
         * @description Unique session identifier (UUID, agent-{hex}, ses_{id}, sess_{id} (ACP), msg_{id}, or a Strike session ID)
         * @example 99d59925-36bc-424c-a789-8be54d9702ba
         * @example agent-a3aee4f
         * @example ses_3cd91f52effeXd3QAJ54jOyzv5
         * @example sess_3cd91f52effeXd3QAJ54jOyzv5
         * @example 20260728T123456.123456789Z-ABCDEFGHIJKLMNOPQRST234567
         * @example ABCDEFGHIJKLMNOPQRST234567
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
         * @description What is being annotated: session-level, entry-level (turn/tool call), meta-annotation, project-level, a specific file version (content-hash keyed read-state receipt), or a durable session-to-commit association
         * @example session
         * @example entry
         * @example file_version
         * @example association
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
         * TranscriptUpdateLicense
         * @description License value for the owner transcript update operation: a canonical menu license, or the empty string to clear. Clearing a license that was actually granted is refused, because a granted Creative Commons license is irrevocable.
         * @example CC-BY-4.0
         * @example
         * @enum {string}
         */
        SchemaTranscriptUpdateLicense: Schema.TranscriptUpdateLicense;
        SchemaTranscriptUpdateRequest: Schema.TranscriptUpdateRequest;
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
                "application/json": components["schemas"]["SchemaTranscriptUpdateRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
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
