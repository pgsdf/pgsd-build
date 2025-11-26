# PGSD Build System

**Pacific Grove Software Distribution (PGSD)** is a FreeBSD-based distribution build and installation system featuring ZFS-backed system images, declarative Lua configuration, and a user-friendly TUI installer.

## Overview

This system provides a complete solution for building, distributing, and installing FreeBSD-based systems:

- **`pgsdbuild`** - Build tool for creating ZFS system images and bootable ISOs
- **`pgsd-inst`** - Interactive TUI installer with image selection and ZFS installation
- **Lua Configuration** - Declarative image and variant recipes
- **ZFS-Native** - Atomic installations with snapshot/rollback capabilities
- **Comprehensive Error Handling** - User-friendly validation and helpful error messages

## Features

### Build System
- **Lua-based configuration** - Declarative recipes for images and boot environments
- **ZFS image creation** - Atomic, snapshot-based system images
- **Bootable ISO generation** - Live boot environments with embedded installer
- **Overlay system** - Filesystem overlays for customization
- **Comprehensive validation** - Pre-flight checks with helpful error messages

### Installer
- **Interactive TUI** - Built with Bubble Tea for a smooth user experience
- **Image selection** - Choose from multiple system configurations
- **Disk detection** - Automatic disk discovery with `diskinfo` integration
- **Progress tracking** - Real-time installation progress with detailed logging
- **Error recovery** - Clear error messages with actionable hints

### Error Handling
- **Pre-flight validation** - Catches configuration errors before destructive operations
- **System requirements checking** - Verifies all required tools are available
- **File existence validation** - Ensures all required files are present
- **Helpful hints** - Context-aware suggestions for common issues
- **Root privilege checking** - Confirms proper permissions before installation

## Quick Start

### Prerequisites

- FreeBSD system (or compatible OS)
- Go 1.21 or later
- ZFS support
- Standard FreeBSD tools (`gpart`, `newfs_msdos`, `zpool`, `zfs`, `xzcat`, `dd`)

### Building

```bash
# Clone the repository
git clone https://github.com/pgsdf/pgsd-build.git
cd pgsd-build

# Build both tools
make

# Or build individually
make build-pgsdbuild
make build-installer
```

Binaries will be created in the `bin/` directory.

### Installation

```bash
# Install system-wide (requires root)
sudo make install

# Uninstall
sudo make uninstall
```

## Usage

### Building Images

```bash
# List available images
./bin/pgsdbuild list-images

# Build a specific image
./bin/pgsdbuild image pgsd-desktop

# Build all images
make build-all-images
```

This creates artifacts in `artifacts/<image-id>/`:
- `root.zfs.xz` - Compressed ZFS snapshot
- `efi.img` - EFI system partition
- `manifest.toml` - Build metadata

### Building Boot ISOs

```bash
# List available variants
./bin/pgsdbuild list-variants

# Build a bootable ISO (recommended: use FreeBSD distribution archives)
# Place base.txz and kernel.txz in freebsd-dist/ directory
./bin/pgsdbuild iso pgsd-bootenv-arcan

# Build with make
make build-iso VARIANT=pgsd-bootenv-arcan

# Build all ISOs (includes all images)
make build-all-isos
```

#### FreeBSD Base System Requirements

**Automatic Method (Recommended):** Let the build system fetch archives automatically

The build system can **automatically download** FreeBSD's official **base.txz** and **kernel.txz** archives from FreeBSD mirrors. This is the **easiest and recommended approach**:

```bash
# Auto-fetch is enabled by default - just build!
./bin/pgsdbuild iso pgsd-bootenv-arcan

# The build system will:
# 1. Check if archives exist in freebsd-dist/
# 2. If not, download them from FreeBSD mirrors
# 3. Verify checksums for integrity
# 4. Cache them for future builds
```

**Configuration:**

The auto-fetch feature is controlled by environment variables:

```bash
# Specify FreeBSD version (default: 15.0-RC3)
export FREEBSD_VERSION=15.0-RC3

# Specify architecture (default: amd64)
export FREEBSD_ARCH=amd64

# Use a custom mirror (optional)
export FREEBSD_MIRROR=https://mirror.example.com

# Disable auto-fetch (if you prefer manual downloads)
export PGSD_AUTO_FETCH=0
```

**Manual Method:** Download archives yourself

If you prefer to download archives manually or auto-fetch is disabled:

```bash
# Download FreeBSD distribution archives (example for FreeBSD 14.2)
cd pgsd-build
mkdir -p freebsd-dist
cd freebsd-dist

# Download from FreeBSD mirrors
fetch https://download.freebsd.org/releases/amd64/14.2-RELEASE/base.txz
fetch https://download.freebsd.org/releases/amd64/14.2-RELEASE/kernel.txz

# Or extract from FreeBSD installation media
# (if you have FreeBSD-14.2-RELEASE-amd64-disc1.iso mounted)
cp /mnt/usr/freebsd-dist/base.txz .
cp /mnt/usr/freebsd-dist/kernel.txz .

# Return to repo root and build
cd ..
./bin/pgsdbuild iso pgsd-bootenv-arcan
```

The build system will automatically find and extract these archives.

**What's in the archives:**
- **base.txz** - Complete FreeBSD base system (/bin, /sbin, /lib, /etc, /usr, etc.)
- **kernel.txz** - FreeBSD kernel and all boot files (/boot directory with loader, Lua scripts, etc.)

**Why this works better:**
- ✅ **Automatic** - No manual downloads needed
- ✅ **Cross-platform** - Works on Linux, macOS, and FreeBSD
- ✅ **Complete and guaranteed to boot** - Official FreeBSD builds
- ✅ **No missing files** - All boot files included
- ✅ **Reproducible** - Same builds across different systems
- ✅ **Verified** - Automatic checksum verification
- ✅ **Cached** - Downloads once, reuses for future builds

**Alternative Method (Fallback):** Copy from FREEBSD_ROOT

If base.txz and kernel.txz are not available, the build system falls back to copying from `FREEBSD_ROOT`:

```bash
# Extract archives to freebsd-dist/root/
mkdir -p freebsd-dist/root
cd freebsd-dist/root
tar -xJf ../base.txz
tar -xJf ../kernel.txz
cd ../..

# Build with auto-detection
./bin/pgsdbuild iso pgsd-bootenv-arcan

# Or explicitly set FREEBSD_ROOT
FREEBSD_ROOT=/path/to/freebsd-dist/root ./bin/pgsdbuild iso pgsd-bootenv-arcan
```

**Note:** The FREEBSD_ROOT fallback method is less reliable and may have missing files. Always prefer using the distribution archives directly.

ISOs are created in `iso/<variant-id>.iso`

### Running the Installer

```bash
# Run the TUI installer
sudo ./bin/pgsd-inst
```

The installer will guide you through:
1. **Image Selection** - Choose your desired system configuration
2. **Disk Selection** - Select target disk (with size/model info)
3. **Confirmation** - Review and confirm installation
4. **Installation** - Automated ZFS-based installation
5. **Completion** - Ready to reboot into new system

## Directory Structure

```
pgsd-build/
├── cmd/
│   └── pgsdbuild/          # Build tool CLI
├── installer/
│   ├── pgsd-inst/          # Installer TUI
│   └── internal/install/   # Installation pipeline
├── internal/
│   ├── config/             # Lua config loader
│   ├── image/              # Image build pipeline
│   └── iso/                # ISO build pipeline
├── images/                 # Image recipes (*.lua)
├── variants/               # Boot variant recipes (*.lua)
├── overlays/               # Filesystem overlays
│   ├── common/
│   ├── arcan/
│   └── bootenv/
├── artifacts/              # Built images (generated)
├── iso/                    # Built ISOs (generated)
└── docs/                   # Architecture documentation
```

## Configuration

### Image Recipe Example (`images/my-image.lua`)

```lua
return {
  id = "my-image",
  version = "1.0.0",
  zpool_name = "mypool",
  root_dataset = "mypool/ROOT/default",
  pkg_lists = {
    "base",
    "desktop/xorg",
  },
  overlays = {
    "common",
  },
}
```

### Required Fields
- `id` - Unique image identifier (max 64 chars)
- `version` - Semantic version string
- `zpool_name` - ZFS pool name
- `root_dataset` - Root dataset path

### Variant Recipe Example (`variants/my-bootenv.lua`)

```lua
return {
  id = "my-bootenv",
  name = "My Boot Environment",
  pkg_lists = {
    "base",
    "installer/pgsd-inst",
  },
  overlays = {
    "common",
    "bootenv",
  },
  images_dir = "/usr/local/share/pgsd/images",
}
```

## Makefile Targets

### Building
- `make` or `make all` - Build both tools
- `make build-pgsdbuild` - Build only pgsdbuild
- `make build-installer` - Build only installer
- `make clean` - Remove all build artifacts

### Images & ISOs
- `make list-images` - List available images
- `make list-variants` - List available variants
- `make build-image IMAGE=<name>` - Build specific image
- `make build-iso VARIANT=<name>` - Build specific ISO
- `make build-all-images` - Build all images
- `make build-all-isos` - Build all images and ISOs

### Development
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Run linter
- `make deps` - Update dependencies

### System Installation
- `make install` - Install to /usr/local/bin
- `make uninstall` - Remove from /usr/local/bin

## Error Messages & Troubleshooting

The system provides helpful error messages with context and hints:

### Configuration Errors
```
image config missing required fields: [version zpool_name]
Example: return { id = "my-image", version = "1.0", zpool_name = "mypool", ... }
```

### System Requirements
```
system requirements not met: required commands not found: [zpool zfs]
Please ensure these tools are installed and in PATH
```

### Installation Errors
```
ZFS pool creation failed: zpool create failed
Hint: Ensure ZFS kernel module is loaded (kldload zfs)
```

### Common Solutions

**Missing ZFS tools:**
```bash
# Load ZFS kernel module
sudo kldload zfs

# Install ZFS (if not present)
sudo pkg install zfs
```

**Permission denied:**
```bash
# Installer requires root
sudo pgsd-inst
```

**Missing image files:**
- Ensure you've run `pgsdbuild image <name>` before building ISOs
- Check `artifacts/<image-id>/` contains `root.zfs.xz`, `efi.img`, and `manifest.toml`

**Boot errors with ISOs:**
```
ERROR: cannot open /boot/lua/loader.lua: no such file or directory
can't access /etc/rc : No such file or directory
login_getclass : unknown class 'daemon'
```
These errors indicate the base system is incomplete. **Solution: Enable auto-fetch (default) or manually download archives:**

```bash
# Method 1: Auto-fetch (recommended - enabled by default)
# Just rebuild the ISO - archives will be downloaded automatically
./bin/pgsdbuild iso pgsd-bootenv-arcan

# Method 2: Manual download (if auto-fetch is disabled)
mkdir -p freebsd-dist
cd freebsd-dist
fetch https://download.freebsd.org/releases/amd64/14.2-RELEASE/base.txz
fetch https://download.freebsd.org/releases/amd64/14.2-RELEASE/kernel.txz
cd ..

# Rebuild the ISO - archives will be auto-detected and extracted
./bin/pgsdbuild iso pgsd-bootenv-arcan
```

The archive-based approach ensures all boot files are present and eliminates these recurring issues.

## Architecture

### Build Pipeline

1. **Image Building** (`pgsdbuild image`)
   - Creates memory-backed disk
   - Partitions with GPT (EFI + ZFS)
   - Creates ZFS pool and datasets
   - Installs packages and overlays
   - Snapshots and exports as compressed stream

2. **ISO Building** (`pgsdbuild iso`)
   - Builds bootenv filesystem
   - Installs Arcan, Durden, and installer
   - Copies system images to ISO
   - Registers Arcan targets
   - Creates bootable ISO

### Installation Pipeline

1. **Validation** - Configuration and system requirements
2. **Partitioning** - GPT with EFI and ZFS partitions
3. **Filesystems** - FAT32 EFI and ZFS pool
4. **Extraction** - Decompress and receive ZFS stream
5. **EFI Setup** - Copy EFI partition and bootloader
6. **Finalization** - Set bootfs and export pool

## Documentation

Detailed documentation available in `docs/`:
- `ARCHITECTURE.md` - System architecture overview
- `DESIGN.md` - Design principles and philosophy
- `BUILD_PIPELINE.md` - Build pipeline details
- `ROADMAP.md` - Development roadmap

## Contributing

This is a prototype demonstrating the PGSD build system architecture. Contributions welcome for:
- Production-ready FreeBSD integration
- Additional package management
- Multiple architecture support
- Secure Boot integration
- Automated testing

## License

See LICENSE file for details.

## Resources

- Architecture Diagram: `docs/ARCHITECTURE_DIAGRAM.md`
- Design Document: `docs/DESIGN.md`
- Build Pipeline: `docs/BUILD_PIPELINE.md`

---

**Note:** This is a prototype implementation. For production use, additional hardening, testing, and FreeBSD-specific integration is required.
