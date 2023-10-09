FROM --platform=$BUILDPLATFORM node:20-alpine3.18 as assets

RUN apk update && apk upgrade && apk add --no-cache just curl

COPY . /stuff

WORKDIR /stuff

RUN just build-js build-js build-styles build-icons _fonts

FROM --platform=$BUILDPLATFORM golang:1.21.2-alpine3.18 as builder

RUN apk update && apk upgrade && apk add --no-cache just

COPY . /stuff
COPY --from=assets /stuff/static/build /stuff/static/build

WORKDIR /stuff

ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    GO_LDFLGAS="-s -w" GO_BUILD_FLAGS="-trimpath" \
    just _build-go

# Create a nonroot user to copy
RUN echo "nobody:*:65534:65534:nobody:/_nonexistent:/bin/false" > /etc/passwd.nonroot

RUN mkdir -p /run/stuff/files /tmp/stuff

FROM scratch

COPY --from=builder /stuff/build/stuff /stuff
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd.nonroot /etc/passwd
COPY --from=builder --chown=nobody:nobody /run/stuff /run
COPY --from=builder --chown=nobody:nobody /tmp/stuff /tmp

ENV STUFF_DATABASE_PATH="/run/stuff.db"
ENV STUFF_FILE_DIR="/run/files"

USER nobody

CMD ["/stuff"]
