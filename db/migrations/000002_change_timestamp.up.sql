-- name: UpdateTimestampsToTimestamptz :up
-- Update timestamp fields to timestamptz in various tables

-- Update users table
ALTER TABLE users
  ALTER COLUMN createdAt SET DATA TYPE timestamptz,
  ALTER COLUMN updatedAt SET DATA TYPE timestamptz;

-- Update verify_email table
ALTER TABLE verify_email
  ALTER COLUMN created_at SET DATA TYPE timestamptz,
  ALTER COLUMN expired_at SET DATA TYPE timestamptz;

-- Update sessions table
ALTER TABLE sessions
  ALTER COLUMN expires_at SET DATA TYPE timestamptz,
  ALTER COLUMN created_at SET DATA TYPE timestamptz;

-- Update wallet table
ALTER TABLE wallet
  ALTER COLUMN createdAt SET DATA TYPE timestamptz,
  ALTER COLUMN updatedAt SET DATA TYPE timestamptz;

-- Update transaction table
ALTER TABLE transaction
  ALTER COLUMN createdAt SET DATA TYPE timestamptz,
  ALTER COLUMN updatedAt SET DATA TYPE timestamptz;

-- Update redeem table
ALTER TABLE redeem
  ALTER COLUMN createdAt SET DATA TYPE timestamptz,
  ALTER COLUMN updatedAt SET DATA TYPE timestamptz;

-- Update voucher table
ALTER TABLE voucher
  ALTER COLUMN createdAt SET DATA TYPE timestamptz,
  ALTER COLUMN updatedAt SET DATA TYPE timestamptz,
  ALTER COLUMN expireAt SET DATA TYPE timestamptz;

-- End of migration
