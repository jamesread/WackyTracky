FROM registry.fedoraproject.org/fedora-minimal:40

LABEL org.opencontainers.image.source=https://github.com/wacky-tracky/wacky-tracky-server
LABEL org.opencontainers.image.authors=James Read
LABEL org.opencontainers.image.title=wacky-tracky-server

EXPOSE 8082/tcp

ADD wacky-tracky-server /app/wt
ADD frontend/dist /app/webui/

ENTRYPOINT [ "/app/wt" ]
