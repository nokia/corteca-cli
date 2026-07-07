# Configuration

Corteca reads its configuration from a YAML file named `corteca.yaml`. Settings are applied in a **cascading** manner: each layer is merged on top of the previous one, with later layers taking precedence over earlier ones.

## Configuration Locations

Configuration is loaded in the following order (later entry values take higher priority):

1. System-wide:
    - Linux: `/etc/corteca`
    - macOS: `/usr/local/etc/corteca`
    - Windows: `%PROGRAMDATA%\Corteca`
2. User-level:
    - Linux/macOS: `${HOME}/.config/corteca`
    - Windows: `%APPDATA%\Corteca`
3. Project-level:
    - The nearest file named `corteca.yaml` found by walking up the directory tree from the current working directory; the folder containing this is the `projectRoot` folder.
4. CLI overrides:
    - Values passed via the `-c` / `--config` flag as `key=value` pairs

The user-level layer is optional; if its directory does not exist it is silently skipped. The project-level layer is also optional and is not loaded when commands are invoked with the `--no-local-config` flag or when no `corteca.yaml` is found in any ancestor directory.

You can override the system configuration root with the `-r` / `--configRoot` flag, and specify an explicit project root with the `-C` / `--projectRoot` flag.

---

## Template Expressions

Many string fields across all sections support **template expressions** — dynamic values resolved at runtime from the current command context. See [Annex A: Template Context Fields](#annex-a-template-context-fields) for the full list of available fields.

The general syntax is:

```bash
${ ["<prefix>":].<field.path>[:<kv-sep>[:<entry-sep>]] }
```

For details and examples of the expression syntax, refer to the annex.

---

## Top-level Sections

The `corteca.yaml` file is divided into the following top-level sections:

```yaml
app:       # Application metadata and runtime settings
build:     # Build pipeline configuration
publish:   # Named publish targets (registries, servers, …)
devices:   # Named remote devices for deployment
sequences: # Named command sequences executed on devices
templates: # Template-file-to-destination mappings
```

---

## `app`; application properties and runtime config

The `app` section describes the application's identity and runtime
requirements. Most fields in this section are project-specific and are
generated or filled in during `corteca create`.

```yaml
app:
    name: <app-name>       # Single-word identifier; used for artifact naming
    author: <author>       # Author metadata (used in Corteca cloud)
    version: <version>     # Semantic version string (used in Corteca cloud)
    duid: <uuid>           # Device-unique ID; auto-generated from FQDN
    dependencies:
        compile:           # Packages required at compile time
            - <package1>
            - <package1>
        runtime:           # Packages required at runtime
            - <package1>
            - <package2>
    env:                   # Environment variables injected into the app container
        <NAME>: <value>
    entrypoint:            # Command to run when the container starts
        - /bin/<app-name>
    runtime: {}            # OCI runtime spec overrides (see OCI Runtime Specification)
```

### Fields

| Field                  | Type            | Description                                                                                                                                                 |
| -------                | ------          | -------------                                                                                                                                               |
| `name`                 | string          | Single-word application identifier. Cannot contain spaces. Used for artifact file naming.                                                                   |
| `author`               | string          | Name of the application author. Used as metadata in Corteca cloud registrations.                                                                            |
| `version`              | string          | Application version string. Used for artifact naming and cloud metadata.                                                                                    |
| `duid`                 | string          | Deterministic UUID derived from the application's FQDN. Auto-generated; do not edit manually.                                                               |
| `dependencies.compile` | list of strings | System packages required during the build step.                                                                                                             |
| `dependencies.runtime` | list of strings | System packages that must be present inside the container at runtime.                                                                                       |
| `env`                  | map             | Key-value pairs injected as environment variables into the running container.                                                                               |
| `entrypoint`           | list of strings | The command (and optional arguments) that serves as the container entry point. Defaults to `/bin/<name>` if not set.                                        |
| `runtime`              | object          | OCI runtime specification overrides (e.g., resource limits, mounts, hooks). Follows the [OCI Runtime Spec](https://github.com/opencontainers/runtime-spec). |

---

## `build`; build and packaging settings

The `build` section controls how the application is compiled and packaged. It
defines the target architectures, build output format, and cross-compilation
settings. This section is tightly coupled with the `corteca build` command.

```yaml
build:
    architectures:
        <arch-name>:
            platform: <docker-platform>   # e.g. "linux/arm/v7"
    default: <arch-name>                  # Architecture used when none is specified
    options:
        outputType: <type>                # One of: rootfs | oci | docker
        debug: <bool>                     # Include debug tooling in the container
        skipHostEnv: <bool>               # Do not inherit host environment variables
        env:
            <NAME>: <value>               # Additional build-time environment variables
    crossCompile:
        enabled: <bool>                   # Enable QEMU-based cross-compilation
        image: <docker-image>             # QEMU helper image
        args: [ <arg>, … ]               # Arguments passed to the QEMU helper
```

### Build Fields

| Field                           | Type            | Default                      | Description                                                                                                                                 |
| -------                         | ------          | ---------                    | -------------                                                                                                                               |
| `architectures`                 | map             | —                            | Named map of supported target architectures. Built-in entries: `armv7l`, `aarch64`, `x86_64`.                                               |
| `architectures.<name>.platform` | string          | —                            | Docker platform string passed to `docker buildx` (e.g., `linux/arm/v7`).                                                                    |
| `default`                       | string          | `aarch64`                    | Default architecture selected when none is provided on the command line.                                                                    |
| `options.outputType`            | string          | `oci`                        | Build output format. `rootfs` produces a compressed root filesystem; `oci` produces an OCI image; `docker` produces a Docker image tarball. |
| `options.debug`                 | bool            | `false`                      | When `true`, additional debugging tools are included in the generated container image.                                                      |
| `options.skipHostEnv`           | bool            | `false`                      | When `true`, host environment variables (proxies, etc.) are not forwarded into the build container.                                         |
| `options.env`                   | map             | —                            | Extra environment variables available inside the build container.                                                                           |
| `crossCompile.enabled`          | bool            | `true`                       | Enable cross-compilation via QEMU user-mode emulation.                                                                                      |
| `crossCompile.image`            | string          | `multiarch/qemu-user-static` | Docker image used to register QEMU binfmt handlers.                                                                                         |
| `crossCompile.args`             | list of strings | `["--reset", "-p", "yes"]`   | Arguments forwarded to the QEMU helper image on startup.                                                                                    |

---

## `publish`; publish container artifacts

The `publish` section defines a named map of publish targets. Each entry must
include a `method` field that selects the delivery mechanism. **The remaining
fields are method-dependent** — each method decodes the raw YAML entry into a
different endpoint configuration struct, so only the fields documented for that
method are meaningful. This section is tightly coupled with the
`corteca publish` command.

The only field shared by all methods is:

| Field    | Type   | Description                                                                          |
| -------  | ------ | -------------                                                                        |
| `method` | string | **(Required)** Delivery mechanism. One of `listen`, `put`, `push`, or `registry-v2`. |

---

### Method: `listen`

Corteca binds a local HTTP file server to `addr` and serves the build artifact
from the `<projectRoot>/dist/` directory.

```yaml
publish:
    <alias>:
        method: listen
        addr: http://0.0.0.0:8080       # Address to bind the HTTP server to
        certificate: <path/to/cert.pem> # TLS certificate (optional, for HTTPS)
        key: <path/to/key.pem>          # TLS private key (optional, for HTTPS)
```

| Field         | Type              | Required   | Description                                                                        |
| -------       | ------            | ---------- | -------------                                                                      |
| `addr`        | string (template) | Yes        | Host and port for the HTTP server to listen on.                                    |
| `certificate` | string (template) | No         | Path to a PEM-encoded TLS certificate. Enables HTTPS when set together with `key`. |
| `key`         | string (template) | No         | Path to a PEM-encoded TLS private key.                                             |

---

### Method: `put`

Uploads the build artifact to a remote HTTP/HTTPS server using the PUT verb.
Credentials are resolved from config or prompted interactively if absent.

```yaml
publish:
    <alias>:
        method: put
        addr: https://upload.example.com/artifacts/   # Target URL
        auth: basic                                    # One of: basic | bearer | digest
        username: <username>                           # Template expressions supported
        password: <password>                           # Template expressions supported
        token: <bearer-token>                          # Used when auth is bearer
        skipTLSVerification: false                     # Skip TLS certificate verification
```

| Field                 | Type              | Required   | Description                                                                                                                                      |
| -------               | ------            | ---------- | -------------                                                                                                                                    |
| `addr`                | string (template) | Yes        | Destination URL. Must use `http://` or `https://` scheme. The artifact filename is appended to the path automatically.                           |
| `auth`                | string            | No         | Authentication scheme. One of `basic`, `bearer`, or `digest`. When omitted, bearer auth is tried first if a `token` is present, otherwise basic. |
| `username`            | string (template) | No         | Username for `basic` or `digest` auth. Prompted interactively if absent and required.                                                            |
| `password`            | string (template) | No         | Password for `basic` or `digest` auth. Prompted interactively if absent and required.                                                            |
| `token`               | string (template) | No         | Bearer token for `bearer` auth. Required when `auth: bearer`.                                                                                    |
| `skipTLSVerification` | bool              | No         | When `true`, skips TLS certificate verification. Useful for self-signed certificates. Defaults to `false`.                                       |

> **Note:** `digest` authentication is not yet implemented and will return an error at runtime.

---

### Method: `push`

Pushes the build artifact as an OCI image to a remote container registry
following the [Docker Distribution (Registry V2)](https://distribution.github.io/distribution/)
protocol. Uses the same configuration fields as `put`. Keep in mind that the
`addr` field will be interpreted as: `registry/<repository>:<reference>`

```yaml
publish:
    <alias>:
        method: push
        addr: https://registry.example.com/aarch64/${.app.name}:${.app.version}
        auth: basic
        username: <username>
        password: <password>
        token: <bearer-token>
        skipTLSVerification: false                     # Skip TLS certificate verification
```

| Field                 | Type              | Required   | Description                                                                                                |
| -------               | ------            | ---------- | -------------                                                                                              |
| `addr`                | string (template) | Yes        | Full registry URL including repository and tag. Template expressions are commonly used here.               |
| `auth`                | string            | No         | Authentication scheme. Either `basic` or `bearer`.                                                         |
| `username`            | string (template) | No         | Registry username. Prompted interactively if absent and `auth: basic`.                                     |
| `password`            | string (template) | No         | Registry password. Prompted interactively if absent and `auth: basic`.                                     |
| `token`               | string (template) | No         | Bearer token for `bearer` auth.                                                                            |
| `skipTLSVerification` | bool              | No         | When `true`, skips TLS certificate verification. Useful for self-signed certificates. Defaults to `false`. |

---

### Method: `registry-v2`

Spins up a local [Docker Distribution (Registry
V2)](https://distribution.github.io/distribution/) server, pushes the artifact
to it and keeps it running so devices can pull from it. This publish method
should be used for development/test purposes only and **should never be used in
production**.

```yaml
publish:
    <alias>:
        method: registry-v2
        addr: 0.0.0.0:8080              # Address for the local registry to bind to
        namespace: <repository-path>    # Repository namespace (supports template expressions)
        reference: <tag>                # Image tag or reference (supports template expressions)
        certificate: <path/to/cert.pem> # TLS certificate (optional, for HTTPS)
        key: <path/to/key.pem>          # TLS private key (optional, for HTTPS)
```

| Field          | Type              | Required   | Description                                                                                                                                      |
| -------------- | ----------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `addr`         | string (template) | Yes        | Address the local registry listens on the form of `<host>[:<port>]` (no schema prefix)                                                           |
| `namespace`    | string (template) | Yes        | Repository path within the registry (e.g., `myapp` or `org/myapp`). Combined with `reference` to form the image path `/<namespace>:<reference>`. |
| `reference`    | string (template) | Yes        | Image tag or digest reference (e.g., `1.0.0` or `latest`). Template expressions such as `${ .app.version }` are commonly used here.              |
| `certificate`  | string (template) | No         | Path to a PEM-encoded TLS certificate. Enables HTTPS when set together with `key`.                                                               |
| `key`          | string (template) | No         | Path to a PEM-encoded TLS private key.                                                                                                           |

---

## `devices`; deployment targets

The `devices` section defines a named map of remote devices. Each entry
describes a device that Corteca can connect to in order to run deployment
sequences. **The device type is inferred from the URL scheme of the `addr`
field** — each scheme decodes the raw YAML entry into a different device
configuration struct, so only the fields documented for that scheme are
meaningful.

The fields shared by all device types are:

| Field | Type | Description |
| --- | --- | --- |
| `addr` | string (template) | **(Required)** Connection URL. Its scheme (`ssh://`, `cwmp://`, `cwmps://`) selects the device type. |
| `architecture` | string | Architecture identifier matching an entry in `build.architectures`. |

---

### Type: `ssh`

Connects to the device over SSH. The `addr` field must use the `ssh://`
scheme. Credentials may be embedded directly in the URL
(`ssh://user:password@host`) or provided via separate fields.

```yaml
devices:
    <alias>:
        addr: ssh://root@192.168.1.1        # SSH connection URL (user and password may be embedded)
        architecture: <arch-name>           # Target architecture of the device
        username: <username>                # Overrides any username in addr
        password: <password>                # SSH password; prompted interactively if absent
        password2: <password>               # Secondary (escalation) password for certain firmware
        privateKeyFile: <path/to/keyfile>   # Path to a PEM-encoded private key
```

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `addr` | string (template) | Yes | SSH connection URL. Must use the `ssh://` scheme. Username and password may be embedded (e.g., `ssh://user:pass@host`). |
| `username` | string (template) | No | SSH username. Overrides any username embedded in `addr`. |
| `password` | string (template) | No | SSH password. Falls back to the password in `addr`, then prompts interactively if absent. |
| `password2` | string (template) | No | Secondary (escalation) password required by certain device firmware (e.g., for Quagga deactivation). |
| `privateKeyFile` | string (template) | No | Path to a PEM-encoded private key file. When provided, public-key authentication is attempted first. |

---

### Type: `cwmp` / `cwmps`

Communicates with the device using the
[TR-069 (CWMP)](https://www.broadband-forum.org/technical/download/TR-069.pdf)
protocol. Corteca first starts a local HTTP(S) server on `server.addr`
(defaulting to port `7547`) to receive the incoming CWMP session, then sends a
connection request to the CPE at `addr` to trigger it.
Use the `cwmp://` scheme for plain HTTP or `cwmps://` for HTTPS (TLS).

```yaml
devices:
    <alias>:
        addr: cwmp://192.168.1.1:7547       # Connection request URL sent to the CPE
        architecture: <arch-name>           # Target architecture of the device
        auth: basic                         # One of: basic | bearer | digest
        username: <username>                # Template expressions supported
        password: <password>                # Template expressions supported
        token: <bearer-token>               # Used when auth is bearer
        skipTLSVerification: false          # Skip TLS certificate verification
        server:
            addr: http://0.0.0.0:7547       # Address for the local CWMP listener to bind to
            certificate: <path/to/cert.pem> # TLS certificate for the local listener (optional)
            key: <path/to/key.pem>          # TLS private key for the local listener (optional)
```

| Field                 | Type              | Required | Description                                                                                                                                       |
| -------               | ------            | -------- | -------------                                                                                                                                     |
| `addr`                | string (template) | Yes      | Connection request URL sent to the CPE. Use `cwmp://` for HTTP or `cwmps://` for HTTPS. Defaults to port `7547` if no port is specified.          |
| `auth`                | string            | No       | Authentication scheme for the connection request. One of `basic`, `bearer`, or `digest`. Defaults to bearer if a `token` is present, else basic.  |
| `username`            | string (template) | No       | Username for `basic` or `digest` auth on the connection request.                                                                                  |
| `password`            | string (template) | No       | Password for `basic` or `digest` auth on the connection request.                                                                                  |
| `token`               | string (template) | No       | Bearer token for `bearer` auth on the connection request.                                                                                         |
| `skipTLSVerification` | bool              | No       | When `true`, skips TLS certificate verification for the connection request. Defaults to `false`.                                                  |
| `server.addr`         | string (template) | No       | Address for the local HTTP(S) server that receives the incoming CWMP session from the CPE.                                                        |
| `server.certificate`  | string (template) | No       | Path to a PEM-encoded TLS certificate for the local server. Enables HTTPS when set together with `server.key`.                                    |
| `server.key`          | string (template) | No       | Path to a PEM-encoded TLS private key for the local server.                                                                                       |

---

## `sequences`; deployment sequences

The `sequences` section defines named lists of steps that Corteca executes on a
remote device. Sequences are referenced by name when running `corteca exec`.
Each step shares a common set of flow-control fields, but **the meaning of
`cmd` depends on the target device type**.

```yaml
sequences:
    <alias>:
        - cmd: <command-or-rpc>      # Shell command (ssh) or RPC name (cwmp); supports template expressions
          delay: <duration>          # Wait after the step completes (e.g. "1s", "500ms")
          timeout: <duration>        # Maximum time to wait for the step to complete
          retries: <uint>            # Number of retry attempts on failure
          ignoreFailure: <bool>      # Continue the sequence even if this step fails
        # ...more steps can follow
```

### Common fields

| Field           | Type              | Default   | Description                                                                                                                  |
| -------         | ------            | --------- | -------------                                                                                                                |
| `cmd`           | string (template) | —         | **(Required)** Command to execute. Interpreted differently per device type; see below.                                       |
| `delay`         | string (duration) | `0`       | Time to wait after the step (or each retry) completes. Refer to [this](https://pkg.go.dev/time#ParseDuration) for syntax.    |
| `timeout`       | string (duration) | 5 minutes | Maximum execution time. If exceeded, the step is considered failed. Same syntax as `delay`.                                  |
| `retries`       | uint              | `0`       | How many additional attempts to make if the step fails.                                                                      |
| `ignoreFailure` | bool              | `false`   | When `true`, a failing step does not abort the sequence.                                                                     |

The `cmd` field also supports two special forms for reuse and substitution:

- **Template expressions** — `${.field}` substitutes a value from the current
  configuration context (e.g., `${.app.duid}`, `${.publish.addr}`).
- **Sequence calls** — `$(sequenceName)` inline-expands another named sequence,
  enabling composition.

---

### SSH sequences

For `ssh` devices, `cmd` is a **shell command string** that is executed on the
remote device via SSH. An optional `params` field provides additional arguments
that are appended to `cmd` before execution.

| Field    | Type                       | Description                                                                     |
| -------  | ------                     | -------------                                                                   |
| `cmd`    | string (template)          | Shell command to run on the device.                                             |
| `params` | list of strings (template) | Optional list of arguments appended to `cmd`, separated by spaces.              |

---

### CWMP sequences

For `cwmp`/`cwmps` devices, `cmd` is a **TR-069 RPC name**. Each step is
translated into the corresponding SOAP RPC and sent to the CPE during the
active CWMP session. Additional fields in the step are decoded as the RPC
payload — the available parameters vary per RPC. For full parameter semantics
refer to the latest
[TR-069 specification](https://www.broadband-forum.org/technical/download/TR-069.pdf).

The following CPE RPCs are currently supported:

| RPC                  | Description                                                                                   |
| -------              | -------------                                                                                 |
| `GetRPCMethods`      | Requests the list of RPC methods supported by the CPE.                                        |
| `GetParameterNames`  | Retrieves the names of parameters (or a sub-tree) exposed by the CPE data model.              |
| `GetParameterValues` | Reads the current values of one or more CPE parameters by name.                               |
| `SetParameterValues` | Writes new values for one or more CPE parameters.                                             |
| `ChangeDUState`      | Instructs the CPE to install, update, or uninstall a Deployment Unit (application container). |

---

## `templates`; generate arbitrary config files from your corteca configuration

The `templates` section maps template source files to their rendered output destinations. Entries are (re)generated in the below cases:

- `corteca regen` -- triggered explicitly
- `corteca build` -- before build is triggered
- `corteca config set` -- after a value has been succesfully update

```yaml
templates:
    <path/to/template/file>: <destination/path/to/rendered/file>
```

Each key is a path to a Go template file (relative to the `.corteca` directory inside the project), and the corresponding value is the output path where the rendered result should be written.

### Templates Fields

| Field          | Description                                                              |
| -------        | -------------                                                            |
| Key (string)   | Relative path to the source Go template file.                            |
| Value (string) | Destination path for the rendered output. Supports template expressions. |

---

## Annex A: Template Context Fields

Template expressions are evaluated against a **command context** that is populated progressively as a command runs. Not all fields are available in every context — for example, `.device.*` is only set when executing a sequence against a device, and `.publish.*` is only set during a publish operation.

### Expression Syntax

The full syntax of a template expression is:

```bash
${ ["<prefix>":].<field.path>[:<kv-sep>[:<entry-sep>]] }
```

| Part | Description |
| ------ | ------------- |
| `"<prefix>":` | Optional literal string prepended to each resolved value. |
| `.<field.path>` | Dot-separated path into the context (e.g., `.app.name`, `.env.HOME`). |
| `:<kv-sep>` | For **map** fields: separator between each key and its value (default: space). |
| `:<entry-sep>` | For **map** fields: separator between entries (default: space). |

Examples:

| Expression | Result |
| ------ | ------------- |
| `${ .app.name }` | Value of `app.name` |
| `${ .app.version }` | Value of `app.version` |
| `${ "v":.app.version }` | `v` prepended to `app.version` (e.g., `v1.2.0`) |
| `${ .env.HOME }` | Value of the `HOME` environment variable |
| `${ .env }` | All env vars expanded as `KEY value KEY value …` |
| `${ "--env ":.build.options.env:= }` | Build env vars as `--env KEY=value --env KEY=value …` |
| `${ "--env ":.build.options.env:=:, }` | Same but comma-separated |

Template fields can reference other template fields; circular references will cause a runtime panic.

---

### `.app` — Application Settings

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.app.name` | string | Application name. |
| `.app.author` | string | Application author. |
| `.app.version` | string | Application version string. |
| `.app.duid` | string | Deterministic unique identifier (UUID). |
| `.app.dependencies.compile` | list | Compile-time dependency packages. |
| `.app.dependencies.runtime` | list | Runtime dependency packages. |
| `.app.env` | map | Environment variables injected into the app container. |
| `.app.entrypoint` | list | Container entry point command and arguments. |

### `.build` — Build Settings

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.build.default` | string | Default target architecture name. |
| `.build.options.outputType` | string | Build output type (`rootfs`, `oci`, or `docker`). |
| `.build.options.debug` | bool | Whether debug mode is enabled. |
| `.build.options.skipHostEnv` | bool | Whether host env vars are suppressed in the build container. |
| `.build.options.env` | map | Extra build-container environment variables. |
| `.build.crossCompile.enabled` | bool | Whether QEMU cross-compilation is enabled. |
| `.build.crossCompile.image` | string | QEMU helper Docker image. |
| `.build.crossCompile.args` | list | Arguments forwarded to the QEMU helper. |

### `.arch` and `.platform` — Current Architecture

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.arch` | string | Name of the target architecture currently being built (e.g., `aarch64`). Set during `build`, `regen`, and `exec`. |
| `.platform` | string | Docker platform string for the current architecture (e.g., `linux/arm64`). Set alongside `.arch`. |

### `.artifact` — Build Artifact

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.artifact` | string | Absolute path to the current build artifact (e.g., `/project/dist/app-1.0-aarch64-oci.tar`). Set during `publish` and `exec`. |

### `.publish` — Active Publish Target

Populated when a publish target is selected (during `publish` or `exec` with a publish target).

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.publish.name` | string | Alias name of the active publish target. |
| `.publish.method` | string | Delivery method of the active publish target. |

### `.device` — Active Device

Populated when a device is selected (during `exec`).

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.device.name` | string | Alias name of the active device. |
| `.device.addr` | string (template) | Connection URL of the device. |
| `.device.architecure` | string | Architecture identifier of the device. |

### `.env` — Host Environment Variables

| Field | Type | Description |
| ------- | ------ | ------------- |
| `.env` | map | All host environment variables at the time the command was invoked. Access individual variables as `.env.<NAME>` (e.g., `.env.HOME`, `.env.PATH`). |

---

## Annex B: Configuration Examples

### Publish targets

#### `listen` — serve the artifact over a local HTTP server

Useful during development to let a device pull the artifact directly from the
build machine.

```yaml
publish:
    dev-server:
        method: listen
        addr: http://0.0.0.0:8080
```

For HTTPS, provide a certificate and key:

```yaml
publish:
    dev-server-tls:
        method: listen
        addr: https://0.0.0.0:8443
        certificate: /etc/corteca/certs/server.pem
        key: /etc/corteca/certs/server.key
```

#### `put` — upload the artifact to a remote HTTP server

Credentials are read from environment variables so they are never stored in the
configuration file:

```yaml
publish:
    artifact-store:
        method: put
        addr: https://artifacts.example.com/uploads/
        auth: basic
        username: "${ .env.UPLOAD_USER }"
        password: "${ .env.UPLOAD_PASS }"
```

#### `push` — push the artifact as an OCI image to a container registry

Template expressions in `addr` let the image path and tag be derived
automatically from the application settings:

```yaml
publish:
    registry:
        method: push
        addr: https://registry.example.com/${ .app.name }:${ .app.version }
        auth: basic
        username: "${ .env.REGISTRY_USER }"
        password: "${ .env.REGISTRY_PASS }"
```

For a registry that uses bearer tokens:

```yaml
publish:
    registry-bearer:
        method: push
        addr: https://registry.example.com/${ .app.name }:${ .app.version }
        auth: bearer
        token: "${ .env.REGISTRY_TOKEN }"
        skipTLSVerification: true
```

#### `registry-v2` — run a local registry and keep it alive for devices to pull from

Intended for development and testing only.

```yaml
publish:
    local-registry:
        method: registry-v2
        addr: 0.0.0.0:5000
        namespace: "${ .app.name }"
        reference: "${ .app.version }"
```

---

### Devices

#### SSH device with inline credentials

Username and password are embedded directly in the `addr` URL:

```yaml
devices:
    my-device:
        addr: ssh://root:secret@192.168.1.100
        architecture: aarch64
```

#### SSH device with credentials from environment variables

No sensitive values are stored in the configuration file:

```yaml
devices:
    my-device:
        addr: "ssh://${ .env.DEVICE_USER }:${ .env.DEVICE_PASS }@192.168.1.100"
        architecture: aarch64
        password2: "${ .env.DEVICE_PASS2 }"
```

#### SSH device with public-key authentication

```yaml
devices:
    my-device:
        addr: ssh://root@192.168.1.100
        architecture: aarch64
        privateKeyFile: ~/.ssh/id_ed25519
```

#### CWMP device with credentials from environment variables

Corteca listens on `server.addr` for the CPE to connect back, and uses the
credentials from the environment to authenticate the initial connection request
sent to the CPE:

```yaml
devices:
    my-cpe:
        addr: "cwmp://${ .env.CPE_USER }:${ .env.CPE_PASS }@192.168.1.1:7547"
        architecture: aarch64
        auth: basic
        server:
            addr: http://0.0.0.0:7547
```

---

### Sequences

#### SSH — install and start a DU on a prplOS device

prplOS exposes DU lifecycle management through `ubus`. The install step
downloads the artifact from the active publish target, and the start step
brings the container up:

```yaml
sequences:
    install-prplos:
        - cmd: ubus
          params:
              - call
              - DU.Manager
              - Download
              - '{"URL":"${ .publish.addr }","UUID":"${ .app.duid }","ExecutionEnvRef":"EE.1"}'
          timeout: 2m
          retries: 2

    start-prplos:
        - cmd: ubus
          params:
              - call
              - EE.1.DU.myapp-1.0.0
              - Start
          timeout: 30s
```

#### CWMP — read the device serial number

Uses `GetParameterValues` to retrieve the serial number from the standard
TR-069 `Device.DeviceInfo` object:

```yaml
sequences:
    read-serial:
        - cmd: GetParameterValues
          ParameterNames:
              - Device.DeviceInfo.SerialNumber
```

#### CWMP — install a DU using `ChangeDUState`

The artifact URL is resolved from the active publish target at runtime. The
operation type is selected with a YAML tag (`!InstallOpStruct`,
`!UpdateOpStruct`, or `!UninstallOpStruct`):

```yaml
sequences:
    install-cwmp:
        - cmd: ChangeDUState
          CommandKey: "${ .app.name }-install"
          Operations:
              - !InstallOpStruct
                URL: "${ .publish.addr }"
                UUID: "${ .app.duid }"
                ExecutionEnvRef: SoftwareModules.ExecEnv.1
          timeout: 5m
```

#### CWMP — start a DU using `SetParameterValues`

Sets the `RequestedState` parameter of the Execution Unit to `Active`:

```yaml
sequences:
    start-cwmp:
        - cmd: SetParameterValues
          ParameterKey: "${ .app.name }-start"
          ParameterList:
              - Name: Device.SoftwareModules.ExecEnv.1.ExecUnit.1.RequestedState
                Type: xsd:string
                Value: Active
          timeout: 30s
```
