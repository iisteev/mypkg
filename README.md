# Basic config

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
