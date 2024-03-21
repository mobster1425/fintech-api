
-- name: CreateWallet :one
INSERT INTO wallet (
  owner,
  balance
) VALUES (
  $1, $2
) RETURNING *;


-- name: UpdateWallet :one
UPDATE wallet
SET balance = $2
WHERE id = $1
RETURNING *;



-- name: GetWallet :one
SELECT * FROM wallet
WHERE id = $1 LIMIT 1;


-- name: AddWalletBalance :one
UPDATE wallet
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;



-- name: GetWalletbyOwner :one
SELECT * FROM wallet
WHERE owner = $1 LIMIT 1;