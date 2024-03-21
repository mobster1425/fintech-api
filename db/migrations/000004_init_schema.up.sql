-- Add the new column 'UsedBy' to the 'vouchers' table
ALTER TABLE voucher
ADD COLUMN UsedBy VARCHAR[];