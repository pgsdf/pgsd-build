# PGSD Build Pipeline (Summary)

The PGSD build pipeline produces two main artifact types:

1. **System Images** (e.g., `pgsd-desktop`)
   - Exported as `root.zfs.xz` + `efi.img` + `manifest.toml`
   - Located in `artifacts/<image-id>/`

2. **Boot Environment ISO** (`pgsd-bootenv-arcan`)
   - Root filesystem built from packages and overlays
   - Contains Arcan, Durden, and the installer (`pgsd-inst`)
   - Contains system images under `/usr/local/share/pgsd/images`

High‑level flow:

```text
pgsdbuild image pgsd-desktop
    -> artifacts/pgsd-desktop/{root.zfs.xz, efi.img, manifest.toml}

pgsdbuild iso pgsd-bootenv-arcan
    -> bootenv ISO that includes pgsd-inst and system images
```

For a detailed description, see the main BUILD_PIPELINE.md you maintain in your repo;
this file is a compact version for the prototype.
