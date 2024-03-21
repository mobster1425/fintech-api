-- name: CreateUser :one
INSERT INTO users (
  username, email, password, role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;


-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;


-- name: UpdateUser :one
UPDATE users
SET
  password = COALESCE(sqlc.narg(password), password),
  email = COALESCE(sqlc.narg(email), email),
    updatedAt = COALESCE(sqlc.narg(updatedAt), updatedAt),
  is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified)
WHERE
  username = sqlc.arg(username)
RETURNING *;



-- name: UpdateUserStatus :one
UPDATE users
SET  status = $2
WHERE username = $1
RETURNING *;