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
corteca create <appdir>
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

## Writing templates

A template is collection of files that are rendered using Golang's [text/template](https://pkg.go.dev/text/template) package. `corteca` scans for available templates in a `templates` folder that resides in the global config folder (defaults to `$HOME/.config/corteca`). You can override the global config folder using the `--configRoot` flag.

Each subfolder containing a `.template-info.yaml` file will be treated as a template and will be rendered with all configuration variables available for rendering. For more information, see [Configuration](#configuration) below.

### `.template-info.yaml`

Here is an overview of the file's contents:

```yaml
  # name of the template
  name: "c"
  description: "Short template description"
  # a collection of custom options available for rendering
  options:
    # this will be available inside template context
    - name: "include_libhlapi"
      # text used to prompt the user
      description: "Include hlapi C lib"
      # One of: "boolean", "text" or "choice"
      type: "boolean"
      # true/false for boolean type, any text value for "text" or "choice"
      default: true
      # available options for "choices" type
      # values:
      #    - option1
```

### Rendering

The template is rendered using the following rules:

1. Starting from the parent folder containing the `.template-info.yaml` file, a list of full paths to every regular file in the folder structure is produced
1. Each entry in the list is rendered by the template engine into a path string; this allowes template variable usage in both file and directory names
1. A file is created in the destination folder using the full relative (rendered) path and is filled with the content of the original rendered template file.
1. If any path element is rendered into an empty string, this path entry is skipped
from the rest of the sequence; this allows for conditional generation of folder sub-structure. For example, consider the following structure:

```text
    {{.app.name}}
        ├── {{.app.name}}.c
        ├── {{if .app.options.include_libhlapi}}libs{{end}}
        │   └── lib.h
        └── {{if .app.options.include_libhlapi}}libs.inc{{end}}

    If the `.app.options.use_libhlapi` evaluates to false, the whole filename template will be rendered empty; thus both the `libs` subfolder (and all its containing items), as well as the `libs.inc` regular file will be omitted from the resulting file structure.
```

### Configuration

Configuration values are read cascadingly from the following sources:

1. System-wide configuration: `/etc/corteca/corteca.yaml`
1. User global configuration: `$HOME/.config/corteca/corteca.yaml`
1. user local (project) configuration: `.corteca.yaml`

The last is searched in the working directory where `corteca` binary is executed; if
not found, it will continue searching in the parent folder(s), until it finds a
configuration file or reaches filesystem root. Thus, provided that a local
configuration file exists in the project root folder, `corteca` commands can be run from
any subfolder inside that.
