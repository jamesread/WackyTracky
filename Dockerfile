FROM fedora

COPY . /opt/

RUN dnf install python3-configargparse -y && dnf clean all

EXPOSE 8082/tcp

ENTRYPOINT /opt/api/server.py
