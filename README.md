# Q2PRO Anticheat Server

Server de anticheat para servidores Q2PRO. Recibe conexiones TCP de servidores de juego, analiza hashes de archivos y valores de cvar para detectar cheats, y provee un dashboard web para revisar violaciones y capturas de pantalla.

## Puertos

| Puerto | Protocolo | Descripcion |
|--------|-----------|-------------|
| 27915  | TCP       | Servidor de juego (conexiones desde Q2PRO servers) |
| 27916  | TCP       | Dashboard web (panel de administracion) |

## Variables de Entorno

| Variable | Default | Descripcion |
|----------|---------|-------------|
| `ADMIN_USER` | `admin` | Usuario del dashboard web |
| `ADMIN_PASS` | `admin123` | Contrasena del dashboard web |
| `TZ` | `UTC` | Zona horaria del contenedor |

## Despliegue con Docker

```bash
# Construir y ejecutar
docker compose up -d

# Ver logs
docker compose logs -f

# Detener
docker compose down
```

## Despliegue con Coolify (Persistencia de Datos)

Para que los datos (base de datos SQLite, capturas de pantalla, bloqueos de blacklist y sesiones) **persistan entre deploys y reinicios**, debes configurar el almacenamiento persistente en Coolify:

### Opción 1: Despliegue con Dockerfile (Recomendado)
1. En Coolify, crea una nueva aplicación conectando tu repositorio de GitHub.
2. Coolify detectará el `Dockerfile` automáticamente.
3. Configura los puertos: `27915:27915` (TCP) y `27916:27916` (TCP / Dashboard).
4. **IMPORTANTE — Configurar Almacenamiento Persistente**:
   - Dirígete a la pestaña **Storages** (Almacenamiento) de la aplicación en Coolify.
   - Haz clic en **Add Storage** / **Add Persistent Volume**.
   - En **Destination Path** (Ruta dentro del contenedor), ingresa exactamente: `/app/data`
   - En **Volume Name / Host Path**, asigna un nombre (ej. `ac-data`) o déjalo por defecto.
   - Guarda los cambios.
5. Configura las variables de entorno en Coolify:
   - `ADMIN_USER` (usuario del dashboard web, ej: `admin`)
   - `ADMIN_PASS` (contraseña inicial del dashboard web)
6. Haz clic en **Deploy**.

> **Nota sobre permisos:** El contenedor incluye un script de inicialización (`entrypoint.sh`) que ajusta automáticamente los permisos del volumen `/app/data` al iniciar, garantizando que el usuario no-root `acuser` pueda leer y escribir la base de datos y capturas sin errores de permisos.

### Opción 2: Despliegue con Docker Compose
1. En Coolify, crea un nuevo recurso de tipo **Docker Compose**.
2. Conecta el repositorio de GitHub. Coolify usará el archivo `docker-compose.yml` que ya define el volumen persistente `ac-data:/app/data`.
3. Configura las variables de entorno (`ADMIN_USER`, `ADMIN_PASS`) y realiza el deploy.

---

## Retención Automática y Mantenimiento de Almacenamiento

El servidor incluye gestión de almacenamiento para evitar que el disco se llene con capturas antiguas:

- **Retención Automática**: Desde el panel web (**Configuración** -> **Retención y Limpieza**), puedes definir cuántos días se conservan las capturas (por defecto 30 días). El servidor ejecuta una tarea diaria en segundo plano que elimina archivos y registros antiguos y ejecuta `VACUUM` en SQLite.
- **Purga Inmediata**: Desde la misma pantalla puedes forzar una limpieza manual seleccionando el período deseado y visualizar el espacio liberado.
- **Métricas en Vivo**: Consulta el tamaño en disco de la base de datos (`anticheat.db`), total de capturas y espacio ocupado en tiempo real.

---

## Build local (sin Docker)

```bash
CGO_ENABLED=0 go build -o anticheat-server .
./anticheat-server --admin-user miusuario --admin-pass mipassword
```

## Archivos de datos

- `data/anticheat.db` - Base de datos SQLite (se crea automáticamente)
- `data/screenshots/` - Capturas de pantalla de jugadores (organizadas por fecha `YYYY-MM-DD/`)