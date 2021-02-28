FROM golang

RUN apt update && apt install -y --no-install-recommends build-essential gcc-mingw-w64 cmake file clang

WORKDIR /go/src/app

CMD ["app"]
