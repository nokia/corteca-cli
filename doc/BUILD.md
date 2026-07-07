# Build

## Prerequisites

| Requirement     | Version  | Notes                                          |
| -----------     | -------- | ---------------------------------------------- |
| Go              | ≥ 1.21   | Required to build from source                  |
| Docker          | ≥ 23.0   | Required to build application container images |
| Docker BuildKit | ≥ 0.11   | Required for `docker build --output`           |
| make            | any      | Used to drive the build and install targets    |

## Build locally

Clone the repository and run:

```shell
make
```

The compiled binary is placed in the `dist/` directory.

## Build using Docker

If you do not have a local Go toolchain, you can build entirely inside Docker
(BuildKit is required):

```shell
docker build --output ./dist .
```

### Installing Docker BuildKit

On Docker Engine < 23.0, BuildKit must be enabled manually. On Ubuntu 22.04:

```shell
sudo apt-get install docker-buildx-plugin
```

Then either prefix each `docker build` invocation with `DOCKER_BUILDKIT=1`, or
follow the [official instructions](https://docs.docker.com/build/buildkit/#getting-started)
to enable it globally.

## Build the corteca-cli container image

`corteca-cli.Dockerfile` produces a self-contained runtime image that bundles
the corteca binary together with the Docker CLI (required for
`docker buildx build`). This is the image published to GHCR on every merge to
`main`.

### Build the image locally

The image expects pre-built binaries in `dist/bin/`. Build them first, then
build the image:

```shell
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Build the Linux binary for your host architecture (amd64 or arm64)
make DESTOS=linux DESTARCH=amd64

docker build \
  -f corteca-cli.Dockerfile \
  --platform linux/amd64 \
  --build-arg VERSION=$VERSION \
  -t corteca-cli:local .
```

### Test the image

Corteca needs access to the host Docker daemon at runtime, so mount the Docker
socket:

```shell
# Verify the version reported by the binary
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  corteca-cli:local --version

# Open an interactive session with a project directory mounted
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  -w /workspace \
  corteca-cli:local
```
