-- +goose Up
-- Fix XID wraparound issues by using MAX_VALUE instead of 0 for expired_tx_id

-- Update existing records to use MAX_VALUE for active records
UPDATE relation_tuples 
SET expired_tx_id = '9223372036854775807'::xid8 
WHERE expired_tx_id = '0'::xid8;

UPDATE attributes 
SET expired_tx_id = '9223372036854775807'::xid8 
WHERE expired_tx_id = '0'::xid8;

-- Update default values for future records
ALTER TABLE relation_tuples 
ALTER COLUMN expired_tx_id SET DEFAULT '9223372036854775807'::xid8;

ALTER TABLE attributes 
ALTER COLUMN expired_tx_id SET DEFAULT '9223372036854775807'::xid8;

-- +goose Down
-- Revert changes back to using 0

-- Revert default values
ALTER TABLE relation_tuples 
ALTER COLUMN expired_tx_id SET DEFAULT '0'::xid8;

ALTER TABLE attributes 
ALTER COLUMN expired_tx_id SET DEFAULT '0'::xid8;

-- Revert existing records (only active ones)
UPDATE relation_tuples 
SET expired_tx_id = '0'::xid8 
WHERE expired_tx_id = '9223372036854775807'::xid8;

UPDATE attributes 
SET expired_tx_id = '0'::xid8 
WHERE expired_tx_id = '9223372036854775807'::xid8;