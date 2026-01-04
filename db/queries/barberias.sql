-- name: CreateBarberia :one
INSERT INTO barberias (nombre, slug, hora_apertura, hora_cierre)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBarberiaBySlug :one
SELECT *
FROM barberias
WHERE slug = $1
  AND activa = true;

-- name: ListBarberosByBarberia :many
SELECT id, nombre
FROM usuarios
WHERE barberia_id = $1
  AND rol = 'barbero'
  AND activo = true
ORDER BY nombre;
