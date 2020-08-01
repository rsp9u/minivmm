FROM alpine:3.11
RUN apk add --no-cache qemu-system-x86_64 qemu-img cdrkit sudo curl iproute2 bash &&\
    [ "$(uname -m)" == "x86_64" ] && apk add --no-cache seabios || true
COPY bin/minivmm /usr/bin/minivmm
COPY script/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
