# Install

## Install manually from source

You can run the below command to build the application and install the binary to
the `/usr/bin` folder and the default configuration files to the `/etc/corteca`
folder:

```shell
sudo make install
```

To remove a previous manual installation

```shell
sudo make uninstall
```

You can customize the destination folder by overriding the `$DESTDIR`
environment variable:

```shell
# no sudo required; will be installed to ~/.local/share/usr/bin
DESTDIR=~/.local/share make install
```

## Install with package manager

If you are using debian/ubuntu or redhat-based distributions, you can create a relevant package and let your package manager handle installation. E.g. for ubuntu:

```shell
make deb
make rpm
```
