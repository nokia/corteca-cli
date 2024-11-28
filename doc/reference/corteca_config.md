# `config`

The config command group allows users to read and write configuration values within Corteca applications, applicable to both project-specific (local) and global configurations.

Configuration values are derived from `corteca.yaml`, a central file that defines all settings and parameters. Users can view and modify these fields using the commands in this group. For a detailed view of all available configuration options, see [corteca.yaml](../../data/corteca.yaml).

## Usage

```sh
corteca config [subcommand]
```

### Available Subcommands

- [`add`](./corteca_config_add.md)
- [`get`](./corteca_config_get.md)
- [`set`](./corteca_config_set.md)
