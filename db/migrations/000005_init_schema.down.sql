-- Down migration

-- Drop the second foreign key constraint
ALTER TABLE wallet DROP CONSTRAINT wallet_owner_fkey;

-- Add back the first foreign key constraint
ALTER TABLE users ADD FOREIGN KEY (username) REFERENCES wallet(owner) ON DELETE CASCADE;


-- Revert the foreign key constraint on the "redeem" table
ALTER TABLE redeem DROP CONSTRAINT redeem_transactionId_fkey;

-- Recreate the foreign key constraint on the "transaction" table
ALTER TABLE transaction ADD CONSTRAINT transaction_id_fkey FOREIGN KEY (id) REFERENCES redeem(transactionId) ON DELETE CASCADE;