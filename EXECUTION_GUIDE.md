# 🚀 Guía de Ejecución - AgendaFacil

## Requisitos Previos

- Go 1.24.6+
- PostgreSQL 13+
- Docker (opcional, para ejecutar BD en contenedor)

---

## 🐘 Configurar Base de Datos

### Opción 1: PostgreSQL Local

```bash
# Crear base de datos
createdb barberia

# Crear usuario
createuser -P postgres  # contraseña: postgres
```

### Opción 2: Docker Compose

```bash
# Ejecutar desde el directorio del proyecto
docker-compose up -d

# Verificar que está funcionando
docker-compose ps
```

---

## ⚙️ Variables de Entorno

Crear archivo `.env`:

```bash
# Base de Datos
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=barberia

# Servidor
PORT=8080

# JWT (IMPORTANTE: Cambiar en producción)
JWT_SECRET=tu_secreto_super_seguro_aqui_minimo_32_caracteres

# Logging
LOG_LEVEL=info
```

---

## 🏃 Ejecutar el Servidor

### Método 1: Go Directo

```bash
# Descargar dependencias
go mod download

# Ejecutar tests
go test ./...

# Compilar
go build -o ./server ./cmd/server/

# Ejecutar
./server
```

### Método 2: Docker

```bash
# Construir imagen
docker build -t agendafacil:latest .

# Ejecutar con docker-compose
docker-compose up

# Ejecutar solo app (BD por separado)
docker run \
  -e DB_HOST=host.docker.internal \
  -p 8080:8080 \
  agendafacil:latest
```

### Método 3: Desarrollo (hot reload)

```bash
# Instalar air para hot reload
go install github.com/cosmtrek/air@latest

# Ejecutar con hot reload
air
```

---

## 📡 Probar la API

### 1. Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'

# Respuesta:
# {
#   "token": "eyJhbGciOiJIUzI1NiIs...",
#   "rol": "admin"
# }
```

### 2. Obtener Info de Barbería

```bash
curl http://localhost:8080/b/la-barberia

# Retorna: JSON con info de barbería, servicios y barberos
```

### 3. Listar Servicios

```bash
curl http://localhost:8080/b/la-barberia/servicios

# Retorna: Array de servicios activos
```

### 4. Listar Barberos

```bash
curl http://localhost:8080/b/la-barberia/barberos

# Retorna: Array de barberos disponibles
```

### 5. Obtener Disponibilidad

```bash
curl "http://localhost:8080/b/la-barberia/disponibilidad?fecha=2025-01-09&servicio_id=1"

# Retorna: Array de slots disponibles
```

### 6. Crear Reserva

```bash
curl -X POST http://localhost:8080/b/la-barberia/reservar \
  -H "Content-Type: application/json" \
  -d '{
    "servicio_id": 1,
    "barbero_id": 1,
    "fecha": "2025-01-09",
    "hora_inicio": "10:00",
    "cliente_nombre": "Juan Pérez",
    "cliente_telefono": "5551234567"
  }'

# Retorna: 201 Created con datos de la reserva
```

---

## 🧪 Ejecutar Tests

```bash
# Tests unitarios
go test -v ./internal/handlers/...

# Tests con cobertura
go test -cover ./...

# Tests con reporte de cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Tests específicos
go test -run TestCalcularSlots_Basic ./internal/handlers/...
```

---

## 📊 Verificar Salud del Servidor

```bash
# Health check (si está implementado)
curl http://localhost:8080/health

# Verificar logs
tail -f logs/app.log  # Cuando se implemente logging
```

---

## 🔧 Estructura de Directorios Durante Ejecución

```
AgendaFacil/
├── cmd/server/
│   └── main.go              # Punto de entrada
├── internal/handlers/        # Handlers HTTP
├── db/
│   ├── sqlc/                # Queries generadas
│   └── queries/             # SQL files
├── web/                      # Archivos estáticos
├── .env                      # Configuración (gitignore)
├── server                    # Ejecutable compilado
└── docker-compose.yml       # Configuración Docker
```

---

## 🛠️ Troubleshooting

### Error: "could not connect to database"

```bash
# Verificar conexión PostgreSQL
psql -U postgres -d barberia -c "SELECT 1"

# Si usa Docker Compose
docker-compose logs db
docker-compose exec db psql -U postgres -d barberia -c "SELECT 1"
```

### Error: "port 8080 already in use"

```bash
# Cambiar puerto en .env
PORT=8081

# O matar proceso en puerto 8080
lsof -i :8080
kill -9 <PID>
```

### Error: "JWT_SECRET not set"

```bash
# Asegurarse que .env está configurado
export JWT_SECRET="tu_secreto_aqui"

# O ejecutar
source .env
go run ./cmd/server/main.go
```

### Error: "Go version mismatch"

```bash
# Verificar versión actual
go version

# Esperada: go1.24.6 o superior
# Si es diferente, actualizar Go
```

---

## 📈 Monitoreo

### Logs en Consola

```bash
# Ver logs en tiempo real
go run ./cmd/server/main.go

# Con logging color
LOG_LEVEL=debug go run ./cmd/server/main.go
```

### Métricas (cuando se implemente)

```bash
# Prometheus metrics (si se agrega)
curl http://localhost:8080/metrics
```

---

## 🚀 Deployar a Producción

### Checklist Pre-Producción

- [ ] Cambiar JWT_SECRET a algo seguro
- [ ] Configurar CORS si es necesario
- [ ] Habilitar HTTPS
- [ ] Configurar logging remoto
- [ ] Setup de backups de BD
- [ ] Implementar monitoring
- [ ] Tests de carga
- [ ] Security review

### Docker Producción

```bash
# Build multi-stage para tamaño reducido
docker build -f Dockerfile.prod -t agendafacil:prod .

# Ejecutar con recursos limitados
docker run \
  -e DB_PASSWORD=$(cat /run/secrets/db_password) \
  -e JWT_SECRET=$(cat /run/secrets/jwt_secret) \
  --memory="512m" \
  --cpus="1" \
  -p 8080:8080 \
  agendafacil:prod
```

### Kubernetes

```bash
# Ejemplo de deployment
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

---

## 📞 Soporte

Para problemas:

1. Revisar [ANALYSIS_REPORT.md](./ANALYSIS_REPORT.md)
2. Consultar [IMPROVEMENTS.md](./IMPROVEMENTS.md)
3. Ejecutar tests para diagnosticar
4. Revisar logs de BD y aplicación

---

## ✅ Verificación Inicial

Después de ejecutar, deberías ver:

```
2025-01-08T10:30:45Z Conectado a la DB correctamente
2025-01-08T10:30:45Z Servidor en puerto 8080
```

Y poder acceder:
- Servidor: `http://localhost:8080`
- API: `http://localhost:8080/b/la-barberia`

---

*Guía de ejecución - AgendaFacil 2025*
