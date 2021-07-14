FROM --platform=${BUILDPLATFORM} alpine:3.13.5

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

COPY build/linux/bosh-compile /usr/bin/bc
COPY entrypoint.sh /usr/bin/entrypoint.sh
