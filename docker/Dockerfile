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
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o scpir-server ./cmd/server

FROM caddy:2-alpine

WORKDIR /app

RUN apk add --no-cache tini

COPY --from=go-builder /build/scpir-server .

COPY --from=web-builder /build/web/dist/scpir-web/browser /srv

COPY docker/Caddyfile /etc/caddy/Caddyfile

COPY docker/start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/start.sh"]
