# Requirements

### sso platform
- OIDC Provider (tested with `hydra`)

### server certificate and key (optional)

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

### Standalone
```
$ export VMM_ORIGIN=https://...
$ export VMM_OIDC_URL=https://...
$ export VMM_SERVER_CERT=/path/to/server.crt
$ export VMM_SERVER_KEY=/path/to/server.key
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

### Multi-node

#### Install with UI
```
$ export VMM_DIR=/opt/minivmm
$ export VMM_ORIGIN=https://...
$ export VMM_OIDC_URL=https://...
$ export VMM_AGENTS="hypervisor1=https://hypervisor1.localdomain/api/v1,hypervisor2=https://hypervisor2.localdomain/api/v1"
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

#### Install without UI
```
$ export VMM_NO_UI=true
$ export VMM_CORS_ALLOWED_ORIGINS=https://hypervisor1.localdomain:14151
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

### Update

```
$ export VMM_UPDATE=true
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

# Environments

| Name                     | Required(ui) | Required(no-ui) | Default           | Description                                                         |
|--------------------------|--------------|-----------------|-------------------|---------------------------------------------------------------------|
| VMM_DIR                  | yes          | yes             | '/opt/minivmm'    | base directory path to store state files                            |
| VMM_ORIGIN               | yes          |                 |                   | origin url of minivmm server                                        |
| VMM_OIDC_URL             | yes          |                 |                   | oidc auth url                                                       |
| VMM_LISTEN_PORT          | yes          | yes             | '14151'           | listen port                                                         |
| VMM_AGENTS               | yes          |                 |                   | agents' API endpoint (comma separated)                              |
| VMM_CORS_ALLOWED_ORIGINS |              | yes             |                   | allowed origin urls (comma separated)                               |
| VMM_NAME_SERVERS         | yes          | yes             | '1.1.1.1,1.0.0.1' | domain name servers' address sent via DHCP server (comma separated) |
| VMM_SERVER_CERT          | yes          | yes             |                   | path to the server certificate file                                 |
| VMM_SERVER_KEY           | yes          | yes             |                   | path to the server private key file                                 |
| VMM_NO_TLS               |              |                 |                   | disable tls if set "1" or "true"                                    |
| VMM_NO_AUTH              |              |                 |                   | skip API authentication if set "1" or "true"                        |
| VMM_NO_KVM               |              |                 |                   | disable kvm if set "1" or "true"                                    |

# Installer environments

| Name        | Default | Description                     |
|-------------|---------|---------------------------------|
| VMM_VERSION | latest  | minivmm version to be installed |
| VMM_NO_UI   | false   | to install without UI           |
| VMM_UPDATE  | false   | to update minivmm               |
