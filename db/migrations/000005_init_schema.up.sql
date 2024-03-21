-- Up migration

-- Drop the first foreign key constraint
ALTER TABLE users DROP CONSTRAINT users_username_fkey;

-- Add the second foreign key constraint
ALTER TABLE wallet ADD FOREIGN KEY (owner) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE transaction DROP CONSTRAINT transaction_id_fkey;


ALTER TABLE redeem ADD FOREIGN KEY (transactionId) REFERENCES transaction(id) ON DELETE CASCADE;

