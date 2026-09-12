# ==========================================
# ETAPA 1: Compilação dos binários em Go
# ==========================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY . .

# Compila os 3 executáveis
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/driver-client ./cmd/driver-client/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/passenger-client ./cmd/passenger-client/main.go

# Servidor
FROM alpine:latest AS server
WORKDIR /app
COPY --from=builder /app/bin/server .
EXPOSE 8080
CMD ["./server"]

#Motorista
FROM alpine:latest AS driver-client
WORKDIR /app
COPY --from=builder /app/bin/driver-client .
CMD ["./driver-client"]

#Passageiro
FROM alpine:latest AS passenger-client
WORKDIR /app
COPY --from=builder /app/bin/passenger-client .
CMD ["./passenger-client"]