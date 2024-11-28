# `config set`

Set a specific configuration value in [corteca.yaml](../../data/corteca.yaml).

## Usage

```sh
corteca config set KEY VALUE
```

**key** (mandatory): The configuration key to set.
**value** (mandatory): The value to set for the specified key.

### Flags

```text
--no-regen  boolean     Skip template regeneration after setting the value.
--global    boolean     Affect the global configuration instead of the project-local configuration.
```

**Note**: When running the command outside of a project scope, the command *sets* to your global configuration.

### Example

Using the [hello-world application](./corteca_create.md#example), you can se below how we can set project-local or global configuration values.

#### For a specific key

```sh
corteca config set app.runtime.hostname foo

```
