FROM registry.fedoraproject.org/fedora-minimal:43

LABEL org.opencontainers.image.source=https://github.com/wacky-tracky/wacky-tracky-server
LABEL org.opencontainers.image.authors="James Read"
LABEL org.opencontainers.image.title=WackyTracky

# HTTP server (API + web UI) on 8080.
EXPOSE 8080/tcp

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/wacky-tracky-server /app/wt
ADD frontend/dist /app/webui/

WORKDIR /app

# Defaults so config.yaml is optional: todotxt driver with /app/data as tasks dir.
ENV DATABASE_DRIVER=todotxt
ENV DATABASE_DATABASE=/app/data

RUN mkdir -p /config /app/data
VOLUME ["/config", "/app/data"]

ENTRYPOINT [ "/app/wt" ]
