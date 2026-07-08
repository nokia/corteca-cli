<div align="center">
  <img src="./doc/images/nokia_logo_blue.svg" width="250" alt="Nokia Logo" title="Nokia Logo" />
</div>

# Corteca Developer Toolkit

[![Build](https://github.com/nokia/corteca-cli/actions/workflows/build.yml/badge.svg)](https://github.com/nokia/corteca-cli/actions/workflows/build.yml)
[![Tests](https://github.com/nokia/corteca-cli/actions/workflows/test.yml/badge.svg)](https://github.com/nokia/corteca-cli/actions/workflows/test.yml)
[![Coverage](https://codecov.io/gh/nokia/corteca-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/nokia/corteca-cli)

Corteca Developer Toolkit is a command-line tool for building, packaging, and
deploying container applications to Nokia broadband devices. It covers the full
application lifecycle — from scaffolding a new project to publishing the build
artifact and running deployment sequences on a live device.

## Features

- **Project scaffolding** — bootstrap a new application project from a
  language template. Default templates available for C, C++ and Go, but
  virtually any technology stack can be integrated.

- **Cross-architecture builds** — build and package the application producing
  OCI, Docker, or plain rootfs output.

- **Flexible publishing** — publish the build artifact via a local HTTP(S)
  server, HTTP PUT upload, OCI registry push, or a locally-hosted Docker
  Distribution registry.

- **Device deployment** — run named sequences of deployment steps on a remote
  device over SSH or CWMP (TR-069).

- **Configuration management** — inspect and modify any setting through the
  `corteca config` command; most fields support template expressions that are
  evaluated at runtime against the current command context.

- **Template-driven file generation** — keep generated project files (such as
  `Dockerfile`) in sync with application settings using `corteca regen`.

## What is a Corteca application?

A Corteca application is a container image (OCI or Docker format) designed to
run on the managed execution environment of a Nokia broadband device. Each
application has a unique identifier (DUID), a target architecture, and a
self-contained runtime that is isolated from the host firmware. Applications
are described by a `corteca.yaml` manifest that captures everything needed to
build, publish, and deploy them: source dependencies, build options, publish
targets, and deployment sequences.

Corteca supports three target architectures out of the box — `aarch64`,
`armv7l`, and `x86_64` — and can produce OCI images, Docker images, or Nokia
rootfs archives depending on the target device.

## What is a Corteca device?

A Corteca device is any Nokia broadband device that Corteca can connect to in
order to deploy and manage applications. Two connectivity protocols are
supported:

- **SSH** — for devices that expose a shell, such as development boards or
  devices running [prplOS](https://prplos.eu). Corteca opens an SSH session
  and runs a user-defined sequence of shell commands on the device.
- **CWMP (TR-069 / TR-369)** — for carrier-grade CPE managed via the
  [TR-069](https://www.broadband-forum.org/technical/download/TR-069.pdf)
  protocol. Corteca acts as an ACS: it starts a local HTTP(S) listener, sends
  a connection request to the CPE, and drives the session by issuing RPCs such
  as `ChangeDUState` (install/remove a Deployment Unit) and
  `SetParameterValues` (configure or start an Execution Unit).

Devices and the sequences to run on them are configured in `corteca.yaml` and
can be targeted by name when running `corteca exec`.

<hr>

## Prerequisites

| Requirement | Version  | Notes                                          |
| ----------- | -------- | ---------------------------------------------- |
| Go          | ≥ 1.21   | Required to build from source                  |
| Docker      | ≥ 23.0   | Required to build application container images |
| Docker BuildKit | ≥ 0.11 | Required for `docker build --output`         |
| make        | any      | Used to drive the build and install targets    |

## Build

### Building Locally

To build locally, you need to have a Go toolchain installed. Clone the repository and run:

```bash
$ make
```

The compiled binary is placed in the `dist/` directory. You can also build packages for your specific platform. The available `make` targets are `deb`, `rpm`, `osx` and `msix`:

```bash
$ make rpm
```

### Build using Docker

If you do not have a local Go toolchain, you can build entirely inside Docker (BuildKit is required). The following command builds all supported packages:

```bash
$ docker build --output ./dist .
```

#### Selecting the Packages to Build

By default, packages are created for all architectures (`arm64` and `amd64`). You can select the architectires to build for by setting the `ARCH` build-time variable to the space-separated list of architectures to build for. For example, to only build packages for the `amd64` architecture run:

```bash
$ docker build --build-arg ARCH="amd64" --output ./dist .
```

The package types created by default are `deb`, `rpm`, `osx` and `msix`. To only build selected package types, set the `PACKAGE` build-time variable to the space-separated list of package types to build. The following command only builds `deb` and `rpm` packages on all architectures:

```bash
$ docker build --build-arg PACKAGE="deb rpm" --output ./dist .
```

By setting both the `ARCH` and `PACKAGE` build-time variables, you can further limit the build scope to specific packages. For example, to only build the `deb` package for the `amd64` architecture, run:

```bash
$ docker build --build-arg ARCH="amd64" --build-arg PACKAGE="deb" --output ./dist .
```

#### Installing Docker BuildKit

On Docker Engine < 23.0, BuildKit must be enabled manually. On Ubuntu 22.04:

```bash
$ sudo apt-get install docker-buildx-plugin
```

Then either prefix each `docker build` invocation with `DOCKER_BUILDKIT=1`, or
follow the [official instructions](https://docs.docker.com/build/buildkit/#getting-started)
to enable it globally.

## Install

### Install manually from source

You can run the below command to build the application and install the binary to
the `/usr/bin` folder and the default configuration files to the `/etc/corteca`
folder:

```bash
$ sudo make install
```

To remove a previous manual installation

```bash
$ sudo make uninstall
```

You can customize the destination folder by overriding the `$DESTDIR`
environment variable:

```bash
# no sudo required; will be installed to ~/.local/share/usr/bin
$ DESTDIR=~/.local/share make install
```

### Install with package manager

If you are using debian/ubuntu or redhat-based distributions, you can create a relevant package and let your package manager handle installation. E.g. for ubuntu:

```bash
$ make deb
$ make rpm
```

## Getting Started

The fastest way to get up and running is to create a project, build it, and
publish it to a local registry in three commands:

```bash
$ corteca create my-app          # scaffold a new application project
$ cd my-app
$ corteca build aarch64          # build an OCI image for aarch64
$ corteca publish localRegistry  # push it to a local OCI registry
```

For a step-by-step walkthrough — including how to configure a device target and
run a deployment sequence — see **[doc/GettingStarted.md](doc/GettingStarted.md)**.

## Configuration

All Corteca settings live in `corteca.yaml`. Configuration is read
cascadingly from three locations:

| Precedence | Location                            | Scope            |
| ---------- | ----------------------------------- | ---------------- |
| Lowest     | `/etc/corteca/corteca.yaml`         | System-wide      |
|            | `$HOME/.config/corteca/corteca.yaml`| User global      |
| Highest    | `./corteca.yaml` (project root)     | Per-project      |

The project-level file is found by walking up from the current working
directory, so `corteca` commands work from any subdirectory of a project.

The `corteca config` command can be used to inspect or modify any value
without editing YAML by hand:

```bash
corteca config get publish          # show all publish targets
corteca config set app.version 1.1  # update a value
```

For a full reference of every configuration key, their types, defaults, and
supported template expressions, see the [configuration
reference](doc/Configuration.md).

## Command Line Reference

| Command | Description |
| ------- | ----------- |
| [`corteca create`](doc/reference/corteca_create.md) | Scaffold a new application project from a template |
| [`corteca build`](doc/reference/corteca_build.md) | Build and package the application for a target architecture |
| [`corteca publish`](doc/reference/corteca_publish.md) | Upload or serve the build artifact via a configured publish target |
| [`corteca exec`](doc/reference/corteca_exec.md) | Run a named deployment sequence on a configured device |
| [`corteca config`](doc/reference/corteca_config.md) | Inspect or modify configuration values |
| [`corteca regen`](doc/reference/corteca_regen.md) | Regenerate template-derived project files |

For a broader overview of all commands, flags, and usage patterns see
[doc/USAGE.md](doc/USAGE.md).

Every command also accepts `--help` for inline usage information:

```bash
corteca --help
corteca build --help
```
