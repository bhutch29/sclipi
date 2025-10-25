#!/bin/sh
set -e

echo "Starting sclipi-server on port 8080..."
./sclipi-server &
SERVER_PID=$!

echo "Waiting for server to be ready..."
for i in $(seq 1 30); do
	if wget -q -O /dev/null http://localhost:8080/health 2>/dev/null; then
		echo "Server is ready!"
		break
	fi
	if [ $i -eq 30 ]; then
		echo "Server failed to start"
		exit 1
	fi
	sleep 1
done

echo "Starting Caddy..."
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
