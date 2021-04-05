FROM --platform=${BUILDPLATFORM} alpine:3.13.4

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

LABEL maintainer="Gareth Evans <gareth@bryncynfelin.co.uk>"
COPY dist/captain-hook-${TARGETOS}_${TARGETOS}_${TARGETARCH}/captain-hook /usr/bin/captain-hook

ENTRYPOINT [ "/usr/bin/captain-hook" ]

CMD [ "listen" , "--debug" ]
