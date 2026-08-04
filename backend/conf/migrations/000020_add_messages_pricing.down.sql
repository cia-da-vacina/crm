DROP INDEX IF EXISTS messages_pricing_category_idx;
ALTER TABLE messages DROP COLUMN IF EXISTS pricing_category;
ALTER TABLE messages DROP COLUMN IF EXISTS pricing_billable;
ALTER TABLE messages DROP COLUMN IF EXISTS pricing_model;
ALTER TABLE messages DROP COLUMN IF EXISTS pricing_confirmed;
ALTER TABLE messages DROP COLUMN IF EXISTS cost_brl;
