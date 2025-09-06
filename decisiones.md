# Decisiones del Proyecto – TP 02: Introducción a Docker

## Tarea 1: Elegir y preparar la aplicación
- Aplicación: API de cursos en Go (backend) con frontend incluido.
- Repositorio de entrega: **pmanavella/TP2-Docker-IS3** (nuevo), para no afectar el repo original.
- Entorno: Docker Desktop + VS Code.
- Identidad Docker Hub: `pmanavella`.

## Tarea 2: Construir una imagen personalizada
- **Dockerfile (multi-stage)**:
  - Build: `golang:1.23` (requerido por `go.mod`), cacheando dependencias con `go.mod`/`go.sum`.
  - Runtime: `alpine:3.20` para imagen final liviana.
  - Binario: `CGO_ENABLED=0` → binario estático Linux/amd64.
  - Puerto expuesto: **8080**.
- Motivos:
  - Multi-stage reduce tamaño y separa build/runtime.
  - `golang:1.23` evita errores de toolchain y alinea con el proyecto.
  - Alpine minimiza superficie y tiempo de despliegue.

## Tarea 3: Publicar la imagen en Docker Hub
- Usuario: `pmanavella`.
- Tags publicadas:
  - `pmanavella/courses-api:dev` (desarrollo).
  - `pmanavella/courses-api:v1.0` (estable inicial).
  - `pmanavella/courses-api:v1.1` (incluye `GET /health`).
- Convención:
  - `dev` para iteración.
  - `vX.Y` como versiones estables inmutables.

## Tarea 4: Integrar una base de datos en contenedor
- Base elegida: **MySQL 8** (imagen oficial).
- Persistencia: volumen **`mysql_data`** montado en `/var/lib/mysql`.
- Inicialización: scripts existentes en `./database` montados en el contenedor.
- Conexión desde la app (DNS interno de Compose):
  - `DB_HOST=database`, `DB_PORT=3306`, `DB_USER=app`, `DB_PASS=app`, `DB_NAME=arqsoft1`.
  - `DB_DSN=app:app@tcp(database:3306)/arqsoft1?parseTime=true`.
- Verificación de persistencia:
  - Crear tabla `tp_check` + insertar fila.
  - Reiniciar MySQL y confirmar que el dato persiste.

## Tarea 5: Configurar QA y PROD con la misma imagen
- Misma imagen para ambos: **`pmanavella/courses-api:v1.1`**.
- Variación por entorno vía archivos:
  - `.env.qa` → `APP_ENV=qa`, DSN a `database`.
  - `.env.prod` → `APP_ENV=prod`, DSN a `database`.
- Puertos (host → contenedor 8080):
  - QA: `8081:8080`
  - PROD: `8082:8080`
- Verificaciones:
  - Variables cargadas dentro de cada contenedor.
  - `/health` responde 200 en QA y PROD.
  - Inspección de imagen muestra **la misma tag** en ambos servicios.

## Tarea 6: Entorno reproducible con docker-compose
- Servicios definidos:
  - `database` (MySQL 8) con healthcheck y volumen `mysql_data`.
  - `api-qa` y `api-prod` usando **la misma imagen**.
  - `frontend` en `3000`.
- Criterios de reproducibilidad:
  - Único requisito: Docker (no se expone `3306` en el host).
  - Healthcheck garantiza que las APIs arranquen cuando la BD está lista.
  - Volúmenes aseguran persistencia entre reinicios.
  - Comando único: `docker compose up -d --pull always --build`.

## Tarea 7: Crear una versión etiquetada
- `v1.0`: primera estable.
- `v1.1`: agrega endpoint `GET /health` y se usa en QA/PROD.
- Política:
  - No sobrescribir tags estables (inmutables).
  - Nuevas features → nueva tag; Compose apunta a la estable.

## Evidencia de funcionamiento
- Misma imagen en QA/PROD:
  - `docker inspect $(docker compose ps -q api-qa) --format '{{.Config.Image}}'`
  - `docker inspect $(docker compose ps -q api-prod) --format '{{.Config.Image}}'`
- Salud:
  - `curl -i http://localhost:8081/health` (QA → 200)
  - `curl -i http://localhost:8082/health` (PROD → 200)
- Persistencia:
  - `SELECT * FROM tp_check;` devuelve el dato tras reiniciar MySQL.
- Volumen:
  - `docker volume ls` muestra `mysql_data`.

## Problemas encontrados y soluciones
- Puerto 8080 ocupado en host (Apache) → usar 8081/8082 para QA/PROD y 8090/8085 en pruebas locales.
- Error 401 al pull de imágenes base → `docker login` + verificación de email Docker Hub.
- Toolchain Go insuficiente → actualizar builder a `golang:1.23`.
- DNS al ejecutar `docker run` aislado → unir a la red de Compose (`--network <proyecto>_default`) o probar directamente con Compose.
- `/health` faltante en `v1.0` → cortar **`v1.1`** y apuntar QA/PROD a esa imagen.

## Cómo reproducir
```bash
git clone git@github.com:pmanavella/TP2-Docker-IS3.git
cd TP2-Docker-IS3
docker compose up -d --pull always --build
curl -i http://localhost:8081/health
curl -i http://localhost:8082/health
docker inspect $(docker compose ps -q api-qa)   --format '{{.Config.Image}}'
docker inspect $(docker compose ps -q api-prod) --format '{{.Config.Image}}'
docker compose exec database mysql -uapp -papp -e "SELECT * FROM tp_check;" arqsoft1 || true
