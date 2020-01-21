minivmm
=======

This is a minimal and lightweight virtual machine manager.

## Features
* Web UI
* Create/Delete/Start/Stop/Resize the VMs
* TCP/UDP DNAT to the VMs

## Requirements

### Packages

#### yum
```
# yum install qemu-system-x86 qemu-img seabios iproute genisoimage
```

#### pacman
```
# pacman -Sy qemu seabios iproute2 cdrkit
```

### OpenID Connect Provider (optional)
* We are testing with `ory/hydra`.

### Server certificate and key (optional)

## Getting started

### Standalone
```
$ export VMM_ORIGIN=https://<hostname>:14151
$ export VMM_OIDC_URL=https://<OIDCProvidor>:<OIDCProvidorPort>
$ export VMM_SERVER_CERT=/path/to/server.crt
$ export VMM_SERVER_KEY=/path/to/server.key
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

### Multi-node

#### Install with UI
```
$ export VMM_ORIGIN=https://<hostname>:14151
$ export VMM_OIDC_URL=https://<OIDCProvidor>:<OIDCProvidorPort>
$ export VMM_SERVER_CERT=/path/to/server.crt
$ export VMM_SERVER_KEY=/path/to/server.key
$ export VMM_AGENTS="hypervisor1=https://<hostname-other-node>/api/v1,hypervisor2=https://<hostname-other-node2>/api/v1"
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

#### Install without UI
```
$ export VMMINST_NO_UI=true
$ export VMM_OIDC_URL=https://<OIDCProvidor>:<OIDCProvidorPort>
$ export VMM_SERVER_CERT=/path/to/server.crt
$ export VMM_SERVER_KEY=/path/to/server.key
$ export VMM_CORS_ALLOWED_ORIGINS=https://<hostname-UI-installed-in>:14151
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

### Update

```
$ export VMMINST_UPDATE=true
$ curl -Lo - https://github.com/rsp9u/minivmm/releases/latest/download/install.sh | sh -
```

### Using Web UI
Open `https://<hostname>:14151` in your browser.

## Environments

| Name                     | Required(ui) | Required(no-ui) | Default            | Description                                                         |
|--------------------------|--------------|-----------------|--------------------|---------------------------------------------------------------------|
| VMM_DIR                  | yes          | yes             | '/opt/minivmm'     | base directory path to store state files                            |
| VMM_ORIGIN               | yes          |                 |                    | origin url of minivmm server                                        |
| VMM_OIDC_URL             | yes          | yes             |                    | oidc auth url                                                       |
| VMM_LISTEN_PORT          | yes          | yes             | '14151'            | listen port                                                         |
| VMM_AGENTS               | yes          |                 |                    | agents' API endpoint (comma separated)                              |
| VMM_CORS_ALLOWED_ORIGINS |              | yes             |                    | allowed origin urls (comma separated)                               |
| VMM_SUBNET_CIDR          | yes          | yes             | '192.168.200.0/24' | subnet CIDR for the network containing VMs                          |
| VMM_NAME_SERVERS         | yes          | yes             | '1.1.1.1,1.0.0.1'  | domain name servers' address sent via DHCP server (comma separated) |
| VMM_SERVER_CERT          | yes          | yes             |                    | path to the server certificate file                                 |
| VMM_SERVER_KEY           | yes          | yes             |                    | path to the server private key file                                 |
| VMM_NO_TLS               |              |                 |                    | disable tls if set "1" or "true"                                    |
| VMM_NO_AUTH              |              |                 |                    | skip API authentication if set "1" or "true"                        |
| VMM_NO_KVM               |              |                 |                    | disable kvm if set "1" or "true"                                    |

## Installer environments

| Name            | Default | Description                     |
|-----------------|---------|---------------------------------|
| VMMINST_VERSION | latest  | minivmm version to be installed |
| VMMINST_NO_UI   | false   | to install without UI           |
| VMMINST_UPDATE  | false   | to update minivmm               |
