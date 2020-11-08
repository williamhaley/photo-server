#!/usr/bin/env bash

vipsthumbnail "${1}" --size ${3}x${3} -o "${2}/libvips_thumb_$(basename "${1}")"
