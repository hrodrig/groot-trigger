# FreeBSD port for groot-trigger

Port files for **groot-trigger** (HTTP binary; no rc.d in this skeleton — operators
run under their preferred supervisor, or add a local rc.d later).

## Install from port

When the port is in the official tree:

```bash
cd /usr/ports/sysutils/groot-trigger
make install
```

Local port (copy `Makefile`, `pkg-plist`, `pkg-descr` from this directory):

```bash
cd ~/ports/sysutils/groot-trigger
make install
```

## Test with a local distfile

1. From the **groot-trigger** repo root:

   ```bash
   make port-freebsd-sync
   make man-sync
   make dist-freebsd
   ```

   Output: `dist/groot-trigger_v<version>_freebsd_<arch>.tar.gz` (default `FREEBSD_ARCH=amd64`).

2. Copy into **DISTDIR** or use **`MASTER_SITES=file:///.../`**.

Tarball layout: `groot-trigger`, `share/man/man1/groot-trigger.1`, `share/doc/groot-trigger/{LICENSE,README.md}`.
