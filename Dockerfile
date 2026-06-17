# ------------------
# Frontend build stage
# ------------------
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml* ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
ENV API_BASE_URL=/
RUN pnpm run generate

# ------------------
# Backend build stage
# ------------------
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY --from=frontend-builder /app/frontend/.output/public ./internal/static/web/
RUN CGO_ENABLED=0 GOOS=linux go build -o orchestrator ./cmd/main.go

# ------------------
# Backend dev stage
# ------------------
FROM golang:alpine AS backend-dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
WORKDIR /app
ENV ORCHESTRATOR_API_BIND=:8081
ENV MC_SERVERS_DIR=/app/servers
ENV MC_INSTANCE_FILE=/app/data/instance.yml
ENV ORCHESTRATOR_LOG_LEVEL=info
CMD ["air"]

# ------------------
# Frontend dev stage
# ------------------
FROM node:22-alpine AS frontend-dev
RUN corepack enable
WORKDIR /app/frontend
ENV CI=true
ENV BACKEND_URL=http://backend:8081
ENV API_BASE_URL=http://localhost:3000
ENV WS_BASE_URL=ws://localhost:8081
CMD ["sh", "-c", "pnpm install && pnpm run dev --host"]

# ------------------
# Runtime stage
# ------------------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/backend/orchestrator /app/orchestrator
RUN chmod +x /app/orchestrator

RUN mkdir -p /app/servers /app/data

ENV ORCHESTRATOR_API_BIND=:8081
ENV MC_SERVERS_DIR=/app/servers
ENV MC_INSTANCE_FILE=/app/data/instance.yml
ENV ORCHESTRATOR_LOG_LEVEL=info

EXPOSE 8081

CMD ["/app/orchestrator"]
