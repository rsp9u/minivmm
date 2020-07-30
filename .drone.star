
TAG_PATTERN = "${DRONE_TAG:-${DRONE_COMMIT_SHA:0:7}}"
def main(ctx):
  pl = {
    "kind": "pipeline",
    "name": "default",
    "steps": []
  }

  pl["steps"].append({
    "name": "build web ui",
    "image": "node:12.13.0-stretch",
    "commands": ["make web/dist"]
  })
  pl["steps"].extend(build(ctx, "linux", "amd64"))
  pl["steps"].extend(build(ctx, "linux", "arm64"))
  pl["steps"].extend([
    {
      "name": "make release",
      "image": "alpine",
      "commands": [
        "apk add upx",
        "upx -9 bin/*",
        "mkdir -p release",
        "cp bin/* release",
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
        "platforms": [
          "linux/amd64",
          "linux/arm64"
        ]
      }
    }
  ])

  return pl


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
