-- Session entries: indexed per-turn/per-part rows for transcript content.
-- Maps 1:1 to pkg/schema.SessionEntry (camelCase wire) with snake_case columns.
CREATE TABLE session_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transcript_id   UUID NOT NULL REFERENCES transcripts(id) ON DELETE CASCADE,
    session_id      TEXT NOT NULL,
    entry_index     INTEGER NOT NULL,
    provider        TEXT NOT NULL,
    entry_type      TEXT NOT NULL,
    role            TEXT NOT NULL,
    timestamp_ms    BIGINT,
    content_preview TEXT,
    tokens_in       INTEGER,
    tokens_out      INTEGER,
    has_tool_use    BOOLEAN NOT NULL DEFAULT FALSE,
    tool_kind       TEXT,
    tool_names_csv  TEXT,
    has_thinking    BOOLEAN NOT NULL DEFAULT FALSE,
    is_error        BOOLEAN NOT NULL DEFAULT FALSE,
    stop_reason     TEXT,
    raw_byte_length INTEGER,
    tool_call_id    TEXT,
    entry_id        TEXT,
    parent_entry_id TEXT,
    depth           INTEGER NOT NULL DEFAULT 0,
    parent_index    INTEGER,
    tool_input      TEXT,
    tool_output     TEXT,
    extra           JSONB,
    UNIQUE(transcript_id, entry_index)
);

CREATE INDEX idx_session_entries_transcript ON session_entries(transcript_id);
CREATE INDEX idx_session_entries_entry_type ON session_entries(entry_type);
CREATE INDEX idx_session_entries_tool_kind ON session_entries(tool_kind) WHERE tool_kind IS NOT NULL;
