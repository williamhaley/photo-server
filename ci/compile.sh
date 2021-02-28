#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

(
  cd "${DIR}/../ui"
  npm i
  npm run build
  cd "${DIR}/../"
  env CGO_ENABLED=1 go build -o photo-server main.go
)
