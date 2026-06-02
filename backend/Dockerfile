# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

COPY backend/go.mod ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/delgate-api ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/delgate-api /app/delgate-api

ENV PORT=8080 \
    BACKEND_BIND_ADDR=0.0.0.0 \
    QUOTE_HISTORY_LIMIT=1000 \
    MAX_REQUEST_BODY_BYTES=1048576 \
    HTTP_READ_TIMEOUT=10s \
    HTTP_READ_HEADER_TIMEOUT=5s \
    HTTP_WRITE_TIMEOUT=15s \
    HTTP_IDLE_TIMEOUT=60s

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:${PORT}/api/health >/dev/null || exit 1
USER app
CMD ["/app/delgate-api"]
