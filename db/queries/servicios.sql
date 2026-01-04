-- name: CreateServicio :one
INSERT INTO servicios (
  barberia_id, nombre, duracion_minutos, precio
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListServicios :many
SELECT *
FROM servicios
WHERE barberia_id = $1
  AND activo = true
ORDER BY nombre;

-- name: DeactivateServicio :exec
UPDATE servicios
SET activo = false
WHERE id = $1
  AND barberia_id = $2;

-- name: ListServiciosByBarberia :many
SELECT id, nombre, duracion_minutos, precio
FROM servicios
WHERE barberia_id = $1
  AND activo = true
ORDER BY nombre;
