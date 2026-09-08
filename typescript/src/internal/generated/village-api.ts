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
    "/api/v1/groups": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List collectives the authenticated caller belongs to, with the caller's role and membership time. */
        get: operations["listGroups"];
        put?: never;
        /** @description Create a collective. Name is required; acceptance_mode defaults to open and data_access defaults to members_only when omitted. */
        post: operations["createGroup"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/public": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List collectives for the public directory as compact rows with member and approved-transcript counts. */
        get: operations["listPublicGroups"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/search": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Search visible collectives. An empty q returns an empty collectives array; limit defaults to 10 and is capped at 50. */
        get: operations["searchCollectives"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/visible": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List collectives the authenticated caller may see. Role and member_since are null when the caller sees the collective through public or open visibility but is not a member. */
        get: operations["listVisibleGroups"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get a collective with roster, stats, model breakdown, contributors, the caller's role, and a transcript page. pending_members is present only for owners. transcripts is an empty array when the caller may see the collective but may not read its data. */
        get: operations["getGroup"];
        put?: never;
        post?: never;
        /** @description Delete an owned collective and return a status receipt. */
        delete: operations["deleteGroup"];
        options?: never;
        head?: never;
        /** @description Update an owned collective. Name and description are written from the request body's string values, so a caller that wants to preserve them must send their current values. linked_github_org omitted means preserve, empty string clears, and non-empty sets after visible-org validation. */
        patch: operations["updateGroup"];
        trace?: never;
    };
    "/api/v1/groups/{id}/contributable": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List every transcript the authenticated caller may offer to one collective, plus whether each transcript is already live in that collective. The response is deliberately unpaginated and bounded by the server row limit. */
        get: operations["listContributableTranscripts"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/join": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Join an open or verified collective as a contributor. Curated collectives require invitation and refuse this route. */
        post: operations["joinGroup"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/members": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Add a platform user to an owned collective by username. */
        post: operations["addGroupMember"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/members/{userID}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** @description Remove a member from a collective. The member may remove themselves; removing somebody else requires owner access. retract=true or a mandatory deletion policy retracts that member's live submissions. */
        delete: operations["removeGroupMember"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/members/{userID}/role": {
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
        /** @description Change a collective member's role to contributor or member. Owner and pending are not accepted by this operation. */
        patch: operations["updateGroupMemberRole"];
        trace?: never;
    };
    "/api/v1/groups/{id}/my-shares": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List the authenticated caller's own live submissions to one collective, including pending and approved rows. */
        get: operations["listMyGroupShares"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/pending": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List pending transcript submissions for an owned collective, oldest first. */
        get: operations["listPendingShares"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/repositories": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List GitHub repositories linked to a collective. The caller must be an authenticated collective member. */
        get: operations["listGroupRepositories"];
        put?: never;
        /** @description Link a GitHub repository to an owned collective after validating that the configured GitHub App installation can access it. */
        post: operations["linkGroupRepository"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/repositories/{owner}/{name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** @description Unlink a GitHub repository from an owned collective. */
        delete: operations["unlinkGroupRepository"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/repositories/{owner}/{name}/commits": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List cached commits for a linked repository. refresh=true fetches from GitHub and requires collective owner access; otherwise any collective member may read the cache. */
        get: operations["listGroupRepositoryCommits"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/groups/{id}/shares": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Offer selected transcripts from one owned project to one collective in one transaction. Omitted or empty transcript_ids means every owned transcript in the project. visibility_confirmed must be true when any selected private transcript would become visible to the collective. */
        post: operations["batchShareProject"];
        delete?: never;
        options?: never;
        head?: never;
        /** @description Approve or reject many pending submissions in one owner-only action. The response separates ids that were decided by this action from ids that were already decided, never submitted to this collective, or stale. */
        patch: operations["batchReviewShares"];
        trace?: never;
    };
    "/api/v1/groups/{id}/shares/{transcriptID}": {
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
        /** @description Approve or reject one pending submission to an owned collective. The decision changes the share-attempt status and does not mutate transcript visibility. */
        patch: operations["reviewShare"];
        trace?: never;
    };
    "/api/v1/groups/{id}/transcripts/{transcriptID}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** @description Remove or revoke a transcript contribution from an owned collective without deleting the transcript from its owner's library. */
        delete: operations["removeGroupTranscript"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/integrations/github/webhook": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Receive a GitHub App webhook delivery. The body is GitHub's payload, authenticated by the X-Hub-Signature-256 HMAC over the raw body; X-GitHub-Delivery makes redelivery idempotent. Handles installation, pull_request, check_run, and issue_comment events and acknowledges every other event without acting. Returns 501 when the App or its webhook secret is not configured. */
        post: operations["receiveGitHubWebhook"];
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
    "/api/v1/pulls/{owner}/{name}/{number}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Read one pull request's prompt attachment. The digest is null until the attachment reaches preview or attached; viewer_is_author is true when the caller is the pull request author. Readers who may not see the attached transcripts receive 403. */
        get: operations["getPullRequestAttachment"];
        put?: never;
        post?: never;
        /** @description Detach the prompts from a pull request. Only the pull request author may detach. Deletes the comment, resets the check, and restores each transcript's previous visibility; an attachment already detached returns 409. */
        delete: operations["detachPullRequestAttachment"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/pulls/{owner}/{name}/{number}/confirm": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Confirm a preview and post it. Only the pull request author may confirm; the attachment must be in preview, otherwise 409. A failed post to GitHub returns 502 and leaves the attachment in preview. */
        post: operations["confirmPullRequestAttachment"];
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
        /** @description Advertise the annotation schema version plus push and pull contract windows and the deployment-specific content-capability set. contentCapabilities omitted or empty means none and null is invalid. Its items are opaque, forward-open revision strings: clients match exactly without SemVer or suffix parsing, ignore unknown tokens, and tolerate/deduplicate duplicates; server output contains only pinned known tokens, rejects duplicates, has no semantic order, and is serialized lexicographically. observed_model_v1 is required exactly when any root or nested assistant turn carries observedModel, not for session model alone. It guarantees assistant-only value validation before persistence with no invalid database or blob side effects and byte-exact accepted observedModel strings through storage, typed migration, rewrite, serving, and pull; JSON whitespace and key order are not guaranteed. */
        get: operations["getSchemaVersion"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcript-groups/{groupId}/members": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Page the authorized saved helper threads selected by an originating grouped list. The required opaque scope is replayed with current authorization and cannot widen the original route filters. */
        get: operations["listTranscriptGroupMembers"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List authorized transcripts using the existing global, recent, explore, profile, project, and search filters. view=grouped returns grouped list items; omitting view preserves the flat transcript array. */
        get: operations["listTranscripts"];
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
        /** @description Publish exact transcript bytes with typed JSON metadata. Capability negotiation is deployment-specific and uses exact string membership, never SemVer ordering. A client with required capabilities missing from the destination MUST refuse before remote upload and MUST NOT strip evidence; dry-run negotiation is local only. A server advertising observed_model_v1 MUST validate assistant-only attribution and observedModel values before persistence or any side effect, and MUST preserve accepted observedModel string bytes through storage, migration, rewrite, and serving; JSON formatting is excluded. Assistant includes nested subagent output represented by role assistant. Creation returns 201 and replacement returns 200; both carry the complete authoritative receipt. */
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
        /** @description Get transcript metadata and viewer-authorized relationship navigation. Navigation is read-only and is never stored in transcript content. */
        get: operations["getTranscriptMetadata"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** @description Update an owned transcript's metadata and governance axes. Every field is optional: omission preserves stored state; empty title, description, or tags clear those fields; license null requests clear; and a canonical license replaces. Explicit null for title, description, tags, or visibility is rejected, as are unknown fields and invalid or duplicate tags. Clearing an already granted Creative Commons license remains subject to the server's irrevocability rule. Only the owner may call this operation. A successful update returns the complete typed authoritative editable state, including the canonical transcript URL and positive update timestamp. Visibility accepts private and public; organization-scoped visibility remains deferred. */
        patch: operations["updateTranscript"];
        trace?: never;
    };
    "/api/v1/transcripts/{id}/collectives": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List visible collectives that hold a transcript. The server returns an empty collectives array, not a refusal, when memberships are hidden by collective visibility or contributor opt-in policy. */
        get: operations["listTranscriptCollectives"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts/{id}/content": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Get the durable transcript content envelope. This response never includes authorized relationship navigation. */
        get: operations["getTranscriptContent"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts/{id}/share": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** @description Offer one owned transcript to one or more collectives. The response is the transcript's share list, while duplicate or refused collectives are skipped unless every requested collective is already live. */
        post: operations["shareTranscriptWithGroups"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/transcripts/{id}/share/{groupID}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** @description Withdraw one owned transcript from one collective and return a status receipt. */
        delete: operations["unshareTranscriptFromGroup"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/collectives/contributions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List collectives the authenticated caller has offered transcripts to, with approved and pending transcript counts and rejected and withdrawn attempt counts. */
        get: operations["listMyCollectiveContributions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/collectives/{groupId}/submissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List every transcript-collective pair the authenticated caller has offered to one collective, including pairs whose latest event was a withdrawal. A caller with no pair receives 404, matching no-such-collective. */
        get: operations["listMyCollectiveSubmissions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/collectives/{groupId}/transcripts/{transcriptId}/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List the full owner-only event history for one transcript and collective pair, oldest event first. The decided_by_actor field is an actor class, never a moderator user id. */
        get: operations["listShareEventHistory"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/prompt-requests": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description List the caller's attachments that are waiting for a transcript from the caller's machine, with each repository's normalized remote so a client can match the repository it is about to push. */
        get: operations["listMyPromptRequests"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/users/me/settings": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** @description Read the caller's settings, including preview_before_attach. */
        get: operations["getMySettings"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** @description Patch the caller's settings. A field that is omitted is left unchanged. */
        patch: operations["updateMySettings"];
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
        /** @enum {string} */
        OpenapiVillageGroupedView: "grouped";
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
        SchemaChildSessionRef: Schema.ChildSessionRef;
        SchemaCommandInvocation: Schema.CommandInvocation;
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
        /**
         * Digest Item Kind
         * @description Kind of one item in the prompt digest chain: a session boundary, a human prompt, a skill invocation marker, or a commit anchor
         * @example session
         * @example prompt
         * @example skill
         * @example commit
         * @enum {string}
         */
        SchemaDigestItemKind: Schema.DigestItemKind;
        SchemaEarlierHistorySection: Schema.EarlierHistorySection;
        /**
         * Earlier History State
         * @description Closed session graph value
         * @example uncertain_migrated
         * @example uncertain_unresolved
         * @enum {string}
         */
        SchemaEarlierHistoryState: Schema.EarlierHistoryState;
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
        SchemaExchangeCodeRequest: Schema.ExchangeCodeRequest;
        SchemaExchangeCodeResponse: Schema.ExchangeCodeResponse;
        SchemaHelperContextSummary: Schema.HelperContextSummary;
        SchemaHelperGroupSummary: Schema.HelperGroupSummary;
        /**
         * Host Slug
         * @description Sanitized, filesystem-safe identifier derived from git remote; contains only [a-zA-Z0-9._<>-]
         * @example github.com--user--repo
         * @example local--home-user-projects-myapp
         */
        SchemaHostSlug: Schema.HostSlug;
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
        SchemaPromptDigest: Schema.PromptDigest;
        SchemaPromptDigestHeader: Schema.PromptDigestHeader;
        SchemaPromptDigestItem: Schema.PromptDigestItem;
        SchemaPromptDigestSkill: Schema.PromptDigestSkill;
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
        /**
         * Role
         * @description Sender role of a message turn
         * @example user
         * @example assistant
         * @enum {string}
         */
        SchemaRole: Schema.Role;
        SchemaSchemaVersionResponse: Schema.SchemaVersionResponse;
        SchemaSessionDetailPayload: Schema.SessionDetailPayload;
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
        /**
         * Source Entry Reference
         * Format: public-ref-utf8-96-bytes
         */
        SchemaSourceEntryRef: Schema.SourceEntryRef;
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
        SchemaTurnDetail: Schema.TurnDetail;
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
        /**
         * Village Assignable Group Role
         * @description Collective roles an owner may assign through the member role endpoint
         * @example contributor
         * @example member
         * @enum {string}
         */
        SchemaVillageAssignableGroupRole: Schema.VillageAssignableGroupRole;
        SchemaVillageBatchReviewRequest: Schema.VillageBatchReviewRequest;
        SchemaVillageBatchReviewResponse: Schema.VillageBatchReviewResponse;
        SchemaVillageBatchShareEntry: Schema.VillageBatchShareEntry;
        SchemaVillageBatchShareRequest: Schema.VillageBatchShareRequest;
        SchemaVillageBatchShareResponse: Schema.VillageBatchShareResponse;
        SchemaVillageCollectiveSearchResponse: Schema.VillageCollectiveSearchResponse;
        SchemaVillageCollectiveSearchResult: Schema.VillageCollectiveSearchResult;
        SchemaVillageCollectiveSubmission: Schema.VillageCollectiveSubmission;
        SchemaVillageContributableResponse: Schema.VillageContributableResponse;
        SchemaVillageContributableTranscript: Schema.VillageContributableTranscript;
        SchemaVillageContributedCollective: Schema.VillageContributedCollective;
        SchemaVillageContributedCollectivesResponse: Schema.VillageContributedCollectivesResponse;
        /**
         * Village Contribution Status
         * @description Status assigned when a new collective contribution is opened
         * @example approved
         * @example pending
         * @enum {string}
         */
        SchemaVillageContributionStatus: Schema.VillageContributionStatus;
        SchemaVillageCreateGroupRequest: Schema.VillageCreateGroupRequest;
        SchemaVillageEnrichedTranscriptShare: Schema.VillageEnrichedTranscriptShare;
        SchemaVillageErrorResponse: Schema.VillageErrorResponse;
        /**
         * GitHub Webhook Payload
         * @description GitHub event payload forwarded verbatim; validated by the X-Hub-Signature-256 HMAC over the raw body, never by shape
         */
        SchemaVillageGitHubWebhookPayload: Schema.VillageGitHubWebhookPayload;
        SchemaVillageGroup: Schema.VillageGroup;
        /**
         * Village Group Acceptance Mode
         * @description How a collective accepts new members and contributions
         * @example open
         * @example verified_only
         * @example curated
         * @enum {string}
         */
        SchemaVillageGroupAcceptanceMode: Schema.VillageGroupAcceptanceMode;
        SchemaVillageGroupContributor: Schema.VillageGroupContributor;
        /**
         * Village Group Data Access
         * @description Who may read a collective's pooled transcript data
         * @example members_only
         * @example contributors
         * @example public
         * @enum {string}
         */
        SchemaVillageGroupDataAccess: Schema.VillageGroupDataAccess;
        SchemaVillageGroupDetailResponse: Schema.VillageGroupDetailResponse;
        SchemaVillageGroupMember: Schema.VillageGroupMember;
        SchemaVillageGroupMemberRoleRequest: Schema.VillageGroupMemberRoleRequest;
        SchemaVillageGroupMemberUsernameRequest: Schema.VillageGroupMemberUsernameRequest;
        SchemaVillageGroupModelBreakdown: Schema.VillageGroupModelBreakdown;
        /**
         * Village Group Role
         * @description A user's role in one collective
         * @example owner
         * @example member
         * @example contributor
         * @example pending
         * @enum {string}
         */
        SchemaVillageGroupRole: Schema.VillageGroupRole;
        SchemaVillageGroupStatusRoleResponse: Schema.VillageGroupStatusRoleResponse;
        SchemaVillageGroupTranscript: Schema.VillageGroupTranscript;
        SchemaVillageGroupTranscriptStats: Schema.VillageGroupTranscriptStats;
        /**
         * Village Group Viewer Role
         * @description The caller's role in a collective, or an empty string when the caller has none
         * @example
         * @example owner
         * @example member
         * @example contributor
         * @example pending
         * @enum {string}
         */
        SchemaVillageGroupViewerRole: Schema.VillageGroupViewerRole;
        SchemaVillageGroupedContributableResponse: Schema.VillageGroupedContributableResponse;
        SchemaVillageGroupedGroupDetailResponse: Schema.VillageGroupedGroupDetailResponse;
        SchemaVillageHelperMembersPayload: Schema.VillageHelperMembersPayload;
        SchemaVillageLinkRepositoryRequest: Schema.VillageLinkRepositoryRequest;
        SchemaVillageLinkedRepositoriesResponse: Schema.VillageLinkedRepositoriesResponse;
        SchemaVillageLinkedRepository: Schema.VillageLinkedRepository;
        SchemaVillageListTranscriptAttestation: Schema.VillageListTranscriptAttestation;
        SchemaVillageListUserOrganization: Schema.VillageListUserOrganization;
        SchemaVillageMetadataUserOrganization: Schema.VillageMetadataUserOrganization;
        SchemaVillagePendingShare: Schema.VillagePendingShare;
        /**
         * Village Project Name Source
         * @description Which source tier produced a resolved project display name
         * @example override
         * @example consented
         * @example remote
         * @example path
         * @example privacy
         * @enum {string}
         */
        SchemaVillageProjectNameSource: Schema.VillageProjectNameSource;
        SchemaVillagePromptRequest: Schema.VillagePromptRequest;
        SchemaVillagePromptRequestsResponse: Schema.VillagePromptRequestsResponse;
        /**
         * Village Prompts Check Mode
         * @description Check-run conclusion policy for a collective's linked repositories when no prompts are attached
         * @example informational
         * @example required
         * @enum {string}
         */
        SchemaVillagePromptsCheckMode: Schema.VillagePromptsCheckMode;
        SchemaVillagePublicGroup: Schema.VillagePublicGroup;
        SchemaVillagePullRequestAttachedTranscript: Schema.VillagePullRequestAttachedTranscript;
        SchemaVillagePullRequestAttachment: Schema.VillagePullRequestAttachment;
        SchemaVillagePullRequestAttachmentResponse: Schema.VillagePullRequestAttachmentResponse;
        /**
         * Village Pull Request Attachment State
         * @description Current lifecycle state of a pull request's prompt attachment
         * @example requested
         * @example waiting
         * @example preview
         * @example attached
         * @example detached
         * @enum {string}
         */
        SchemaVillagePullRequestAttachmentState: Schema.VillagePullRequestAttachmentState;
        SchemaVillageRemoveGroupMemberResponse: Schema.VillageRemoveGroupMemberResponse;
        SchemaVillageRepositoryCommit: Schema.VillageRepositoryCommit;
        SchemaVillageRepositoryCommitsResponse: Schema.VillageRepositoryCommitsResponse;
        /**
         * Village Review Decision
         * @description Decision applied by a collective owner to pending submissions
         * @example approved
         * @example rejected
         * @enum {string}
         */
        SchemaVillageReviewDecision: Schema.VillageReviewDecision;
        SchemaVillageReviewShareRequest: Schema.VillageReviewShareRequest;
        SchemaVillageReviewShareResponse: Schema.VillageReviewShareResponse;
        SchemaVillageSessionListItem: Schema.VillageSessionListItem;
        SchemaVillageSessionListPayload: Schema.VillageSessionListPayload;
        SchemaVillageSessionRow: Schema.VillageSessionRow;
        SchemaVillageShareEvent: Schema.VillageShareEvent;
        /**
         * Village Share Event Actor
         * @description Actor class that decided one collective share event
         * @example
         * @example owner
         * @example collective
         * @example moderator
         * @enum {string}
         */
        SchemaVillageShareEventActor: Schema.VillageShareEventActor;
        /**
         * Village Share Status
         * @description Status of one collective share-attempt event. Pending, approved, and rejected can appear in current projections; retracted and revoked are terminal ledger states.
         * @example pending
         * @example approved
         * @example rejected
         * @example retracted
         * @example revoked
         * @enum {string}
         */
        SchemaVillageShareStatus: Schema.VillageShareStatus;
        SchemaVillageShareTranscriptRequest: Schema.VillageShareTranscriptRequest;
        SchemaVillageStatusResponse: Schema.VillageStatusResponse;
        SchemaVillageTag: Schema.VillageTag;
        SchemaVillageTranscript: Schema.VillageTranscript;
        SchemaVillageTranscriptAttestation: Schema.VillageTranscriptAttestation;
        SchemaVillageTranscriptCollective: Schema.VillageTranscriptCollective;
        SchemaVillageTranscriptCollectivesResponse: Schema.VillageTranscriptCollectivesResponse;
        /**
         * Village Transcript Deletion Policy
         * @description Whether leaving a collective retracts contributed transcripts by default
         * @example user_choice
         * @example mandatory
         * @enum {string}
         */
        SchemaVillageTranscriptDeletionPolicy: Schema.VillageTranscriptDeletionPolicy;
        SchemaVillageTranscriptListResponse: Schema.VillageTranscriptListResponse;
        SchemaVillageTranscriptListRow: Schema.VillageTranscriptListRow;
        SchemaVillageTranscriptMetadataResponse: Schema.VillageTranscriptMetadataResponse;
        SchemaVillageTranscriptShare: Schema.VillageTranscriptShare;
        /**
         * Village Transcript Visibility
         * @description Village transcript visibility as stored and served by web routes
         * @example private
         * @example shared
         * @example public
         * @enum {string}
         */
        SchemaVillageTranscriptVisibility: Schema.VillageTranscriptVisibility;
        /**
         * Village UUID
         * Format: uuid
         * @description Village-side canonical lowercase UUID identifier
         * @example 123e4567-e89b-12d3-a456-426614174000
         */
        SchemaVillageUUID: Schema.VillageUUID;
        SchemaVillageUpdateGroupRequest: Schema.VillageUpdateGroupRequest;
        SchemaVillageUpdateUserSettingsRequest: Schema.VillageUpdateUserSettingsRequest;
        SchemaVillageUser: Schema.VillageUser;
        SchemaVillageUserGroup: Schema.VillageUserGroup;
        SchemaVillageUserGroupShare: Schema.VillageUserGroupShare;
        SchemaVillageUserSettings: Schema.VillageUserSettings;
        SchemaVillageVisibleGroup: Schema.VillageVisibleGroup;
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
    listGroups: {
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
                    "application/json": null | components["schemas"]["SchemaVillageUserGroup"][];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    createGroup: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageCreateGroupRequest"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageGroup"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listPublicGroups: {
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
                    "application/json": null | components["schemas"]["SchemaVillagePublicGroup"][];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    searchCollectives: {
        parameters: {
            query?: {
                q?: string;
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
                    "application/json": components["schemas"]["SchemaVillageCollectiveSearchResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listVisibleGroups: {
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
                    "application/json": null | components["schemas"]["SchemaVillageVisibleGroup"][];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    getGroup: {
        parameters: {
            query?: {
                limit?: number;
                offset?: number;
                /** @description Set to grouped to request helper-group list items; omission preserves the legacy response. */
                view?: components["schemas"]["OpenapiVillageGroupedView"];
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageGroupDetailResponse"] | components["schemas"]["SchemaVillageGroupedGroupDetailResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    deleteGroup: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    updateGroup: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageUpdateGroupRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageGroup"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listContributableTranscripts: {
        parameters: {
            query?: {
                /** @description Set to grouped to request helper-group list items; omission preserves the legacy response. */
                view?: components["schemas"]["OpenapiVillageGroupedView"];
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageContributableResponse"] | components["schemas"]["SchemaVillageGroupedContributableResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Request Entity Too Large */
            413: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    joinGroup: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageGroupStatusRoleResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    addGroupMember: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageGroupMemberUsernameRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    removeGroupMember: {
        parameters: {
            query?: {
                retract?: boolean;
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Member user identifier */
                userID: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageRemoveGroupMemberResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    updateGroupMemberRole: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Member user identifier */
                userID: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageGroupMemberRoleRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageGroupStatusRoleResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listMyGroupShares: {
        parameters: {
            query?: {
                /** @description Set to grouped to request helper-group list items; omission preserves the legacy response. */
                view?: components["schemas"]["OpenapiVillageGroupedView"];
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageUserGroupShare"][] | components["schemas"]["SchemaVillageSessionListPayload"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listPendingShares: {
        parameters: {
            query?: {
                /** @description Set to grouped to request helper-group list items; omission preserves the legacy response. */
                view?: components["schemas"]["OpenapiVillageGroupedView"];
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillagePendingShare"][] | components["schemas"]["SchemaVillageSessionListPayload"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listGroupRepositories: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageLinkedRepositoriesResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    linkGroupRepository: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageLinkRepositoryRequest"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageLinkedRepository"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Implemented */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    unlinkGroupRepository: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Repository owner login */
                owner: string;
                /** @description Repository name */
                name: string;
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
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listGroupRepositoryCommits: {
        parameters: {
            query?: {
                refresh?: boolean;
            };
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Repository owner login */
                owner: string;
                /** @description Repository name */
                name: string;
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
                    "application/json": components["schemas"]["SchemaVillageRepositoryCommitsResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Implemented */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Bad Gateway */
            502: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    batchShareProject: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageBatchShareRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageBatchShareResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unprocessable Entity */
            422: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    batchReviewShares: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageBatchReviewRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageBatchReviewResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    reviewShare: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Transcript identifier */
                transcriptID: components["schemas"]["SchemaTranscriptID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageReviewShareRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageReviewShareResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    removeGroupTranscript: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                id: components["schemas"]["SchemaVillageUUID"];
                /** @description Transcript identifier */
                transcriptID: components["schemas"]["SchemaTranscriptID"];
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
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    receiveGitHubWebhook: {
        parameters: {
            query?: never;
            header: {
                /** @description GitHub event name */
                "X-GitHub-Event": string;
                /** @description Unique delivery identifier; Village records it so a redelivery is idempotent */
                "X-GitHub-Delivery": string;
                /** @description HMAC SHA-256 of the raw body under the App webhook secret */
                "X-Hub-Signature-256": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageGitHubWebhookPayload"];
            };
        };
        responses: {
            /** @description Accepted */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Implemented */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
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
    getPullRequestAttachment: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Repository owner login */
                owner: string;
                /** @description Repository name */
                name: string;
                /** @description Pull request number */
                number: number;
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
                    "application/json": components["schemas"]["SchemaVillagePullRequestAttachmentResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    detachPullRequestAttachment: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Repository owner login */
                owner: string;
                /** @description Repository name */
                name: string;
                /** @description Pull request number */
                number: number;
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
                    "application/json": components["schemas"]["SchemaVillagePullRequestAttachmentResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Bad Gateway */
            502: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    confirmPullRequestAttachment: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Repository owner login */
                owner: string;
                /** @description Repository name */
                name: string;
                /** @description Pull request number */
                number: number;
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
                    "application/json": components["schemas"]["SchemaVillagePullRequestAttachmentResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Bad Gateway */
            502: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
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
    listTranscriptGroupMembers: {
        parameters: {
            query: {
                /** @description Opaque scope returned by the originating grouped list */
                scope: string;
                page?: number;
                limit?: number;
            };
            header?: never;
            path: {
                /** @description Stable helper group identifier */
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
                    "application/json": components["schemas"]["SchemaVillageHelperMembersPayload"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listTranscripts: {
        parameters: {
            query?: {
                /** @description Set to grouped to request helper-group list items; omission preserves the flat response. */
                view?: components["schemas"]["OpenapiVillageGroupedView"];
                page?: number;
                limit?: number;
                q?: string;
                provider?: string;
                owner?: string;
                project?: string;
                repo?: string;
                org?: string;
                tags?: string;
                origin?: components["schemas"]["SchemaSessionOrigin"];
                sort?: string;
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
                    "application/json": components["schemas"]["SchemaVillageTranscriptListResponse"] | components["schemas"]["SchemaVillageSessionListPayload"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
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
    getTranscriptMetadata: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Transcript identifier */
                id: components["schemas"]["SchemaTranscriptID"];
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
                    "application/json": components["schemas"]["SchemaVillageTranscriptMetadataResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
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
    listTranscriptCollectives: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Transcript identifier */
                id: components["schemas"]["SchemaTranscriptID"];
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
                    "application/json": components["schemas"]["SchemaVillageTranscriptCollectivesResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    getTranscriptContent: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Transcript identifier */
                id: components["schemas"]["SchemaTranscriptID"];
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
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    shareTranscriptWithGroups: {
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
                "application/json": components["schemas"]["SchemaVillageShareTranscriptRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": null | components["schemas"]["SchemaVillageTranscriptShare"][];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Conflict */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    unshareTranscriptFromGroup: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Transcript identifier */
                id: components["schemas"]["SchemaTranscriptID"];
                /** @description Collective identifier */
                groupID: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": components["schemas"]["SchemaVillageStatusResponse"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Forbidden */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listMyCollectiveContributions: {
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
                    "application/json": components["schemas"]["SchemaVillageContributedCollectivesResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listMyCollectiveSubmissions: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                groupId: components["schemas"]["SchemaVillageUUID"];
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
                    "application/json": null | components["schemas"]["SchemaVillageCollectiveSubmission"][];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listShareEventHistory: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Collective identifier */
                groupId: components["schemas"]["SchemaVillageUUID"];
                /** @description Transcript identifier */
                transcriptId: components["schemas"]["SchemaTranscriptID"];
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
                    "application/json": null | components["schemas"]["SchemaVillageShareEvent"][];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Not Found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    listMyPromptRequests: {
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
                    "application/json": components["schemas"]["SchemaVillagePromptRequestsResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    getMySettings: {
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
                    "application/json": components["schemas"]["SchemaVillageUserSettings"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
    updateMySettings: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SchemaVillageUpdateUserSettingsRequest"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageUserSettings"];
                };
            };
            /** @description Bad Request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
            /** @description Internal Server Error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SchemaVillageErrorResponse"];
                };
            };
        };
    };
}
