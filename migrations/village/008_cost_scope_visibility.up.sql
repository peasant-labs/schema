-- Cost columns: per-session cost breakdown by pricing tier.
ALTER TABLE transcripts ADD COLUMN cost_input_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_output_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_reasoning_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_cache_read_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_cache_write_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_total_usd DOUBLE PRECISION;
ALTER TABLE transcripts ADD COLUMN cost_model_id TEXT;

-- Scope: owner-defined data usage scope (distinct from access visibility).
ALTER TABLE transcripts ADD COLUMN scope TEXT;

-- Update visibility CHECK to canonical set: 'shared' → 'group'.
ALTER TABLE transcripts DROP CONSTRAINT IF EXISTS transcripts_visibility_check;
ALTER TABLE transcripts ADD CONSTRAINT transcripts_visibility_check
    CHECK (visibility IN ('public', 'private', 'group'));
