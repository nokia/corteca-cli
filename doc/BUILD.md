# Build

To build the tool, use the provided `Makefile` as follows:

```bash
make
```

Keep in mind that Go v1.21 is required.

## Build using docker

### Prerequisites

If you don't have the required development environment you can also build using docker (BuildKit is required, see below). Use the following command inside the project's root folder:

```bash
docker build --output ./dist .
```

### Installing docker BuildKit builder

As per the [documentation](https://docs.docker.com/build/buildkit/#getting-started), if you are using a docker engine prior to v23.0 you need to manually (install and) enable the buildkit builder. In Ubuntu v22.04, you can use the following command to install it:

```bash
sudo apt-get install docker-buildx-plugin
```

Afterwards, either prepend `DOCKER_BUILDKIT=1` to each of the aforementioned `docker build ...` commands, or follow the instructions in the provided link, to enable buildkit by default.

## Install

To install the binary in the default `$GOBIN` path (defaults to `$HOME/go/bin`)
as well as the necessary template files (in `$HOME/.config/corteca`), use
the following command:

```bash
make install
```

To uninstall a previous installation, use:

```bash
make uninstall
```
