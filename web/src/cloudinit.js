const pwauth = `#cloud-config
password: Pa$$w0rd
chpasswd:
  expire: False
ssh_pwauth: True`

const pubkeyauth = `#cloud-config
ssh_pwauth: False
ssh_authorized_keys:
  - ssh-rsa AAA...SDvZ user1@domain.com`

export default {
  templates: [
    { text: "empty", value: "" },
    { text: "password auth", value: pwauth },
    { text: "public key auth", value: pubkeyauth },
  ]
};
