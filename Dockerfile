FROM --platform=${BUILDPLATFORM} alpine:3.13.5

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

COPY build/linux/bosh-compile /usr/bin/bosh-compile
COPY entrypoint.sh /usr/bin/entrypoint.sh

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
