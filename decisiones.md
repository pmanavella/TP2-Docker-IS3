TP 2 - COMANDOS 

──────────────────────────────────────────────────────────────────────────────
TAREA 1: Elegir y preparar la aplicación
──────────────────────────────────────────────────────────────────────────────
# Abrir Docker Desktop y Visual Studio Code.
# Clonar el repositorio con el que se trabajará (ya realizado en tu caso).
# En la terminal integrada de VS Code, definir el usuario de Docker Hub:
export DOCKERUSER="pmanavella"

──────────────────────────────────────────────────────────────────────────────
TAREA 2: Construir una imagen personalizada (backend/Dockerfile multi-stage)
──────────────────────────────────────────────────────────────────────────────
# Construir imagen local (desde la raíz del proyecto: FINALarqsot_I)
docker build -t courses-api:dev ./backend

# Si aparece 401 Unauthorized al bajar imágenes base:
docker logout
docker login
docker pull golang:1.21
docker pull alpine:3.20
docker build -t courses-api:dev ./backend

# Si el go.mod requiere Go >= 1.23, usar builder 1.23 y reconstruir:
docker pull golang:1.23
docker build -t courses-api:dev ./backend

# Probar la imagen local (si 8080 está libre)
docker run --rm --name courses-dev -p 8080:8080 courses-api:dev
# En otra terminal:
curl -i http://localhost:8080/health
# (Ctrl+C para detener)

# Si el puerto 8080 está ocupado en el host, usar otro (ej. 8085):
docker run --rm --name courses-dev -p 8085:8080 courses-api:dev
curl -i http://localhost:8085/health

# Nota para decisiones.md:
# - golang:1.23 (lo exige go.mod) + alpine (runtime liviano)
# - cache de dependencias (go.mod/go.sum), binario estático CGO_ENABLED=0
# - contenedor expone 8080; host usa 8085 por choque con Apache

──────────────────────────────────────────────────────────────────────────────
TAREA 3: Publicar la imagen en Docker Hub
──────────────────────────────────────────────────────────────────────────────
docker login
docker tag  courses-api:dev  pmanavella/courses-api:dev
docker push pmanavella/courses-api:dev

docker tag  pmanavella/courses-api:dev  pmanavella/courses-api:v1.0
docker push pmanavella/courses-api:v1.0

# Estrategia de versionado (decisiones.md):
# - dev = desarrollo/iteraciones
# - v1.0 = estable para QA/PROD

──────────────────────────────────────────────────────────────────────────────
TAREA 4: Integrar una base de datos en contenedor + persistencia + conexión
──────────────────────────────────────────────────────────────────────────────
# Levantar/recargar servicios de compose (desde la raíz del proyecto)
docker compose down -v
docker compose up -d --build
docker compose ps
docker compose logs -f database | sed -n '1,120p'

# Prueba de persistencia (crear tabla + insertar dato, reiniciar, leer)
docker compose exec database mysql -uapp -papp -e \
"CREATE TABLE IF NOT EXISTS tp_check (id INT PRIMARY KEY, msg VARCHAR(50));
 INSERT INTO tp_check VALUES (1,'hola-docker')
 ON DUPLICATE KEY UPDATE msg='hola-docker';" arqsoft1

docker compose restart database
docker compose exec database mysql -uapp -papp -e "SELECT * FROM tp_check;" arqsoft1

# Nota para decisiones.md:
# - BD: MySQL 8 (imagen oficial) + scripts en ./database
# - Persistencia: volumen mysql_data en /var/lib/mysql
# - Conexión por nombre de servicio 'database' (DNS interno de Compose), usar DB_HOST/DB_DSN

──────────────────────────────────────────────────────────────────────────────
TAREA 5: QA y PROD con la MISMA imagen (variables de entorno)
──────────────────────────────────────────────────────────────────────────────
# Crear archivos de entorno en la raíz del repo:
# .env.qa
# (si preferís por terminal, usar heredoc)
cat > .env.qa <<'EOF'
APP_ENV=qa
APP_PORT=8080
DB_HOST=database
DB_PORT=3306
DB_USER=app
DB_PASS=app
DB_NAME=arqsoft1
DB_DSN=app:app@tcp(database:3306)/arqsoft1?parseTime=true
EOF

# .env.prod
cat > .env.prod <<'EOF'
APP_ENV=prod
APP_PORT=8080
DB_HOST=database
DB_PORT=3306
DB_USER=app
DB_PASS=app
DB_NAME=arqsoft1
DB_DSN=app:app@tcp(database:3306)/arqsoft1?parseTime=true
EOF

# (En docker-compose.yml reemplazar el servicio 'backend' por 'api-qa' y 'api-prod'
#  apuntando a la MISMA imagen estable del Docker Hub; ver Tarea 7 para v1.1)
# Validar el YAML:
docker compose config

# Levantar todo y verificar que cada servicio cargó su .env
docker compose up -d --build
docker compose ps
docker compose exec api-qa  sh -lc 'echo $APP_ENV; echo $DB_HOST; echo $DB_NAME'
docker compose exec api-prod sh -lc 'echo $APP_ENV; echo $DB_HOST; echo $DB_NAME'

# Ver imagen usada por cada servicio (debe ser la misma)
docker inspect $(docker compose ps -q api-qa)   --format '{{.Config.Image}}'
docker inspect $(docker compose ps -q api-prod) --format '{{.Config.Image}}'

# Probar endpoints de salud (requiere imagen que incluya /health)
curl -i http://localhost:8081/health
curl -i http://localhost:8082/health

# Smoke test del binario local dentro de la red de Compose (si querés validar antes de publicar)
# (reconstruir local)
docker build -t courses-api:dev ./backend
# (correr unido a la red del proyecto y con env de QA; usa 8090 en host)
docker run --rm --name api-dev \
  --network finalarqsot_i_default \
  --env-file .env.qa \
  -p 8090:8080 \
  courses-api:dev
# En otra terminal:
curl -i http://localhost:8090/health
# (Ctrl+C para cortar el contenedor de prueba)

──────────────────────────────────────────────────────────────────────────────
TAREA 6: Entorno reproducible con docker-compose
──────────────────────────────────────────────────────────────────────────────
# Levantar de cero en cualquier máquina con Docker:
git clone git@github.com:pmanavella/TP2-Docker-IS3.git
cd TP2-Docker-IS3
docker compose up -d --pull always --build
curl -i http://localhost:8081/health
curl -i http://localhost:8082/health
# Persistencia verificada:
docker compose exec database mysql -uapp -papp -e "SELECT * FROM tp_check;" arqsoft1 || true

# Notas para decisiones.md/README:
# - Requisito único: Docker
# - No exponemos 3306 (evita choques con MySQL local)
# - Healthcheck en MySQL: APIs esperan disponibilidad
# - Volumen mysql_data: datos persisten entre reinicios
# - Puertos: QA 8081, PROD 8082, Frontend 3000

──────────────────────────────────────────────────────────────────────────────
TAREA 7: Crear versión etiquetada y usarla en compose
──────────────────────────────────────────────────────────────────────────────
# Si la versión v1.0 no tenía /health, cortar v1.1 con el código actualizado:
docker build -t courses-api:dev ./backend
docker tag  courses-api:dev  pmanavella/courses-api:v1.1
docker push pmanavella/courses-api:v1.1

# Editar docker-compose.yml para que api-qa y api-prod usen v1.1
# (image: pmanavella/courses-api:v1.1) y recrear:
docker compose down
docker compose up -d --pull always

# Confirmar imagen y salud
docker inspect $(docker compose ps -q api-qa)   --format '{{.Config.Image}}'
docker inspect $(docker compose ps -q api-prod) --format '{{.Config.Image}}'
# Debe mostrar: pmanavella/courses-api:v1.1
curl -i http://localhost:8081/health
curl -i http://localhost:8082/health

# Persistencia (reverificación final)
docker compose exec database mysql -uapp -papp -e \
"CREATE TABLE IF NOT EXISTS tp_check (id INT PRIMARY KEY, msg VARCHAR(50));
 INSERT INTO tp_check VALUES (1,'hola-docker')
 ON DUPLICATE KEY UPDATE msg='hola-docker';" arqsoft1

docker compose restart database
docker compose exec database mysql -uapp -papp -e "SELECT * FROM tp_check;" arqsoft1

# Volumen de datos presente:
docker volume ls | grep mysql_data
