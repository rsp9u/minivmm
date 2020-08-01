FROM alpine:3.11 AS base
RUN apk add --no-cache qemu-img cdrkit sudo curl iproute2 bash
COPY bin/minivmm /usr/bin/minivmm
COPY script/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

FROM base AS amd64
RUN apk add --no-cache qemu-system-x86_64 seabios

FROM base AS arm64
RUN apk add --no-cache qemu-system-aarch64 &&\
    mkdir -p /usr/share/qemu-efi-aarch64 &&\
    wget -O /usr/share/qemu-efi-aarch64/QEMU_EFI.fd https://releases.linaro.org/components/kernel/uefi-linaro/latest/release/qemu64/QEMU_EFI.fd
