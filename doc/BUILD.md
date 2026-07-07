# Build

## Prerequisites

| Requirement | Version  | Notes                                          |
| ----------- | -------- | ---------------------------------------------- |
| Go          | ≥ 1.21   | Required to build from source                  |
| Docker      | ≥ 23.0   | Required to build application container images |
| Docker BuildKit | ≥ 0.11 | Required for `docker build --output`         |
| make        | any      | Used to drive the build and install targets    |

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
