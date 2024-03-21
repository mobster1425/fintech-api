
-- name: CreateVoucher :one
INSERT INTO voucher (
    value,
   applyFor_username,
type,
    maxUsage,
maxUsageByAccount,
status,
expireAt,
code,
creator_username
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;



-- name: GetVoucher :one
SELECT * FROM voucher
WHERE id = $1 LIMIT 1;


-- name: GetVoucherWithCode :one
SELECT * FROM voucher
WHERE creator_username = $1 AND code = $2 
LIMIT 1;

-- name: ListVouchers :many
SELECT * FROM voucher
WHERE 
   creator_username = $1 
   
ORDER BY id;



-- name: UpdateVoucherStatus :one
UPDATE voucher
SET status = $2
WHERE id = $1
RETURNING *;




-- name: UpdateVoucherUsedBy :exec
UPDATE voucher
SET usedby = COALESCE(UsedBy, ARRAY[]::VARCHAR[]) || ($2)::VARCHAR[]
WHERE id = $1;





-- name: DeleteVoucher :exec
DELETE FROM voucher
WHERE id = $1;
