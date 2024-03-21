
-- name: CreateRedeem :one
INSERT INTO redeem (
  code,
transactionId
) VALUES (
  $1, $2
) RETURNING *;


-- name: GetRedeem :one
SELECT * FROM redeem
WHERE id = $1 LIMIT 1;


-- name: DeleteRedeem :exec
DELETE FROM redeem
WHERE id = $1;



-- name: GetRedeemWithCode :one
SELECT * FROM redeem
WHERE code = $1 LIMIT 1;