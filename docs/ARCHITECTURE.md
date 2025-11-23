# PGSD Build and Install Architecture

This document gives a high‑level view of the PGSD system:

- `pgsdbuild` builds **system images** (e.g., `pgsd-desktop`) as ZFS send streams.
- `pgsdbuild` also builds a **boot environment ISO** (`pgsd-bootenv-arcan`) that runs
  Arcan + Durden and embeds the installer.
- The installer (`pgsd-inst`) consumes image artifacts (`root.zfs.xz`, `efi.img`,
  `manifest.toml`) and installs them atomically onto a target disk using ZFS.

See DESIGN.md and BUILD_PIPELINE.md for more detail.
