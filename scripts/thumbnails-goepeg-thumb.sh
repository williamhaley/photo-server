#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
"${DIR}/goegeg-thumb" "${1}" "${2}/goepeg-thumb_$(basename "${1}")" ${3}
