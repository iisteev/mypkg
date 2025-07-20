# mypkg

mypkg is a simple tool that helps managing application build, installation and removal lifecycle.

mypkg is one of my first GoLang project that I didn't imagine I would need to use today!
Started it as a discovery project to learn GoLang several years ago when I needed to compile applications
in a restricted environment while wanting to keep track every file installed by the simple `make install`
of tarballs.

## Build/Install

```bash
git clone https://github.com/iisteev/mypkg.git
cd mypkg && make build
```

Compiled binary would be in `dist` directory

Or you can directly install it with:
```
# Assuming that you're at the root of this project
make install
```

# Config

Bellow is a simple configuration file to be configured in `~/.mypkg.yaml` file

## Gnu/Linux

> myuser should be replaced by your username

```yaml
prefix: /home/myuser/.opt/usr
installDir: /home/myuser/mypkg/install
buildDir: /home/myuser/mypkg/build
installDBDir: /home/myuser/mypkg/packages
environment:
   - export PATH=${PATH}
   - export CONF_OPTS="--prefix=${PREFIX} --infodir=${PREFIX}/share/info/${PKG_NAME}"
   - export NBJOBS=$(getconf _NPROCESSORS_ONLN)
   - export CFLAGS="-I${PREFIX}/include"
   - export CXXFLAGS="-I${PREFIX}/include"
   - export LDFLAGS="-L${PREFIX}/lib"
   - export PKG_CONFIG_PATH=${PREFIX}/lib/pkg-config
   - export LIBTOOL=llibtool
   - export LIBTOOLIZE=llibtoolize
   - export DYLD_LIBRARY_PATH=${PREFIX}/lib
   - export PKG_CONFIG=${PREFIX}/bin/pkg-config
   - export INSTALL_BIN_DIR=${INSTALL_DIR}/${PREFIX}/bin
dbDir: /home/myuser/.opt/mypkg/packages
```

## MacOS

> myuser should be replaced by your username

```yaml
prefix: /home/myuser/.opt/usr
installDir: /home/myuser/mypkg/install
buildDir: /home/myuser/mypkg/build
installDBDir: /home/myuser/mypkg/packages
environment:
   - export PATH=${PATH}
   - export CONF_OPTS="--prefix=${PREFIX} --infodir=${PREFIX}/share/info/${PKG_NAME}"
   - export NBJOBS=$(getconf _NPROCESSORS_ONLN)
   - export CFLAGS="-I${PREFIX}/include"
   - export CXXFLAGS="-I${PREFIX}/include"
   - export LDFLAGS="-framework CoreFoundation -framework Carbon -L${PREFIX}/lib"
   - export PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
   - export LIBTOOL=llibtool
   - export LIBTOOLIZE=llibtoolize
   - export DYLD_LIBRARY_PATH=${PREFIX}/lib
   - export PKG_CONFIG=${PREFIX}/bin/pkg-config
dbDir: /home/myuser/.opt/mypkg/packages
```

# How it works?

mypkg uses a package/tarball/whatever you would call it, as a reference to see how to handle an app.

Basically it tells from when to download the source, how to build it, how to install it and then create
an archived file that contains built app with few metadata useful to install it later.

The package file format was inspired from [solus/ypkg](https://github.com/getsolus/ypkg)

Here is an example:

```yaml
# Name of the app
name     : htop
# Version of the app
version  : 3.0.5
# Release or the build number
release  : 1
# Home page of the app
homePage : https://htop.dev/
# License of the app
licence  : GPL-2.0-or-later
# Source section contains info from where to download it
source   :
  uri: https://github.com/htop-dev/htop/archive/3.0.5.tar.gz
  sha256: 4c2629bd50895bd24082ba2f81f8c972348aa2298cc6edc6a21a7fa18b73990c
# Setup section is run first to setup things before building it
setup    :
  # Note this is for macos, might not work for gnu/linux
  - sed -i .bak.sh s/glibtoolize/llibtoolize/g ./autogen.sh
  - ./autogen.sh
  - $configure
# build section is run after the setup
build    :
  - $make
# Install section handles installing the app in a temporary installation structure in order to prepare it for archive
install  :
  - $make_install
```

`setup`, `build` and `install` sections accepts any shell command available in your environment.
But also, they provide `macros` (e.g `$configure`, `$make` and `$make_install`) which are predefined
commands based on the environment.

macros are available in `pkg/mpkg/utils.go` file

Another example:
```yaml
---
# Name of the app
name     : helm
# Version Of the app
version  : 3.17.4
# Release or the build number
release  : 1
# Source section contains info from where to download it
source   :
  uri: https://get.helm.sh/helm-v3.17.4-linux-amd64.tar.gz
  sha256: c91e3d7293849eff3b4dc4ea7994c338bcc92f914864d38b5789bab18a1d775d
# Install section on how to install it
install  :
  - mkdir -p $INSTALL_BIN_DIR
  - install $PKG_BUILD_DIR/helm $INSTALL_BIN_DIR/helm
```
