-- name: RevertUpdateTimestampsToTimestamptz :down
-- Revert changes to timestamp fields in various tables

-- Revert users table
ALTER TABLE users
  ALTER COLUMN createdAt SET DATA TYPE timestamp,
  ALTER COLUMN updatedAt SET DATA TYPE timestamp;

-- Revert verify_email table
ALTER TABLE verify_email
  ALTER COLUMN created_at SET DATA TYPE timestamp,
  ALTER COLUMN expired_at SET DATA TYPE timestamp;

-- Revert sessions table
ALTER TABLE sessions
  ALTER COLUMN expires_at SET DATA TYPE timestamp,
  ALTER COLUMN created_at SET DATA TYPE timestamp;

-- Revert wallet table
ALTER TABLE wallet
  ALTER COLUMN createdAt SET DATA TYPE timestamp,
  ALTER COLUMN updatedAt SET DATA TYPE timestamp;

-- Revert transaction table
ALTER TABLE transaction
  ALTER COLUMN createdAt SET DATA TYPE timestamp,
  ALTER COLUMN updatedAt SET DATA TYPE timestamp;

-- Revert redeem table
ALTER TABLE redeem
  ALTER COLUMN createdAt SET DATA TYPE timestamp,
  ALTER COLUMN updatedAt SET DATA TYPE timestamp;

-- Revert voucher table
ALTER TABLE voucher
  ALTER COLUMN createdAt SET DATA TYPE timestamp,
  ALTER COLUMN updatedAt SET DATA TYPE timestamp,
  ALTER COLUMN expireAt SET DATA TYPE timestamp;

-- End of migration
