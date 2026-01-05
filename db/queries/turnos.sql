-- name: CreateTurno :one
INSERT INTO turnos (
  barberia_id,
  barbero_id,
  servicio_id,
  fecha,
  hora_inicio,
  hora_fin,
  cliente_nombre,
  cliente_telefono,
  estado
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListTurnosByFecha :many
SELECT t.*, s.nombre AS servicio_nombre, u.nombre AS barbero_nombre
FROM turnos t
JOIN servicios s ON s.id = t.servicio_id
JOIN usuarios u ON u.id = t.barbero_id
WHERE t.barberia_id = $1
  AND t.fecha = $2
  AND t.estado != 'cancelado'
ORDER BY t.hora_inicio;

-- name: ListTurnosByFechaAndBarbero :many
SELECT t.*, s.nombre AS servicio_nombre
FROM turnos t
JOIN servicios s ON s.id = t.servicio_id
WHERE t.barberia_id = $1
  AND t.fecha = $2
  AND t.barbero_id = $3
  AND t.estado != 'cancelado'
ORDER BY t.hora_inicio;

-- name: HasTurnoOverlap :one
SELECT EXISTS (
  SELECT 1
  FROM turnos
  WHERE barberia_id = $1
    AND barbero_id = $2
    AND fecha = $3
    AND estado != 'cancelado'
    AND (
      -- Lógica correcta de intersección:
      -- (Inicio_Existente < Fin_Nuevo) Y (Fin_Existente > Inicio_Nuevo)
      hora_inicio < sqlc.arg('hora_fin') 
      AND hora_fin > sqlc.arg('hora_inicio')
    )
);

-- name: CancelTurno :exec
UPDATE turnos
SET estado = 'cancelado'
WHERE id = $1
  AND barberia_id = $2;


-- name: ListTurnosOcupados :many
SELECT barbero_id, hora_inicio, hora_fin
FROM turnos
WHERE barberia_id = $1
  AND fecha = $2
  AND estado != 'cancelado'
ORDER BY hora_inicio;
