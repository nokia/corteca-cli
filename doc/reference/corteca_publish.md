# `publish`

The publish command is used to publish application artifacts to a specified target, with optional architecture filtering.

## Usage

```sh
publish TARGET [ARCH]
```

**TARGET** (mandatory): The target to which the artifact should be published. This can be one of the predefined publish targets (see [publish configuration](#configuration)).

**ARCH** (optional): The architecture of the build artifact. If not specified, the command uses the architecture from the current build context.

### Flags

```text
  -a, --artifact string    Specify an artifact in the form of 'architecture:imagetype:/path/to/file', architecture=(aarch64|armv7l|x86_64), imagetype=(rootfs|oci)s
  --global       boolean   Affect global config & ignore any project-local configuration
```

## Configuration

In the `corteca.yaml`, the `publish` section -related to the publish command- specifies publish targets and their authorization/publish methods:

```yaml
publish:
    local:
        addr: http://0.0.0.0:8080
        method: listen
        publicURL: http://172.17.0.1:8080

    #   webserver:
    #       addr: https://upload.example.com/artifacts/
    #       auth: basic
    #       method: put

    localRegistry:
        addr: http://0.0.0.0:8080
        method: registry-v2

    remoteRegistry:
        addr: https://corteca-registry.int.net.nokia.com
        auth: basic
        method: push
```

## Example

### Within a project

```sh
corteca publish localRegistry armv7l
```

### Specifying the artifact path

```sh
corteca publish localRegistry --artifact armv7l:oci:/path/to/artifact
```
