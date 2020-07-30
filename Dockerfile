FROM alpine:3.11
ARG ARCH=amd64
RUN apk add --no-cache qemu-system-x86_64 qemu-img seabios cdrkit sudo curl iproute2 bash
RUN \
 curl -Lo /usr/bin/minivmm https://github.com/rsp9u/minivmm/releases/download/v0.2.10/minivmm_linux_$ARCH &&\
 chmod +x /usr/bin/minivmm
COPY script/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh
ENTRYPOINT /entrypoint.sh
