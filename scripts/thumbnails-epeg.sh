#!/usr/bin/env bash

epeg --width=${3} --height=${3} --max=${3} "${1}" "${2}/epeg_thumb_$(basename "${1}")"
