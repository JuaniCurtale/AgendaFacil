-- name: CreateUsuario :one
INSERT INTO usuarios (
  barberia_id, username, nombre, apellido, email, password_hash, rol
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUsuarioByEmail :one
SELECT *
FROM usuarios
WHERE email = $1
  AND activo = true;

-- name: ListBarberos :many
SELECT id, nombre, apellido
FROM usuarios
WHERE barberia_id = $1
  AND rol = 'barbero'
  AND activo = true
ORDER BY nombre;

-- name: GetUsuarioByUsername :one
SELECT *
FROM usuarios
WHERE username = $1
  AND activo = true;