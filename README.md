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

## Despliegue con Coolify

1. Crear repositorio en GitHub con este codigo
2. En Coolify, crear una nueva applicacion tipo "Dockerfile"
3. Conectar el repositorio de GitHub
4. Configurar los puertos: `27915` (TCP) y `27916` (TCP)
5. Configurar las variables de entorno en Coolify:
   - `ADMIN_USER` (usuario del dashboard)
   - `ADMIN_PASS` (contrasena del dashboard - cambiar de valor por defecto)
6. Coolify detectara el Dockerfile automaticamente

## Build local (sin Docker)

```bash
CGO_ENABLED=0 go build -o anticheat-server .
./anticheat-server --admin-user miusuario --admin-pass mipassword
```

## Archivos de datos

- `data/anticheat.db` - Base de datos SQLite (se crea automaticamente)
- `data/screenshots/` - Capturas de pantalla de jugadores (organizadas por fecha)