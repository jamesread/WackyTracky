FROM registry.fedoraproject.org/fedora-minimal:43

LABEL org.opencontainers.image.source=https://github.com/wacky-tracky/wacky-tracky-server
LABEL org.opencontainers.image.authors="James Read"
LABEL org.opencontainers.image.title=WackyTracky

# HTTP server (API + web UI) on 8080.
EXPOSE 8080/tcp

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/wacky-tracky-server /app/wt
ADD frontend/dist /app/webui/
COPY database/sqlite/migrations /var/app/database/sqlite/migrations

WORKDIR /app

# Defaults so config.yaml is optional: todotxt driver with /app/data as tasks dir.
ENV DATABASE_DRIVER=todotxt
ENV DATABASE_DATABASE=/app/data
ENV IAM_DATABASE=/app/data/iam.sqlite
ENV HOME=/home/wt
ENV WACKYTRACKY_MIGRATIONS_DIR=/var/app/database/sqlite/migrations

RUN microdnf install -y git openssh-clients && microdnf clean all
RUN mkdir -p /config /app/data /home/wt/.ssh \
	&& chmod 700 /home/wt/.ssh
VOLUME ["/config", "/app/data", "/home/wt"]

ENTRYPOINT [ "/app/wt" ]
