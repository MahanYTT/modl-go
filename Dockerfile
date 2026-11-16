# Stage 1: Build Go binary (with embedded panel SPA)
FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/modl-backend ./cmd/modl

# Stage 2: Minimal distroless/alpine runtime (~35 MB image, ~15-30 MB RAM)
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/modl-backend /usr/local/bin/modl-backend

VOLUME ["/app/data"]
ENV PORT=8080 \
    MODL_DB_PATH=/app/data/modl.db \
    MODL_STORAGE_DIR=/app/data/storage \
    MODL_SINGLE_TENANT=true

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/modl-backend"]
