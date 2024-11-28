# `exec`

## Usage

The exec command allows you to execute a predefined sequence on a specified device. This sequence can include various deployment, configuration, or other operational steps for remote devices (see [sequence configuration](#configuration)).

```shell
corteca exec NAMED-SEQUENCE DEVICE
```

The following parameters are supported:

* `NAMED-SEQUENCE` is a mandatory parameter that indicates the sequence that will be executed.

* `DEVICE` is a mandatory parameter that indicates where the sequence will be executed.

### Flags

```text
  -a, --artifact string    Specify an artifact in the form of 'architecture:imagetype:/path/to/file', architecture=(aarch64|armv7l|x86_64), imagetype=(rootfs|oci)s
  --global       boolean   Affect global config & ignore any project-local configuration
  --publish      string    Publish application artifact to specified target
  --ssh-log      string    Specify where SSH logs will be stored (default "/dev/null")
```

### Options inherited from parent commands

```text
  -c, --config stringArray   Override a configuration value in the form of a 'key=value' pair
  -r, --configRoot string    Override configuration root folder (default "/etc/corteca")
  -C, --projectRoot string   Specify project root folder
```

## Configuration

### Sequences

In the `corteca.yaml`, the `sequences` section -related to the exec command- specifies an array of commands to be executed on the specified device:

```yaml
sequences:
    # <alias>:
    #   - cmd: <command-to-run>
    #     delay: <milliseconds-after-cmd>
    #     retries: <attempts>
    #     ignoreFailure: <true/false>
    #     input: <command-input>
    install:
      - delay: 1000
      - cmd: "PluginCli install -u {{.publish.publicURL}}/{{.buildArtifact}} -i
            {{.app.duid}}"
        delay: 1000
      - cmd: "lxc-ls  -P /opt/lxc/NOS/ | grep {{.app.name}}"
        delay: 1000
        retries: 30
    deploy:
      - cmd: $(install)
      - cmd: "PluginCli start -i {{.app.duid}}"
        delay: 20000
      - cmd: "lxc-ls -f -P /opt/lxc/NOS/ | grep -o {{.app.name}}[[:space:]]*RUNNING |
            sed 's/ //g'"
        delay: 1000
        retries: 10
```

### Devices

In the `corteca.yaml`, the `devices` section -related to the exec command- specifies devices to be used for deployment purposes:

```yaml
# A set of devices to deploy the application artifact(s)
# each entry must be in the form of:
#   <alias>:
#       addr: <url>                         # url of device console (only `ssh` protocol is currently supported)
#       auth: <type>                        # authentication type; one of `password`, `publicKey`
#       password2: <password2>              # password2 of device
#       privateKeyFile: <path/to/file/name> # path to keyfile for private key authentication
#       token: <device-token>               # authentication token
devices:
    # beacon device
    beacon:
        addr: ssh://{{ getEnv "BEACON_USER" }}:{{ getEnv "BEACON_PASS" }}@192.168.18.1
        password2: "{{ getEnv \"PASSWORD2\" }}"
    # a qemu-based beacon virtualization
    qemu:
        addr: ssh://{{ getEnv "vBEACON_USER" }}@172.17.0.2:{{ getEnv "vBEACON_PORT" }}

```

## Example

### Within a project

```sh
corteca exec install qemu
```

### Combining `corteca publish`

```sh
corteca exec foo beacon --publish localRegistry
```
