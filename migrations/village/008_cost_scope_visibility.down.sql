-- Revert visibility CHECK to original set.
ALTER TABLE transcripts DROP CONSTRAINT IF EXISTS transcripts_visibility_check;
ALTER TABLE transcripts ADD CONSTRAINT transcripts_visibility_check
    CHECK (visibility IN ('public', 'private', 'shared'));

-- Drop scope column.
ALTER TABLE transcripts DROP COLUMN IF EXISTS scope;

-- Drop cost columns (reverse order of addition).
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_model_id;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_total_usd;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_cache_write_usd;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_cache_read_usd;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_reasoning_usd;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_output_usd;
ALTER TABLE transcripts DROP COLUMN IF EXISTS cost_input_usd;
