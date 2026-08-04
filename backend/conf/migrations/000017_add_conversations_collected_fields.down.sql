ALTER TABLE conversations DROP COLUMN IF EXISTS collected_fields;
ALTER TABLE conversations DROP COLUMN IF EXISTS triage_confidence;
ALTER TABLE conversations DROP COLUMN IF EXISTS triage_ready_for_handoff;
