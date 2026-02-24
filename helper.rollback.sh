#!/usr/bin/env bash

set -e

host="raz.sh"

echo "Fetching razsh images on ${host}..."
images=$(ssh root@${host} "docker images razsh --format '{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}' | sort -r")

if [[ -z "${images}" ]]; then
    echo "No razsh images found on ${host}."
    exit 1
fi

echo ""
echo "Available images (newest first):"
echo "─────────────────────────────────────────────────────────────"

i=1
declare -a tags
while IFS=$'\t' read -r tag size created; do
    printf "  [%d] %s\t(%s, %s)\n" "$i" "$tag" "$size" "$created"
    tags[$i]="$tag"
    ((i++))
done <<< "${images}"

echo "─────────────────────────────────────────────────────────────"
echo ""
read -rp "Select image number to roll back to (or q to quit): " choice

if [[ "${choice}" == "q" || "${choice}" == "Q" ]]; then
    echo "Aborted."
    exit 0
fi

if ! [[ "${choice}" =~ ^[0-9]+$ ]] || [[ "${choice}" -lt 1 ]] || [[ "${choice}" -ge "$i" ]]; then
    echo "Invalid selection."
    exit 1
fi

selected_tag="${tags[$choice]}"
image="razsh:${selected_tag}"

echo ""
echo "Rolling back to ${image} on ${host}..."
ssh root@${host} "cd /root && IMAGE=${image} docker compose up -d --no-deps razsh"

echo ""
echo "Checking container status on ${host}..."
ssh root@${host} "docker compose -f /root/docker-compose.yml ps razsh"

echo ""
echo "Rollback to ${image} complete."
