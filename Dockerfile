# ---- Etapa 1: build ----
FROM golang:1.26-alpine AS build

WORKDIR /app

# Cachea las dependencias: solo se re-descargan si cambian go.mod/go.sum.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd

# ---- Etapa 2: runtime ----
# Imagen mínima con solo el binario.
FROM alpine:3.20

WORKDIR /app
COPY --from=build /server /app/server

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/app/server"]
