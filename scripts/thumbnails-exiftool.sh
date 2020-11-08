#!/usr/bin/env bash

exiftool -b -ThumbnailImage "${1}" > "${2}/exiftool_thumb_$(basename "${1}")"
