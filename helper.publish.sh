#!/usr/bin/env bash

host="raz.sh"

dest="root@${host}:/root/data"

dir="data/blog"

rsync -avz --update --progress "${dir}" "${dest}"


