
-- Remove the 'UsedBy' column from the 'vouchers' table
ALTER TABLE vouchers
DROP COLUMN IF EXISTS UsedBy;