FROM fedora

ADD wacky-tracky-server /server

EXPOSE 8082/tcp

ENTRYPOINT /server
