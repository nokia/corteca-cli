# `build`

The build command builds the application using a specified or default build toolchain. This command is versatile, allowing either a single platform or multiple platforms to be targeted based on the toolchain configurations in corteca.yaml. Below is an overview of its usage, flags, and functionality.

## Usage

### Running the Command

To build for a specific architecture **within a project**, run:

```sh
corteca build TOOLCHAIN
```

**TOOLCHAIN** (optional): Specifies the toolchain architecture to use for the build. If omitted, the default toolchain (aarch64) is used.

### Flags

```yaml
--all          Build the application for all available platforms specified in the configuration. Use this for multi-platform support.
--no-regen     Skip the regeneration of template files before building
```

### Options inherited from parent commands

```sh
  -c, --config stringArray   Override a configuration value in the form of a 'key=value' pair
  -r, --configRoot string    Override configuration root folder (default "/etc/corteca")
  -C, --projectRoot string   Specify project root folder
```

## Configuration

In the `corteca.yaml`, the `build` section -related to the build command- specifies toolchain settings you can alter during the build process.

```yaml
build:
    toolchains:
        image: ghcr.io/nokia/corteca-toolchain:24.2.3
        architectures:
            armv7l:
                platform: "linux/arm/v7"
            aarch64:
                platform: "linux/arm64"
            x86_64:
                platform: "linux/amd64"
    # default toolchain to use when building if none specified
    default: aarch64
    options:
        # rootfs, produces a compressed root filesystem as the build output;
        # oci, the output is an OCI image;
        # docker, for docker image output.
        outputType: rootfs
        # If true, generates a container with additional debugging tools
        # to facilitate troubleshooting during development.
        debug: false
        # Do not inherit variables from the host environment (proxies etc)
        # skipHostEnv: false
        env:
            # Specify custom variables inside build environment
            # <name>: <value>
    crossCompile:
        enabled: true
        image: multiarch/qemu-user-static
        args: [ "--reset", "-p", "yes" ]
```

### Example

This example shows how to use the information of the corteca.yaml file to change the default options during the build:

```bash
corteca build -c 'build.options.debug=true' -c 'build.options.outputType=oci' armv7l
```

Let's try building an application (see [`corteca create hello_world`](./corteca_create.md#example)).
If the command runs successfully the output should be something like this:

![corteca build](../oci-build.PNG?raw=true)
