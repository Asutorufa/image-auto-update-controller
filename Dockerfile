FROM docker.io/library/golang:alpine AS builder

WORKDIR /image-auto-update-controller

COPY . .

ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

RUN set -x\
    && apk add --no-cache make git\
    && go build -ldflags='-s -w -buildid=' -v -o image-auto-update-controller .

FROM docker.io/library/alpine:latest

WORKDIR /image-auto-update-controller

COPY --from=builder /image-auto-update-controller/image-auto-update-controller /image-auto-update-controller/image-auto-update-controller

ENTRYPOINT ["/image-auto-update-controller/image-auto-update-controller"]
