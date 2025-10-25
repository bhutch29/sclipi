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

RUN apk add --no-cache tini

COPY --from=go-builder /build/sclipi-server .

COPY --from=web-builder /build/web/dist/sclipi-web/browser /srv

RUN printf ':80 {\n\
\tencode gzip\n\
\n\
\thandle /api/* {\n\
\t\turi strip_prefix /api\n\
\t\treverse_proxy localhost:8080\n\
\t}\n\
\n\
\thandle {\n\
\t\troot * /srv\n\
\t\ttry_files {path} /index.html\n\
\t\tfile_server\n\
\t}\n\
}\n' > /etc/caddy/Caddyfile

RUN printf '#!/bin/sh\n\
set -e\n\
\n\
echo "Starting sclipi-server on port 8080..."\n\
./sclipi-server &\n\
SERVER_PID=$!\n\
\n\
echo "Waiting for server to be ready..."\n\
for i in $(seq 1 30); do\n\
\tif wget -q -O /dev/null http://localhost:8080/health 2>/dev/null; then\n\
\t\techo "Server is ready!"\n\
\t\tbreak\n\
\tfi\n\
\tif [ $i -eq 30 ]; then\n\
\t\techo "Server failed to start"\n\
\t\texit 1\n\
\tfi\n\
\tsleep 1\n\
done\n\
\n\
echo "Starting Caddy..."\n\
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile\n\
' > /app/start.sh && chmod +x /app/start.sh

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/start.sh"]
