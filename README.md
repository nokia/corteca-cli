# Corteca Command Line Interface (Corteca CLI)

Part of Corteca Developer Toolkit, Corteca CLI facilitates bootstapping and development of container applications compatible with [Corteca Marketplace](https://www.nokia.com/networks/fixed-networks/corteca-applications/). Developers can easily create a basic structure and ensures the appropriate format for the applications by targetting to specific platform/architectures.

## Build

### Build using native Go toolchain

To build the tool, use the provided `Makefile` as follows:

```bash
make
```

Keep in mind that Go v1.21 is required.

### Build using docker

If you don't have the required development environment you can also build using
docker (BuildKit is required, see below). Use the following command inside
the project root folder:

```bash
docker build --output ./dist .
```

#### Installing docker BuildKit builder

As per the [documentation](https://docs.docker.com/build/buildkit/#getting-started),
if you are using a docker engine prior to v23.0 you need to manually (install
and) enable the buildkit builder. In Ubuntu v22.04, you can use the following
cmd to install it:

```bash
sudo apt-get install docker-buildx-plugin
```

Afterwards, either prepend `DOCKER_BUILDKIT=1` to each of the afforementioned
`docker build ...` commands, or follow the instructions in the provided link, to
enable buildkit by default.

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

## Usage

Use `corteca help` to obtain information on how to invoke the various
commands supported by the tool.

### Generate a new application skeleton

To generate a new application in the current folder, use:

```bash
corteca create .
```

This will prompt you to enter information about the application and also select
the application language.

### Build the application to produce an app package

To build the application, use:

```text
corteca build [<TOOLCHAIN>]
```

If `<TOOLCHAIN>` is omitted, the default toolchain is used. To see the
available toolchains, use the following:

```bash
corteca config get toolchain.targets
```

## Configuration

Configuration values are read cascadingly from the following sources:

1. System-wide configuration: `/etc/corteca/corteca.json`
1. User global configuration: `$HOME/.config/corteca/corteca.json`
1. user local (project) configuration: `.corteca.json`

The last is searched in the working directory where `corteca` binary is executed; if
not found, it will continue searching in the parent folder(s), until it finds a
configuration file or reaches filesystem root. Thus, provided that a local
configuration file exists in the project root folder, `corteca` commands can be run from
any subfolder inside that.
