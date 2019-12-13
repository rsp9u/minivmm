#!/bin/sh -ex
BIN=/usr/local/bin/minivmm
USR=minivmm
CACHE_DIR=$HOME/.cache/minivmm

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | awk '{ if($1=="aarch64_be"||$1=="aarch64"||$1=="armv8b"||$1=="armv8l") {print "arm64"} else {print "amd64"} }')
export VMM_DIR=${VMM_DIR:-/opt/minivmm}
export VMM_LISTEN_PORT=${VMM_LISTEN_PORT:-14151}

if [ "$VMM_UPDATE" != "" ]; then
  sudo systemctl stop minivmm.service
fi

# Retrieve binary
if [ ! -f $CACHE_DIR/$VMM_VERSION/minivmm_${os}_${arch} ]; then
  mkdir -p $CACHE_DIR/$VMM_VERSION
  curl -Lo $CACHE_DIR/$VMM_VERSION/minivmm_${os}_${arch} https://github.com/rsp9u/minivmm/releases/download/$VMM_VERSION/minivmm_${os}_${arch}
fi
sudo mv $CACHE_DIR/$VMM_VERSION/minivmm_${os}_${arch} $BIN
sudo chmod +x $BIN
sudo setcap 'CAP_NET_BIND_SERVICE,CAP_NET_RAW=+eip' $BIN

# verify checksum
curl -Lo - https://github.com/rsp9u/minivmm/releases/download/$VMM_VERSION/sha256sum.txt | \
  grep minivmm_${os}_${arch} | \
  sed -e 's,release/'minivmm_${os}_${arch}','$BIN',' | \
  sha256sum -c -
if [ $? -ne 0 ]; then echo "failed to verify checksum.\nplease retry."; exit 1; fi

if [ "$VMM_UPDATE" != "" ]; then
  sudo systemctl start minivmm.service
  exit 0
fi

# Setup service user
grep -q $USR /etc/passwd || sudo useradd $USR -b $(dirname $VMM_DIR)
echo "Defaults:$USR !requiretty" | sudo tee /etc/sudoers.d/$USR > /dev/null
echo "$USR ALL=(ALL) NOPASSWD:/sbin/ip" | sudo tee -a /etc/sudoers.d/$USR > /dev/null
sudo chmod 440 /etc/sudoers.d/$USR

# Setup data directory
sudo mkdir -p $VMM_DIR
sudo chown -R $USR:$USR $VMM_DIR

# Register to systemd
cat << EOS | sudo tee $VMM_DIR/minivmm.environment
VMM_DIR=${VMM_DIR}
VMM_ORIGIN=${VMM_ORIGIN}
VMM_OIDC_URL=${VMM_OIDC_URL}
VMM_LISTEN_PORT=${VMM_LISTEN_PORT}
VMM_AGENTS=${VMM_AGENTS}
VMM_CORS_ALLOWED_ORIGINS=${VMM_CORS_ALLOWED_ORIGINS}
VMM_NAME_SERVERS=${VMM_NAME_SERVERS}
VMM_NO_TLS=${VMM_NO_TLS}
VMM_NO_AUTH=${VMM_NO_AUTH}
VMM_NO_KVM=${VMM_NO_KVM}
EOS
sudo chown $USR:$USR $VMM_DIR/minivmm.environment

if [ "$VMM_INSTALL_UI" != "" ]; then
  UI_ARG="-ui"
else
  UI_ARG=""
fi
cat << EOS | sudo tee /etc/systemd/system/minivmm.service
[Unit]
Description=Minimal VM Manager

[Service]
Restart=always
KillMode=process
User=${USR}
Group=${USR}
EnvironmentFile=${VMM_DIR}/minivmm.environment
ExecStartPre=${BIN} -init-nw
ExecStart=${BIN} ${UI_ARG}
ExecStop=/bin/pkill minivmm

[Install]
WantedBy=multi-user.target
EOS

sudo systemctl enable minivmm.service
