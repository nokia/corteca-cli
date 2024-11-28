# Corteca Command Line Interface

- [Corteca Command Line Interface](#corteca-command-line-interface)
  - [Available commands](#available-commands)
  - [Configuration](#configuration)
  - [Application Templates](#application-templates)

## Available commands

- [`corteca build`](reference/corteca_build.md)
- [`corteca config`](reference/corteca_config.md)
   - [`corteca config add`](reference/corteca_config_add.md)
   - [`corteca config get`](reference/corteca_config_get.md)
   - [`corteca config set`](reference/corteca_config_set.md)
- [`corteca create`](reference/corteca_create.md)
- [`corteca exec`](reference/corteca_exec.md)
- [`corteca publish`](reference/corteca_publish.md)
- [`corteca regen`](reference/corteca_regen.md)

**Note**: For essential guidance on using the Corteca CLI, refer to `corteca --help` for general information, or `corteca <cmd> --help` for details on specific commands.

## Configuration

Configuration values are read cascadingly from the following sources:

1. System-wide configuration: `/etc/corteca/corteca.yaml`
1. User global configuration: `$HOME/.config/corteca/corteca.yaml`
1. user local (project) configuration: `./corteca.yaml`

The last is searched in the working directory; if not found, it will continue searching in the parent folder(s), until it finds a configuration file or reaches filesystem root. Thus, provided that a local configuration file exists in the project root folder, `corteca` commands can be run from any subfolder inside that.

## Application Templates

`corteca` scans for available templates in a `templates/` folder that resides in the global config folder (defaults to `$HOME/.config/corteca`). You can override the global config folder using the `--configRoot` flag.

Each subfolder containing a `.template-info.yaml` file will be treated as a template and will be rendered with all configuration variables available for rendering. For more information, see [configuration section](#configuration) above.

Here is an overview of the content of `.template-info.yaml`

```yaml
    # Name of the template
    name: "foobar"
    description: "Short template description"
    dependencies:
        compile: [ foo ] # Dependencies needed during compile
        runtime: [ bar ] # Dependencies needed during runtime
    # Dynamic files that need regeneration during the project's lifecycle
    regenFiles:
        .corteca/foo.template: foo # Mapping template filepath to regenerated filepath
    # A collection of custom options available for rendering
    options:
```
