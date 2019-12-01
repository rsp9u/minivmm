def main(ctx):
  return {
    "kind": "pipeline",
    "name": "default",
    "steps": [
      {
        "name": "build web ui",
        "image": "node:12.13.0-stretch",
        "commands": ["make web/dist"]
      },
      build("linux", "amd64"),
      build("linux", "arm64"),
      {
        "name": "make release",
        "image": "alpine",
        "commands": [
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
          "files": "release/*"
        },
        "when": {"event": "tag"}
      }
    ]
  }

def build(os, arch):
  return {
    "name": "build %s-%s" % (os, arch),
    "image": "golang:1.13",
    "commands": [
      "GOOS=%s GOARCH=%s make" % (os, arch),
      "mv bin/minivmm bin/minivmm_%s_%s" % (os, arch)
    ]
  }
