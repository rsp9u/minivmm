FROM alpine:3.11
ARG ARCH=amd64
RUN apk add --no-cache qemu-system-x86_64 qemu-img seabios cdrkit sudo curl iproute2 bash
COPY bin/minivmm /usr/bin/minivmm
COPY script/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
