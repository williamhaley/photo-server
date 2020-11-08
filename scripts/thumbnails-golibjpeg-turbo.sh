#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
"${DIR}/golibjpeg-turbo" "${1}" "${2}/golibjpeg-turbo_$(basename "${1}")" ${3}
