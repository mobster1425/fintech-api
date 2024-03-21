
-- name: CreateTransaction :one
INSERT INTO transaction (
  sender_wallet_id,
  receiver_wallet_id,
  amount,
charge,
type,
sendamount,
receiveamount,
note,
status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,$9
) RETURNING *;




-- name: GetTransaction :one
SELECT * FROM transaction
WHERE id = $1 LIMIT 1;



-- name: ListTransactions :many
SELECT * FROM transaction
WHERE 
    sender_wallet_id = $1 OR
    receiver_wallet_id = $2
ORDER BY id
LIMIT $3
OFFSET $4;




-- name: UpdateTransactionStatus :one
UPDATE transaction
SET status = $2
WHERE id = $1
RETURNING *;



-- name: UpdateTransactionReceiverWallet :one
UPDATE transaction
SET receiver_wallet_id = $2
WHERE id = $1
RETURNING *;

