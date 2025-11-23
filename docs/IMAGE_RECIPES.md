# PGSD Image Recipes Guide

This document explains how to create and customize PGSD image recipes.

## Overview

Image recipes are Lua files that define a complete FreeBSD system image with:
- System packages (FreeBSD base, applications, drivers)
- Configuration overlays (system settings, user defaults)
- ZFS layout (datasets, snapshots, boot environments)
- Boot configuration (loader settings, kernel modules)

## Recipe Structure

### Required Fields

Every image recipe must include these fields:

```lua
return {
  id = "pgsd-desktop",              -- Unique identifier
  version = "0.1.0",                 -- Version string
  zpool_name = "pgsd",               -- ZFS pool name
  root_dataset = "pgsd/ROOT/default", -- Root BE dataset
  pkg_lists = { "base", ... },       -- Package lists to install
  overlays = { "common", ... },      -- Overlays to apply
}
```

### Optional Fields

Additional metadata can be included for documentation:

```lua
  -- Boot loader configuration
  boot = {
    loader_conf = {
      ["vfs.root.mountfrom"] = "zfs:pgsd/ROOT/default",
      ["zfs_load"] = "YES",
    },
  },

  -- ZFS dataset layout
  datasets = {
    { name = "pgsd/ROOT/default", mountpoint = "/", canmount = "noauto" },
    { name = "pgsd/home", mountpoint = "/home", canmount = "on" },
  },

  -- System configuration
  system = {
    hostname = "pgsd-desktop",
    timezone = "UTC",
    services = { "sshd", "ntpd" },
    users = {
      { name = "pgsd", groups = { "wheel" } },
    },
  },
}
```

## Package Lists

Package lists are logical groupings of FreeBSD packages. See [PACKAGE_LISTS.md](PACKAGE_LISTS.md) for details.

Common package lists:
- `base` - FreeBSD base system
- `desktop/arcan` - Arcan display server
- `desktop/durden` - Durden window manager
- `desktop/apps` - Desktop applications
- `dev/tools` - Development tools
- `system/network` - Network management
- `system/audio` - Audio subsystem
- `system/graphics` - Graphics drivers

## Overlays

Overlays are directory trees copied into the image filesystem. They contain:
- Configuration files (`/etc`, `/boot`)
- User defaults (`/etc/skel`)
- Scripts and utilities (`/usr/local/bin`)
- Resources (`/usr/local/share`)

Standard overlays:
- `common` - Base configuration for all PGSD systems
- `desktop` - Desktop environment settings
- `arcan` - Arcan/Durden configuration
- `bootenv` - Boot environment (ISO) specific

### Overlay Structure

```
overlays/common/
  boot/
    loader.conf                # Boot loader config
  etc/
    motd                       # Message of the day
    sysctl.conf                # Kernel tuning
    rc.conf.d/                 # Service configs
      zfs
      network
      services
```

## Example: Minimal Server Image

```lua
-- images/pgsd-server.lua
return {
  id = "pgsd-server",
  version = "0.1.0",
  zpool_name = "pgsd",
  root_dataset = "pgsd/ROOT/default",

  pkg_lists = {
    "base",
    "system/network",
  },

  overlays = {
    "common",
    "server",  -- Server-specific overlay
  },

  system = {
    hostname = "pgsd-server",
    services = {
      "sshd",
      "ntpd",
    },
  },
}
```

## Example: Custom Desktop Image

```lua
-- images/pgsd-workstation.lua
return {
  id = "pgsd-workstation",
  version = "0.1.0",
  zpool_name = "pgsd",
  root_dataset = "pgsd/ROOT/default",

  pkg_lists = {
    "base",
    "desktop/arcan",
    "desktop/durden",
    "desktop/apps",
    "dev/tools",
    "system/network",
    "system/audio",
    "system/graphics",
    "multimedia/production",  -- Custom package list
  },

  overlays = {
    "common",
    "desktop",
    "arcan",
    "workstation",  -- Custom overlay
  },

  datasets = {
    { name = "pgsd/ROOT/default", mountpoint = "/", canmount = "noauto" },
    { name = "pgsd/home", mountpoint = "/home", canmount = "on" },
    { name = "pgsd/projects", mountpoint = "/projects", canmount = "on" },
  },
}
```

## Build Process

To build an image:

```bash
# List available images
pgsdbuild list-images

# Build specific image
pgsdbuild image pgsd-desktop

# Build with verbose output
pgsdbuild -v image pgsd-desktop

# Keep work directory for debugging
pgsdbuild --keep-work image pgsd-desktop
```

## Output Artifacts

After building, artifacts are in `artifacts/<image-id>/`:

```
artifacts/pgsd-desktop/
  root.zfs.xz         # Compressed ZFS stream
  efi.img             # EFI boot partition
  manifest.toml       # Build manifest
```

## ZFS Dataset Layout

PGSD uses a standard ZFS dataset layout:

```
pgsd                              # Pool root
├── ROOT                          # Boot environments container
│   └── default                   # Default boot environment
├── home                          # User home directories (shared)
├── var                           # Variable data
│   ├── log                       # Logs
│   └── tmp                       # Temporary files
└── usr
    └── local                     # Local software (shared)
```

Boot environments (BEs) are isolated in `ROOT/*`. User data and local software are shared across BEs.

## Best Practices

1. **Version Everything**: Use semantic versioning (0.1.0, 1.0.0, etc.)
2. **Minimal Base**: Start with minimal package lists, add as needed
3. **Separate Concerns**: Use overlays for different configuration domains
4. **Document Choices**: Add comments explaining non-obvious settings
5. **Test Thoroughly**: Build and test images before deploying
6. **Use Boot Environments**: Always test updates in new BEs

## Advanced Topics

### Custom Package Lists

Create `pkg-lists/<name>.txt`:

```
# pkg-lists/custom/multimedia.txt
mpv
ffmpeg
obs-studio
kdenlive
```

Reference in recipe:

```lua
pkg_lists = { "base", "custom/multimedia" }
```

### Dynamic Configuration

Use Lua code for conditional configuration:

```lua
local hostname = os.getenv("PGSD_HOSTNAME") or "pgsd-desktop"

return {
  id = "pgsd-custom",
  -- ... other fields ...
  system = {
    hostname = hostname,
  },
}
```

### Multiple Variants

Create variant-specific recipes:

- `pgsd-desktop-minimal.lua` - Lightweight desktop
- `pgsd-desktop-full.lua` - Full-featured desktop
- `pgsd-desktop-dev.lua` - Development environment

## Troubleshooting

### Build Failures

Check logs with verbose output:
```bash
pgsdbuild -v image <name>
```

### Missing Overlays

Ensure overlay directories exist:
```bash
ls overlays/common overlays/desktop overlays/arcan
```

### Package Issues

In production, package installation failures would be logged. Check package availability:
```bash
pkg search <package-name>
```

## See Also

- [Package Lists Reference](PACKAGE_LISTS.md)
- [Build Pipeline](BUILD_PIPELINE.md)
- [Architecture](ARCHITECTURE.md)
