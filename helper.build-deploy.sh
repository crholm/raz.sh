#!/usr/bin/env bash

set -e

host="raz.sh"

GOOS=linux GOARCH=amd64 go build razsh.go


echo "Uploading new binary and templates to ${host}..."
# Replace the old binary with the new one
scp ./razsh root@${host}:/root/razsh_new
scp -r ./data/assets ./data/tmpl  root@${host}:/root/data


# Replacing service on remote host

echo "Stopping service on ${host}..."
ssh root@${host} systemctl stop razsh.service

echo "Replacing binary on ${host}..."
ssh root@${host} mv /root/razsh_new /root/razsh

echo "Starting service on ${host}..."
ssh root@${host} systemctl start razsh.service

# Check the status of the service
 ssh root@${host} systemctl status razsh.service


