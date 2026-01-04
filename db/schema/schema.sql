
CREATE TABLE barberias (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    hora_apertura TIME NOT NULL,
    hora_cierre TIME NOT NULL,
    activa BOOLEAN DEFAULT true
);

CREATE TABLE usuarios (
    id SERIAL PRIMARY KEY,
    barberia_id INT NOT NULL,
    nombre VARCHAR(50) NOT NULL,
    apellido VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    rol VARCHAR(20) NOT NULL, -- admin | barbero
    activo BOOLEAN DEFAULT true,

    FOREIGN KEY (barberia_id) REFERENCES barberias(id)
);


CREATE TABLE servicios (
    id SERIAL PRIMARY KEY,
    barberia_id INT NOT NULL,
    nombre VARCHAR(100) NOT NULL,
    duracion_minutos INT NOT NULL,
    precio DECIMAL(10,2) NOT NULL,
    activo BOOLEAN DEFAULT true,

    FOREIGN KEY (barberia_id) REFERENCES barberias(id)
);

CREATE TABLE turnos (
    id SERIAL PRIMARY KEY,
    barberia_id INT NOT NULL,
    barbero_id INT NOT NULL,
    servicio_id INT NOT NULL,

    fecha DATE NOT NULL,
    hora_inicio TIME NOT NULL,
    hora_fin TIME NOT NULL,

    cliente_nombre VARCHAR(100) NOT NULL,
    cliente_telefono VARCHAR(20),

    estado VARCHAR(20) DEFAULT 'pendiente', -- pendiente | confirmado | cancelado
    creado_en TIMESTAMP DEFAULT now(),

    FOREIGN KEY (barberia_id) REFERENCES barberias(id),
    FOREIGN KEY (barbero_id) REFERENCES usuarios(id),
    FOREIGN KEY (servicio_id) REFERENCES servicios(id)
);

INSERT INTO barberias (nombre, slug, hora_apertura, hora_cierre)
VALUES ('Barbería Test', 'test', '09:00', '18:00');

INSERT INTO servicios (barberia_id, nombre, duracion_minutos, precio)
VALUES (1, 'Corte de Cabello', 30, 15.00),
       (1, 'Afeitado', 20, 10.00),
       (1, 'Corte y Afeitado', 45, 22.00);

INSERT INTO usuarios (barberia_id, nombre, apellido, email, password_hash, rol)
VALUES (1, 'Juan', 'Pérez', 'juan@correo.com', 'hashdemo', 'barbero');

INSERT INTO turnos (
  barberia_id,
  barbero_id,
  servicio_id,
  fecha,
  hora_inicio,
  hora_fin,
  cliente_nombre
) VALUES (
  1,
  1,
  1,
  '2026-01-05',
  '10:00',
  '10:30',
  'Juan'
);
