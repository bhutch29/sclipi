#!/bin/sh
set -e

echo "Starting scpir-server and Caddy..."
./scpir-server &
SERVER_PID=$!

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
CADDY_PID=$!

echo "Monitoring processes"
while true; do
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo "scpir-server exited, killing Caddy and quitting"
        kill $CADDY_PID 2>/dev/null || true
        exit 1
    fi
    if ! kill -0 $CADDY_PID 2>/dev/null; then
        echo "Caddy exited, killing scpir-server and quitting"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
    sleep 1
done
