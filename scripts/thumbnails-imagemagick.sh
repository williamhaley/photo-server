#!/usr/bin/env bash

convert -sample ${3}x${3} "${1}" "${2}/imagemagick_thumb_$(basename "${1}")"
