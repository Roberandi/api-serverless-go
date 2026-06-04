# Usamos una versión oficial de Go
FROM golang:1.26.3-alpine

# Creamos una carpeta de trabajo dentro del contenedor
WORKDIR /app

# Copiamos los archivos de dependencias y las descargamos
COPY go.mod go.sum ./
RUN go mod download

# Copiamos todo el resto de tu código
COPY . .

# Construimos (compilamos) la aplicación
RUN go build -o api-server ./cmd/api/main.go

# Exponemos el puerto 8080
EXPOSE 8080

# Le decimos a Docker qué comando ejecutar al encender
CMD ["./api-server"]