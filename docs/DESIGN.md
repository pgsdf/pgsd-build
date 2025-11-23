# PGSD System Design Document

PGSD is built around three major layers:

1. A **Go + Lua build system** (`pgsdbuild`) that creates:
   - Installable ZFS images (e.g., `pgsd-desktop`)
   - Boot environment ISO root for `pgsd-bootenv-arcan`
2. A **Boot Environment** using Arcan and Durden, which:
   - Runs as a live ISO
   - Registers `pgsd-inst` as an Arcan target (`pgsd-installer`)
3. An **Installer** (`pgsd-inst`) implemented in Go with Bubble Tea, which:
   - Lists available images from `/usr/local/share/pgsd/images`
   - Lists disks using `diskinfo`
   - Drives the ZFS install pipeline

The key design principles are:

- Deterministic outputs
- Clear separation of build / boot / installed system
- Minimal shell usage
- Declarative configuration (Lua) and imperative execution (Go)
