#!/usr/bin/env bash

set -e

host="raz.sh"
image="razsh:$(date +%Y-%m-%d_%H.%M)-$(git rev-parse --short HEAD)"

echo "Building Docker image ${image}..."
docker build -t "${image}" -t "razsh:latest" .

echo "Transferring image to ${host}..."
docker save "${image}" "razsh:latest" | ssh root@${host} docker load

echo "Restarting service on ${host}..."
ssh root@${host} "cd /root && IMAGE=${image} docker compose up -d --no-deps razsh"

echo "Checking container status on ${host}..."
ssh root@${host} "docker compose -f /root/docker-compose.yml ps razsh"
