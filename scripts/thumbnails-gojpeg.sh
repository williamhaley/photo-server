#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
"${DIR}/gojpeg" "${1}" "${2}/gojpeg_thumb_$(basename "${1}")" ${3}
