# `create`

The create command is used to generate a new application skeleton/template based on a selected programming language (see details about templates [here](../USAGE.md#application-templates)). This command can be configured interactively or through specified flags.

## Usage

```sh
corteca create DESTFOLDER
```

**DESTFOLDER**: The directory where the application skeleton will be created.

### Flags

```text
--skipPrompts   Bypasses interactive prompts for settings that are not provided through the command line -ideal for automation.
```

### Example

```sh
corteca create hello_world      # This open an interactive dialog for you:
```

If the command runs successfully the output should be something like this:

![corteca create prompt and output](../corteca-create.PNG?raw=true)

## Configuration

In the `corteca.yaml`, the `app` section -related to the create command- specifies app settings you can alter during the build process.

```yaml
# According to the values we put in the above example, the corteca.yaml of the hello_world application -among other fields- should have these:

app:
    author: Foo
    description: My hello world application
    duid: 126decdd-acd4-5229-a8e1-3da5d6b927ae
    fqdn: hello_world.foo.domain
    lang: go
    name: hello_world
    title: hello world
    version: 0.0.1
```
