FROM --platform=${BUILDPLATFORM} alpine:3.13.5

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

COPY build/linux/bosh-compile /usr/bin/bosh-compile

ENTRYPOINT [ "/usr/bin/bosh-compile" ]

CMD ["--help"]
