# Requirements

### sso platform
- OIDC Provider (tested with `hydra`)

### packages

#### yum
```
# yum install qemu-system-x86 qemu-img seabios iproute genisoimage
```

#### pacman
```
# pacman -Sy qemu seabios iproute2 cdrkit
```

# Getting started

### Install with UI
```
$ export VMM_DIR=/opt/minivmm
$ export VMM_ORIGIN=https://...
$ export VMM_OIDC_URL=https://...
$ export VMM_AGENTS="hypervisor1=https://hypervisor1.localdomain/api/v1,hypervisor2=https://hypervisor2.localdomain/api/v1"
$ export VMM_NAME_SERVERS="1.1.1.1,1.0.0.1"
$ export VMM_INSTALL_UI=yes
$ export VMM_VERSION=...
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/download/$VMM_VERSION/install.sh | sh -
$ sudo -u minivmm /usr/local/bin/minivmm -init-nw
$ sudo systemctl start minivmm.service
$ sudo iptables -t nat -A POSTROUTING -o eth1 -s 192.168.200.0/24 -j MASQUERADE
```

### Install without UI
```
$ unset VMM_INSTALL_UI
$ export VMM_CORS_ALLOWED_ORIGINS=https://hypervisor1.localdomain:14151
$ export VMM_VERSION=...
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/download/$VMM_VERSION/install.sh | sh -
$ sudo -u minivmm /usr/local/bin/minivmm -init-nw
$ sudo systemctl start minivmm.service
$ sudo iptables -t nat -A POSTROUTING -o eth1 -s 192.168.200.0/24 -j MASQUERADE
```

### Update

```
$ export VMM_UPDATE=yes
$ export VMM_VERSION=...
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/download/$VMM_VERSION/install.sh | sh -
```

### Server certifications

Put `server.crt` and `server.key` into `${VMM_DIR}`.

# Environments

* VMM_DIR: base directory path to store state files
* VMM_ORIGIN: origin url of minivmm server
* VMM_OIDC_URL: oidc auth url
* VMM_LISTEN_PORT: listen port
* VMM_AGENTS: agents' API endpoint (comma separated)
* VMM_CORS_ALLOWED_ORIGINS: allowed origin urls (comma separated)
* VMM_NAME_SERVERS: domain name servers' address sent via DHCP server (comma separated)
* VMM_NO_TLS: disable tls if set "1" or "true"
* VMM_NO_AUTH: skip API authentication if set "1" or "true"
* VMM_NO_KVM: disable kvm if set "1" or "true"
