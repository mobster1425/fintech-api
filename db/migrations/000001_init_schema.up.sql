CREATE TYPE "user_status" AS ENUM (
  'inactive',
  'active',
  'banned'
);

CREATE TYPE "user_role" AS ENUM (
  'customer',
  'merchant'
);

CREATE TYPE "transaction_status" AS ENUM (
  'INIT',
  'PROCESSING',
  'PENDING',
  'SUCCESS',
  'CANCELED',
  'REJECTED'
);

CREATE TYPE "transaction_type" AS ENUM (
  'TRANSFER',
  'REQUEST',
  'REDEEM',
  'PAYMENT',
  'PAYMENT_VOUCHER',
  'WITHDRAW',
  'DEPOSIT'
);

CREATE TYPE "voucher_status" AS ENUM (
  'AVAILABLE',
  'UNAVAILABLE'
);

CREATE TYPE "voucher_type" AS ENUM (
  'FIXED',
  'PERCENT'
);


CREATE TABLE users (
  createdAt timestamp NOT NULL DEFAULT now(),
  updatedAt timestamp NOT NULL DEFAULT now(),
  username varchar PRIMARY KEY,
  email varchar UNIQUE NOT NULL,
  password varchar NOT NULL,
  role user_role,
  status user_status,
  is_email_verified bool NOT NULL DEFAULT false
);

CREATE TABLE verify_email (
  id bigserial PRIMARY KEY,
  username varchar NOT NULL,
  email varchar NOT NULL,
  secret_code varchar NOT NULL,
  is_used bool NOT NULL DEFAULT false,
  created_at timestamp NOT NULL DEFAULT now(),
  expired_at timestamp NOT NULL DEFAULT now() + interval '15 minutes'
);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  username varchar NOT NULL,
  refresh_token varchar NOT NULL,
  user_agent varchar NOT NULL,
  client_ip varchar NOT NULL,
  is_blocked bool NOT NULL DEFAULT false,
  expires_at timestamp NOT NULL,
  created_at timestamp NOT NULL DEFAULT now()
);

CREATE TABLE wallet (
  id bigserial PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT now(),
  updatedAt timestamp NOT NULL DEFAULT now(),
  owner varchar UNIQUE NOT NULL, -- Add UNIQUE constraint
  balance bigint NOT NULL
);

CREATE TABLE transaction (
  id bigserial PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT now(),
  updatedAt timestamp NOT NULL DEFAULT now(),
  sender_wallet_id bigint NOT NULL,
  receiver_wallet_id bigint,
  charge bigint,
  amount bigint,
  sendAmount bigint,
  receiveAmount bigint,
  note varchar,
  type transaction_type,
  status transaction_status
);

CREATE TABLE redeem (
  id bigserial PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT now(),
  updatedAt timestamp NOT NULL DEFAULT now(),
  code varchar NOT NULL,
  transactionId bigserial UNIQUE NOT NULL
);

CREATE TABLE voucher (
  id bigserial PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT now(),
  updatedAt timestamp NOT NULL DEFAULT now(),
  creator_username varchar,
  value bigint NOT NULL,
  type voucher_type NOT NULL,
  applyFor_username varchar,
  maxUsage integer NOT NULL,
  maxUsageByAccount integer NOT NULL,
  status voucher_status NOT NULL DEFAULT 'AVAILABLE',
  expireAt timestamp NOT NULL,
  code varchar NOT NULL
);

ALTER TABLE users ADD FOREIGN KEY (username) REFERENCES wallet(owner) ON DELETE CASCADE;

ALTER TABLE transaction ADD FOREIGN KEY (sender_wallet_id) REFERENCES wallet(id) ON DELETE CASCADE;

ALTER TABLE transaction ADD FOREIGN KEY (receiver_wallet_id) REFERENCES wallet(id) ON DELETE CASCADE;

ALTER TABLE transaction ADD FOREIGN KEY (id) REFERENCES redeem(transactionId) ON DELETE CASCADE;

ALTER TABLE voucher ADD FOREIGN KEY (creator_username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE voucher ADD FOREIGN KEY (applyFor_username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE verify_email ADD FOREIGN KEY (username) REFERENCES users(username);

ALTER TABLE sessions ADD FOREIGN KEY (username) REFERENCES users(username);
