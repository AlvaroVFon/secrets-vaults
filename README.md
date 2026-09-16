# Secrets Vault

Vault de secretos de configuración pensado para un **home-server**. Almacena la
configuración sensible de los distintos proyectos alojados en un servidor local
y la sirve a cada aplicación según su `apikey`, de forma similar a
_AWS Secrets Manager_.

Incluye:

- **API** en Go (solo `stdlib` + `pgx`).
- **Panel de administración** web (Vue 3 + TypeScript + Vite) para gestionar
  consumers y secretos.
- **PostgreSQL** como almacenamiento, con migraciones versionadas
  (`golang-migrate`).
- **Generador de configuración** tipada (`TypeScript` / `Go`) a partir de los
  secretos de un consumer.

---

## Cómo funciona

```
                    ┌──────────────────────────────┐
   Navegador ──────▶│  web (nginx :80)             │
   (panel admin)    │  sirve el SPA + proxy /api   │
                    └──────────────┬───────────────┘
                                   │ /auth /management
                                   ▼
  Aplicaciones  ──x-apikey──▶  ┌─────────────┐      ┌──────────────┐
  consumidoras  ────────────▶  │  api :8080  │─────▶│ postgres:5432│
                               └─────────────┘      └──────────────┘
```

- **Panel (`web`)** → login con usuario/contraseña y gestión de consumers,
  secretos y roles. nginx hace de reverse proxy hacia `api`.
- **Aplicaciones consumidoras** → obtienen sus secretos con la cabecera
  `x-apikey`, sin usuario ni sesión.

---

## Estructura del repositorio

```
cmd/api/                 # Punto de entrada de la API (main.go, seed.go)
internal/
  auth/                  # Login, tokens HMAC y middleware Bearer
  config/                # Carga de configuración desde variables de entorno
  configgen/             # Generación de interfaces/structs tipadas
  consumers/             # Entidad + repositorio + servicio de consumers
  secrets/               # Entidad + repositorio + servicio + API de secretos
  management/            # Endpoints del panel de administración
  roles/ users/          # Roles y usuarios del panel
  database/              # Pool de conexiones PostgreSQL (pgx v5)
  httpx/                 # Helper de respuestas JSON
migrations/              # Migraciones SQL (golang-migrate)
web/                     # Panel de administración (Vue 3 + Vite)
docker-compose.yml       # Stack completo para el home-server
Makefile                 # Atajos de desarrollo y operación
```

---

## Despliegue en un home-server (Docker Compose)

### Requisitos

- Docker Engine 24+ con el plugin `docker compose` (`docker compose version`).
- Puertos libres: `8081` (panel) y, opcionalmente, `8080` (API) y `5432`
  (PostgreSQL).
- Acceso a internet solo la primera vez (build de las imágenes).

### 1. Clonar el repositorio

```bash
git clone <url-del-repositorio> secrets-vault
cd secrets-vault
```

### 2. Crear el fichero `.env`

El `Dockerfile` de la API copia un fichero `.env` a la imagen, pero `.env` está
en `.gitignore`, así que **hay que crearlo antes de construir**:

```bash
cp .env.example .env
```

> Los valores efectivos del stack de Docker salen de `docker-compose.yml`
> (`environment:`). El fichero `.env` solo es necesario para que el build de la
> imagen de la API funcione.

### 3. Configurar credenciales (importante)

Edita `docker-compose.yml` y cambia los valores por defecto antes de exponer el
servicio:

| Variable           | Servicio | Descripción                                  | Valor por defecto      |
| ------------------ | -------- | -------------------------------------------- | ---------------------- |
| `AUTH_SECRET`      | `api`    | Clave HMAC para firmar los tokens de sesión  | `change-me-in-production` |
| `AUTH_TTL`         | `api`    | Caducidad del token, en horas                | `24`                   |
| `ADMIN_USERNAME`   | `api`    | Usuario inicial del panel (se crea si no hay usuarios) | `admin`    |
| `ADMIN_PASSWORD`   | `api`    | Contraseña inicial del panel                 | `admin`                |
| `POSTGRES_PASSWORD`| `api`, `postgres` | Contraseña de PostgreSQL            | `postgres`             |
| `POSTGRES_URL`     | `api`    | DSN de conexión a PostgreSQL                 | `postgres://postgres:postgres@postgres:5432/app?sslmode=disable` |

Genera un `AUTH_SECRET` distinto por instalación, por ejemplo:

```bash
openssl rand -base64 32
```

> `ADMIN_USERNAME` / `ADMIN_PASSWORD` solo se aplican la **primera vez**, cuando
> la tabla de usuarios está vacía. Si ya arrancaste con los valores por defecto,
> cámbialos antes del primer `up` o borra el volumen de PostgreSQL para
> reinicializar.

### 4. Levantar el stack

```bash
docker compose up -d --build
```

Esto arranca, en orden:

1. `postgres` (con healthcheck).
2. `migrate` — aplica todas las migraciones y termina.
3. `api` — espera a que las migraciones hayan terminado.
4. `web` — panel nginx que proxya `/auth`, `/secrets` y `/management` a `api`.

O con el atajo:

```bash
make up
```

### 5. Acceder

- Panel de administración: **http://<ip-del-servidor>:8081**
- API directa (opcional): **http://<ip-del-servidor>:8080**

Las migraciones crean un consumer `superadmin` de ejemplo con la apikey
`superadmin-api-key` (`migrations/000005_seed_roles_and_superadmin.up.sql`).
Cámbiala o desactívala en producción.

---

## Panel de administración

1. **Login** en `/` con `ADMIN_USERNAME` / `ADMIN_PASSWORD`.
2. **Consumers**: crea una aplicación consumidora con un nombre, una `apikey` y
   un rol (`superadmin` o `admin`).
3. **Secretos**: añade pares clave/valor al consumer. Los valores se guardan
   como *config* (visibles) por defecto; marca `isSecret` para ocultar el valor
   en la interfaz.
4. **Generar interfaz**: desde un consumer puedes exportar una definición de
   configuración tipada en **TypeScript** (`<consumer>.config.ts`) o **Go**
   (`<consumer>_config.go`). Las claves con notación de punto (`db.host`) se
   anidan y el tipo se infiere del valor.

---

## API

Todas las respuestas tienen la forma:

```json
{ "status": 200, "message": "...", "data": {} }
```

### Autenticación del panel

```http
POST /auth/login
Content-Type: application/json

{ "username": "admin", "password": "..." }
```

Devuelve un token que debe enviarse como `Authorization: Bearer <token>` en el
resto de endpoints de `/management`.

| Método | Ruta | Descripción |
| ------ | ---- | ----------- |
| POST | `/auth/login` | Inicia sesión y devuelve el token |

### Gestión (requiere `Authorization: Bearer`)

| Método | Ruta | Descripción |
| ------ | ---- | ----------- |
| GET | `/management/consumers` | Lista consumers |
| POST | `/management/consumers` | Crea un consumer |
| PUT | `/management/consumers/{id}` | Actualiza un consumer |
| DELETE | `/management/consumers/{id}` | Borra un consumer (falla si tiene secretos) |
| GET | `/management/consumers/{id}/config?lang=ts\|go` | Genera la config del consumer |
| GET | `/management/secrets` | Lista secretos (agrupados por consumer) |
| GET | `/management/secrets?consumerId={id}` | Secretos de un consumer |
| POST | `/management/secrets` | Crea un secreto |
| PUT | `/management/secrets/{id}` | Actualiza un secreto |
| DELETE | `/management/secrets/{id}` | Borra un secreto |
| GET | `/management/roles` | Lista roles |

### Uso por aplicaciones consumidoras (requiere `x-apikey`)

| Método | Ruta | Descripción |
| ------ | ---- | ----------- |
| GET | `/secrets/consumer` | Devuelve los secretos del consumer de la apikey |
| POST | `/secrets` | Registra un secreto para el consumer de la apikey |

```bash
curl -H "x-apikey: superadmin-api-key" http://localhost:8080/secrets/consumer
```

---

## Exponerlo de forma segura en el home-server

El stack publica los puertos en todas las interfaces. Para un servidor
doméstico se recomienda:

- **Cerrar los puertos internos.** Deja accesible solo `web`; para la API y
  PostgreSQL limita el bind a loopback:

  ```yaml
  api:
      ports:
        - "127.0.0.1:8080:8080"
  postgres:
      ports:
        - "127.0.0.1:5432:5432"
  ```

- **Poner un reverse proxy con TLS** delante de `web` (Nginx Proxy Manager,
  Caddy, Traefik…) apuntando a `http://<host>:8081`, y usar un dominio o
  certificado local.
- **Cambiar** `AUTH_SECRET`, `ADMIN_PASSWORD` y `POSTGRES_PASSWORD`.
- **No exponer PostgreSQL** a internet: contiene los secretos en claro.

---

## Operación

```bash
docker compose logs -f api      # logs de la API
docker compose ps               # estado de los servicios
docker compose down             # detiene el stack (conserva datos)
docker compose restart api      # reinicia un servicio
```

**Actualizar el código** (sin perder la base de datos):

```bash
git pull
docker compose up -d --build
```

Las migraciones se aplican solas al arrancar (el servicio `migrate` corre antes
que `api`).

### Backup y restauración

Los datos viven en el volumen `postgres_data`.

```bash
docker compose exec -T postgres pg_dump -U postgres app > backup.sql   # backup
cat backup.sql | docker compose exec -T postgres psql -U postgres app  # restore
```

---

## Desarrollo local

### Backend (Go 1.26+)

```bash
cp .env.example .env            # apunta POSTGRES_URL a localhost:5432
make migrate-up                 # aplica migraciones a una BD local
go run ./cmd/api                # API en :8080
```

### Frontend (Node 24+)

```bash
make web-install                # npm install en web/
make web-dev                    # Vite en :5173, proxya a :8080
```

### Tests

Los tests de integración levantan una PostgreSQL dedicada en el puerto `4321`
usando `.env.test`:

```bash
make test                       # levanta la BD de test y ejecuta `go test -p 1 ./...`
make docker-test-down           # apaga la BD de test
```

---

## Comandos del Makefile

```bash
make help                       # lista todos los comandos disponibles
```

| Comando | Descripción |
| ------- | ----------- |
| `make up` / `make down` | Levanta / detiene el stack completo |
| `make migrate-up` | Aplica migraciones pendientes |
| `make migrate-down` | Revierte la última migración |
| `make migrate-reset` | Revierte todo y reaplica |
| `make migrate-create NAME=...` | Crea una nueva migración |
| `make test` | Tests de integración |
| `make web-dev` / `make web-build` | Front en desarrollo / build de producción |

---

## Decisiones técnicas

- **BDD**: PostgreSQL 17 con `pgx v5`.
- **Migraciones**: `golang-migrate` (imagen oficial en el stack).
- **API**: solo `stdlib` (`net/http`, `ServeMux` con patrones por método).
- **Auth del panel**: tokens HMAC-SHA256 propios, sin estado.
- **Auth de consumidores**: `apikey` por cabecera `x-apikey`.
