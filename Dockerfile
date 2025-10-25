FROM node:20-alpine AS web-builder

WORKDIR /build/web

COPY web/package*.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS go-builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN apk add --no-cache git && \
    VERSION=$(git describe --always --long --dirty 2>/dev/null || echo "docker") && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o sclipi-server ./cmd/server

FROM caddy:2-alpine

WORKDIR /app

COPY --from=go-builder /build/sclipi-server .

COPY --from=web-builder /build/web/dist/sclipi-web/browser /srv

RUN echo ':80 {\n\
    root * /srv\n\
    encode gzip\n\
    \n\
    handle /api/* {\n\
        uri strip_prefix /api\n\
        reverse_proxy localhost:8080\n\
    }\n\
    \n\
    try_files {path} /index.html\n\
    file_server\n\
}' > /etc/caddy/Caddyfile

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

CMD ./sclipi-server & caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
