# `regen`

The regen command regenerates template files by reading the source paths specified in the `corteca.yaml` configuration file and rendering the output to the corresponding destination paths, also defined in the same file. This ensures that all templates are up-to-date and correctly rendered according to the latest configurations.

## Usage

To regenerate all templates **within a project**, run:

```bash
corteca regen
```

### Options inherited from parent commands

```text
  -c, --config stringArray   Override a configuration value in the form of a 'key=value' pair
  -r, --configRoot string    Override configuration root folder (default "/etc/corteca")
  -C, --projectRoot string   Specify project root folder
```

This command processes all templates defined in the `corteca.yaml` file and outputs the rendered files to their destination paths.

## Configuration

In the `corteca.yaml`, the `templates` section -related to the regen command- specifies a mapping of source template files to their corresponding destination output files:

```yaml
templates:
  /path/to/source/template: /path/to/destination/file
```

## Integration in Workflow

The regen command is automatically executed as part of the following operations:

- Prior to every `corteca build`.
- After every `corteca config set/add`.

This ensures that the templates remain consistent with any new configuration or change. If you wish to bypass template regeneration in these operations, use the `--no-regen` flag.

### Example

To modify the app configuration without triggering template regeneration:

```bash
corteca config set app.name fooName --no-regen
```

This allows you to update configuration settings without regenerating the templates.
