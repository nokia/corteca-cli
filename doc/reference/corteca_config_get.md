# `config get`

Retrieve the value of a specified configuration key from [corteca.yaml](../../data/corteca.yaml).

## Usage

```sh
corteca config get KEY
```

**key** (optional): The configuration key to retrieve. If omitted, all configurations are displayed.

### Flags

```text
--global    boolean     Retrieve from the global configuration instead of project-local
```

**Note**: When running the command outside of a project scope, the command *gets* you global configuration information.

### Example

Using the [hello-world application](./corteca_create.md#example), you can se below how we can retrieve project-local or global configuration values.

#### For a specific key

```sh
corteca config get app.runtime.hostname

```

Output:

```yaml
hello_world
```

#### For a whole configuration section

```sh
corteca config get app
```

Output:

```yaml
lang: go
name: hello_world
author: Foo
description: My hello world application
version: 0.0.1
fqdn: hello_world.foo.domain
duid: 126decdd-acd4-5229-a8e1-3da5d6b927ae
options: {}
dependencies:
    compile:
        - go
    runtime:
        - json-c
env: {}
entrypoint: /bin/hello_world
runtime:
    ociVersion: 1.0.0
    hostname: hello_world
    mounts:
        - destination: /opt
          type: bind
          source: /var/run/ubus-session
          options:
            - rbind
            - rw
    hooks:
        prestart:
            - path: /bin/prepare_container.sh
        poststop:
            - path: /bin/cleanup_container.sh
    linux:
        resources:
            memory:
                limit: 15728640
                reservation: 15728640
                swap: 31457280
            cpu:
                shares: 1024
                quota: 5
                period: 100
```

#### For global values

```sh
corteca config get --global app
```

Output:

```yaml
lang: ""
name: ""
author: ""
description: ""
version: ""
fqdn: ""
duid: ""
options: {}
dependencies:
    compile: []
    runtime: []
env: {}
entrypoint: ""
runtime:
    ociVersion: ""
    hostname: ""
```
