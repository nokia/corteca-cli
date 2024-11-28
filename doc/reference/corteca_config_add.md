# `config add`

Append a value to a configuration key or add a new one in [corteca.yaml](../../data/corteca.yaml).

## Usage

```sh
corteca config add KEY VALUE
```

**key** (mandatory): The configuration key to append to.
**value** (mandatory): The value to append.

### Flags

```text
--no-regen  boolean     Skip template regeneration after appending the value.
--global    boolean     Affect the global configuration instead of the project-local configuration.
```

**Note**: When running the command outside of a project scope, the command *adds* to your global configuration.

### Example

Using the [hello-world application](./corteca_create.md#example), you can se below how we can add project-local or global configuration values.

#### For a specific key

```sh
corteca config add app.runtime.hostname foo

```

#### For a whole configuration section

```sh
corteca config add devices "foo: {addr: foo}"
```

Output:
The above command adds this section to the project's corteca.yaml

```yaml
devices:
    foo:
        addr: foo
```
