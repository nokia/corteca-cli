# Getting Started

## Install corteca cli

To install or build from source, you can follow the [relevant instructions](../README.md#install) in the main README file. To make sure everything is in place, run the following command in any folder.

```shell
corteca config get
```

You should see your default system-wide configuration.

## Create Application

Scaffolding applications depends on existing templates in your system. By
default, corteca-cli includes templates for C, C++ and Go. For more information
on how to write templates, refer to the [relevant guide](templates.md). We will create a C application in the rest of this example.

Change to the parent folder where you want your application directory to be
created and run the following command:

```shell
corteca create test_app
```

Follow the prompt instructions to populate the application settings. A
folder named after the given parameter (test_app) in the above command will be
created with all the generated files inside. The structure of the folder should
be similar to the below:

```bash
test_app
├── .corteca
│   ├── ADF.template
│   └── Dockerfile.template
├── corteca.yaml # main configuration file
├── .gitignore
├── Makefile
├── src # your application source code will live here
│   ├── Makefile
│   └── test_app.c
└── target # files to be placed directly in your container rootfs
    ├── aarch64
    ├── armv7l
    ├── noarch
    └── x86_64
```

### Build application

To build the application, corteca-cli uses a docker container where all the
needed dependencies and tools are installed. The dockerfile will be auto
generated on build using the `Dockerfile.template` found under `.corteca`
folder.

For this example we will build an OCI image build for armv7l architecture.

Enter the folder created on the previous step.

`cd test_app`
`corteca build -c 'build.options.outputType=oci' armv7l`

When the build process is finished, the generated artifact will be placed in
the `dist/` folder. For this example a `test-1.0-armv7l-oci.tar` will be
generated.

### Publish your application

Corteca cli provides a publish functionality to help users upload their produced artifacts in an OCI registry. There are some registries set up by default in corteca that can be used as a template from the user to set up the target registry. You can check the current configuration for publish using

```shell
corteca config get publish
```

#### Publish in corteca local registry

User can use the already set up local registry.

`corteca publish localRegistry armv7l`

Corteca will set up a local oci registry and will upload the artifact. (check with <https://localhost:8080/v2/_catalog>). If this registry is needed to be visible outside the host device, publicURL needs to be changed to the device IP.

`corteca config set publish.localRegistry.publicURL http://192.168.18.2:8080`

#### Publish in custom registry

User can add his own registry using *corteca config*

`corteca config add publish "{myregistry: { addr: https://my-registry.com, auth: basic, username: user1, password: pass1, method: push}}"`

Notice that a new section was added in corteca.yaml with all the registry info.

After the registry is set up, the application can be uploaded.

`corteca publish myregistry armv7l`

For more info on corteca publish check [corteca publish](reference/corteca_publish.md) guide.

## Configuring application

Application configuration is noted in the corteca.yaml file placed in application's folder when created. The *app* section -related to the create command- specifies app settings you can alter during the build process.
The complete set of settings for Corteca Cli is created by combining this file with /etc/corteca/corteca.yaml.
Corteca cli provides methods to show or edit these settings using the *config* method. You can also check or edit the settings values using the auto complete of Corteca cli by pressing <TAB> key twice like in a shell.

`corteca config get app`

To check a specific value

`corteca config get app.version`

To set a value

`corteca config set app.version 1.0.1`

For more detailed guide on config get/set/add please refer to [corteca config](reference/corteca_config.md)  guide.
