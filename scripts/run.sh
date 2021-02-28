#!/usr/bin/env bash

docker run \
  -it \
  -d \
  --restart=always \
  -p 8080:8080 \
  -p 9090:9090 \
  --name photo-server \
  -v `pwd`:/go/src/app \
  -v $HOME/dev/photo-server-data:/one \
  -v /mnt/data/photo-server-data:/two \
  williamhaley/photo-server \
  go run main.go serve \
    -photos-directory /two/FamilyPhotos \
    -data-directory /two/data \
    -thumbnails-directory /two/thumbs \
    -http-port 8080 \
    -https-port 9090 \
    -https-cert-file /two/photos.willhy.com/cert.pem \
    -https-cert-key /two/photos.willhy.com/privkey.pem \
    -access-code "haley"
