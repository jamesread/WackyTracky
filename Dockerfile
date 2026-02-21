FROM registry.fedoraproject.org/fedora-minimal:43

LABEL org.opencontainers.image.source=https://github.com/wacky-tracky/wacky-tracky-server
LABEL org.opencontainers.image.authors="James Read"
LABEL org.opencontainers.image.title=WackyTracky

# Default single HTTP frontend (API + web UI). Config can also expose 8082 (REST), 8083 (gRPC), 8084 (web UI).
EXPOSE 8443/tcp

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/wacky-tracky-server /app/wt
ADD frontend/dist /app/webui/

# Working dir so viper finds ./config.yaml when you mount a file at /app/config.yaml
WORKDIR /app

# Optional mount points: /config (config.yaml for containers), /app/data (e.g. todotxt DB dir)
RUN mkdir -p /config /app/data
VOLUME ["/config", "/app/data"]

ENTRYPOINT [ "/app/wt" ]
