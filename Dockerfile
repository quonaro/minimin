# syntax=docker/dockerfile:1

# ------------------
# Frontend build stage
# ------------------
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json ./
RUN npm install

COPY frontend/ ./
ENV API_BASE_URL=/
RUN npm run generate

# ------------------
# Backend build stage
# ------------------
FROM golang:alpine AS backend-builder
WORKDIR /app/backend

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orchestrator ./cmd/main.go

# ------------------
# Runtime stage
# ------------------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates caddy supervisor

WORKDIR /app

COPY --from=frontend-builder /app/frontend/.output/public /app/frontend
COPY --from=backend-builder /app/backend/orchestrator /app/orchestrator
RUN chmod +x /app/orchestrator

COPY docker/Caddyfile /etc/caddy/Caddyfile
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN mkdir -p /app/servers /app/data

ENV ORCHESTRATOR_API_BIND=:8081
ENV MC_SERVERS_DIR=/app/servers
ENV MC_INSTANCE_FILE=/app/data/instance.yml
ENV ORCHESTRATOR_LOG_LEVEL=info

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
