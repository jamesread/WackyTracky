FROM fedora

LABEL org.opencontainers.image.source https://github.com/wacky-tracky/wacky-tracky-server
LABEL org.opencontainers.image.title=wacky-tracky-server

ADD wacky-tracky-server /server

EXPOSE 8082/tcp

ENTRYPOINT /server
