# PGSD Architecture Diagram (ASCII)

```text
                      +-----------------------------+
                      |        Developer Host       |
                      |   (FreeBSD / PGSD build)    |
                      +-----------------------------+
                                      |
                                      | runs
                                      v
                          +---------------------+
                          |      pgsdbuild      |
                          |   (Go + Lua CLI)    |
                          +---------------------+
                             /             \
                            /               \
                           v                 v
           +---------------------------+   +----------------------------+
           |   Image Build Pipeline    |   |  Bootenv Pipeline          |
           |   (pgsd-desktop, ...)     |   |  (pgsd-bootenv-arcan)      |
           +---------------------------+   +----------------------------+
                |            |                      |            |
                v            v                      v            v
        +-------------+  +--------+         +---------------+  +-----------+
        | ZFS pool &  |  |Overlays|         | Bootenv root  |  |Overlays   |
        | datasets    |  |(common |         | filesystem    |  |(arcan,    |
        | (md-backed) |  | arcan, |         | (for ISO)     |  | bootenv)  |
        +-------------+  | ... )  |         +---------------+  +-----------+
                |            |                      |
                |            +----------------------+   copies
                |                                       images
                v                                          v
        +-----------------------------------+     +------------------------+
        |  ZFS snapshot @install           |     |  Bootenv ISO root      |
        |  zfs send | xz -> root.zfs.xz   |     |  (Arcan + Durden +     |
        |  dd -> efi.img                  |     |   pgsd-inst + images)  |
        |  + manifest.toml                |     +------------------------+
        +-----------------------------------+                 |
                |                                           |
                v                                           v
        +-----------------------------------+       +----------------------+
        |  System Image Artifact Directory |       |   PGSD Bootable ISO   |
        |  artifacts/pgsd-desktop          |       +----------------------+
        |    root.zfs.xz                   |
        |    efi.img                       |
        |    manifest.toml                 |
        +-----------------------------------+
```

This is a condensed diagram; see your main Architecture Diagram document for the
detailed version you maintain interactively.
