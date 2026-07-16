# Application Templates

`corteca` scans for available templates in a `templates/` folder that resides in the global config folder (defaults to `$HOME/.config/corteca`). You can override the global config folder using the `--configRoot` flag.

Each subfolder containing a `.template-info.yaml` file will be treated as a template and will be rendered with all configuration variables available for rendering. For more information, see the configuration section above.

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
