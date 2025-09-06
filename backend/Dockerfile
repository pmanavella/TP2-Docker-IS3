# Etapa de build
FROM golang:1.23 AS build
WORKDIR /app

# Copiar los archivos de go.mod y go.sum primero (para cachear dependencias)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend .

# Etapa final
FROM alpine:3.20
WORKDIR /app

# Copiar el binario desde la etapa de build
COPY --from=build /app/backend .

# Exponer puerto de la API
EXPOSE 8080

# Ejecutar el binario
CMD ["./backend"]
