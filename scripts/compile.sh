#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

bash ${DIR}/build-containers.sh

# docker run \
#   -it \
#   --rm \
#   --workdir="/app" \
#   -v "${DIR}/../ui:/app" \
#   node \
#   bash

# exit 1

# Cross compiling is kind o a nightmare. Assume this is run from the desired target.
docker run \
  -it \
  --rm \
  -v "${DIR}/../":/go/src/app \
  williamhaley/photo-server \
  env GOROOT=/usr/local/go CGO_ENABLED=1 go build -o photo-server main.go
