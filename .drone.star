
TAG_PATTERN = "${DRONE_TAG:-${DRONE_COMMIT_SHA:0:7}}"
OS_LIST = ["linux"]
ARCH_LIST = ["amd64", "arm64"]

def main(ctx):
  pipeline_list = []

  # per-arch pipelines
  for os in OS_LIST:
    for arch in ARCH_LIST:
      pl = {
        "kind": "pipeline",
        "name": "default-" + os + "-" + arch,
        "platform": {
          "arch": arch
        },
        "steps": []
      }

      pl["steps"].append({
        "name": "build web ui",
        "image": "node:12.13.0-stretch",
        "commands": ["make web/dist"]
      })
      pl["steps"].extend(build(ctx, os, arch))
      # prepare release with each arch
      # drone multiple pipelines can't share directories
      pl["steps"].extend([
        {
          "name": "make release",
          "image": "alpine",
          "commands": [
            "wget https://github.com/upx/upx/releases/download/v3.96/upx-3.96-%s_linux.tar.xz" % (arch),
            "tar xf upx-3.96-%s_linux.tar.xz" % (arch),
            "upx-3.96-%s_linux/upx -9 bin/*" % (arch),
            "mkdir -p release",
            "cp bin/* release",
          ]
        },
        {
          "name": "github release",
          "image": "plugins/github-release",
          "settings": {
            "api_key": {"from_secret": "github_api_key"},
            "files": "release/*",
            "checksum": ["sha256"]
          },
          "when": {"event": "tag"}
        }
      ])

      pipeline_list.append(pl)

  # all-arch pipelines
  pl = {
    "kind": "pipeline",
    "name": "default",
    "depends_on": ["default-%s-%s" % (os, arch) for os in OS_LIST for arch in ARCH_LIST],
    "steps": []
  }
  pl["steps"].extend([
    {
      "name": "make release",
      "image": "alpine",
      "commands": [
        "mkdir -p release",
        "cp script/install.sh release",
        "cp script/uninstall.sh release"
      ]
    },
    {
      "name": "github release",
      "image": "plugins/github-release",
      "settings": {
        "api_key": {"from_secret": "github_api_key"},
        "files": "release/*",
        "checksum": ["sha256"]
      },
      "when": {"event": "tag"}
    },
    {
      "name": "push-manifest",
      "image": "plugins/manifest",
      "settings": {
        "username": {"from_secret": "docker_username"},
        "password": {"from_secret": "docker_password"},
        "target": "%s:%s" % (ctx.repo.slug, TAG_PATTERN),
        "template": "%s:%s-ARCH" % (ctx.repo.slug, TAG_PATTERN),
        "platforms": ["%s/%s" % (os, arch) for os in OS_LIST for arch in ARCH_LIST]
      }
    }
  ])
  pipeline_list.append(pl)

  return pipeline_list

def build(ctx, os, arch):
  return [
    {
      "name": "build %s-%s" % (os, arch),
      "image": "golang:1.13",
      "commands": [
        "GOOS=%s GOARCH=%s make" % (os, arch)
      ]
    },
    {
      "name": "build docker %s-%s" % (os, arch),
      "image": "plugins/docker",
      "settings": {
        "username": {"from_secret": "docker_username"},
        "password": {"from_secret": "docker_password"},
        "repo": ctx.repo.slug,
        "tags": "%s-%s" % (TAG_PATTERN, arch),
        "build_args": [
          "ARCH=" + arch
        ]
      }
    },
    {
      "name": "move release %s-%s" % (os, arch),
      "image": "golang:1.13",
      "commands": [
        "mv bin/minivmm bin/minivmm_%s_%s" % (os, arch)
      ]
    }
  ]
