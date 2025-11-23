# PGSD Installer (pgsd-inst)

This document describes the PGSD system installer, a terminal-based installation tool for deploying PGSD images to disk.

## Overview

`pgsd-inst` is a TUI (Text User Interface) installer built with Bubble Tea that performs automated ZFS-based system installations. It provides an interactive interface for selecting system images and target disks, then handles the complete installation pipeline.

## Features

- Interactive TUI with keyboard navigation
- Automatic disk discovery using FreeBSD geom/sysctl
- ZFS-based installation with compression and boot environments
- EFI bootloader installation
- Real-time installation progress logging
- Comprehensive error handling and validation

## Installation Workflow

The installer follows this workflow:

1. **Welcome Screen** - Introduction and confirmation to proceed
2. **Image Selection** - Choose from available system images
3. **Disk Selection** - Choose target disk with size/model information
4. **Confirmation** - Review selections and confirm destructive operation
5. **Installation** - Automated installation with progress logging
6. **Complete** - Success message and reboot instructions

## Technical Pipeline

The installation pipeline performs these steps automatically:

### 1. Disk Partitioning

Creates GPT partition table with two partitions:

```
Partition 1: EFI System Partition (200 MB, FAT32)
Partition 2: ZFS Root Partition (remaining space)
```

Commands:
```sh
gpart destroy -F <disk>
gpart create -s gpt <disk>
gpart add -t efi -s 200M -l efiboot0 <disk>
gpart add -t freebsd-zfs -l zfsroot0 <disk>
```

### 2. EFI Filesystem Creation

Formats the EFI partition as FAT32:

```sh
newfs_msdos -F 32 -c 1 <disk>p1
```

### 3. ZFS Pool Creation

Creates ZFS pool with optimized settings:

```sh
zpool create -f \
    -o altroot=/mnt \
    -O compression=lz4 \
    -O atime=off \
    pgsd <disk>p2
```

Options:
- `altroot=/mnt`: Temporary mount point during installation
- `compression=lz4`: Fast compression (saves ~20-30% space)
- `atime=off`: Performance optimization (no access time updates)

### 4. Root Filesystem Extraction

Extracts compressed ZFS stream to the pool:

```sh
xzcat root.zfs.xz | zfs receive -F pgsd/ROOT/default
```

This creates the complete system filesystem from the pre-built image.

### 5. EFI Partition Installation

Copies EFI boot files to the EFI partition:

```sh
dd if=efi.img of=<disk>p1 bs=1M
```

### 6. Bootloader Installation

Installs FreeBSD EFI bootloader:

```sh
gpart bootcode -p /boot/boot1.efifat -i 1 <disk>
```

### 7. Finalization

Sets boot properties and exports the pool:

```sh
zpool set bootfs=pgsd/ROOT/default pgsd
zpool export pgsd
```

## Disk Discovery

The installer uses multiple methods to discover available disks:

### Primary: geom disk list

```sh
geom disk list
```

Parses output to extract:
- Device name (e.g., `ada0`, `da0`, `nvd0`)
- Mediasize (converted to human-readable format)
- Description (disk model)

Filters out:
- CD-ROM devices (`cd*`)
- Memory disks (`md*`)
- Pass-through devices (`pass*`)

### Fallback: sysctl kern.disks

```sh
sysctl -n kern.disks
```

Returns space-separated list of disk names. Uses `diskinfo` to get details:

```sh
diskinfo <device>
```

### Development: Dummy Disks

If neither method works (development/testing environment), provides dummy disks:
- `ada0` (500GB Virtual Disk)
- `ada1` (1TB Virtual Disk)

## Image Discovery

Searches for system images in these locations:

1. `/usr/local/share/pgsd/images` (production)
2. `artifacts` (development/testing)

Validates each image directory for required files:
- `root.zfs.xz` - Compressed ZFS stream
- `efi.img` - EFI partition image
- `manifest.toml` - Image metadata

## User Interface

### Keyboard Controls

**Navigation:**
- `↑`/`k` - Move cursor up
- `↓`/`j` - Move cursor down
- `Enter` - Select/Confirm
- `q` - Quit (disabled during installation)

**Confirmation Screen:**
- `y`/`Y` - Confirm installation
- `n`/`N` - Go back to disk selection

### Screen Layouts

All screens use a consistent header design:

```
╔════════════════════════════════════════╗
║   PGSD System Installer (Prototype)    ║
╚════════════════════════════════════════╝
```

**Image Selection:**
```
 > pgsd-desktop
   pgsd-server
   pgsd-minimal

↑/↓ or k/j: Navigate
Enter: Select
q: Quit
```

**Disk Selection:**
```
WARNING: All data on the selected disk
will be DESTROYED!

 > ada0 (500GB) - Virtual Disk
   ada1 (1TB) - Virtual Disk

↑/↓ or k/j: Navigate
Enter: Select
q: Quit
```

**Installation Progress:**
```
Starting installation...
Image: pgsd-desktop
Target disk: ada0
Partitioning disk...
Creating EFI system partition...
Creating ZFS pool...
Extracting root filesystem...
Installing EFI partition...
Installing bootloader...
Finalizing installation...
Installation complete!
```

## Building

### Build Installer Only

```sh
make build-installer
```

Output: `bin/pgsd-inst` (approximately 4-5 MB)

### Build All Binaries

```sh
make build
```

Builds both `pgsdbuild` and `pgsd-inst`.

### Install System-Wide

```sh
sudo make install
```

Installs to `/usr/local/bin/`.

## Usage

### From Boot Environment

The installer is automatically available in PGSD boot environments:

```sh
# Auto-launches in minimal variant, or start manually:
sudo pgsd-inst
```

### From Installed System

```sh
# Install from image directory
sudo pgsd-inst

# Will find images in /usr/local/share/pgsd/images
```

### Development/Testing

```sh
# From source directory
sudo bin/pgsd-inst

# Will find images in artifacts/ directory
```

## Requirements

### System Requirements

- FreeBSD 14.0 or later
- Root privileges (sudo)
- 1 GB RAM minimum (2 GB recommended)
- Target disk with 8 GB minimum (20 GB recommended)

### Required Commands

The installer checks for these FreeBSD utilities:

- `gpart` - Partition management
- `newfs_msdos` - FAT32 filesystem creation
- `zpool` - ZFS pool management
- `zfs` - ZFS dataset management
- `xzcat` - XZ decompression
- `dd` - Disk writing

### Boot Environment Requirements

Boot environments must provide:
- EFI boot support
- ZFS kernel module (`kldload zfs`)
- Network access (for downloading additional packages)

## Error Handling

### Validation Errors

**Missing Image Directory:**
```
Error: image directory not found: /path/to/image
```

**Missing Required Files:**
```
Error: required file missing: root.zfs.xz
The image directory must contain: [root.zfs.xz, efi.img, manifest.toml]
```

**Invalid ZPool Name:**
```
Error: zpool name too long (max 63 characters): very_long_name...
```

### Installation Errors

**Insufficient Privileges:**
```
Error: installation must be run as root
Try: sudo pgsd-inst
```

**Disk Partitioning Failed:**
```
Error: disk partitioning failed: ...
Hint: Ensure the disk is not in use and you have root privileges
```

**ZFS Module Not Loaded:**
```
Error: ZFS pool creation failed: ...
Hint: Ensure ZFS kernel module is loaded (kldload zfs)
```

**Missing Bootloader:**
```
Error: bootloader file not found: /boot/boot1.efifat
The system may be missing EFI boot files
```

### Recovery

If installation fails:

1. **Check Logs** - Review error messages in installation log
2. **Verify Disk** - Ensure disk is not in use: `gpart show <disk>`
3. **Check ZFS** - Verify ZFS module loaded: `kldstat | grep zfs`
4. **Clean Up** - Destroy failed pool: `zpool destroy -f pgsd`
5. **Retry** - Run installer again after fixing issues

## Advanced Configuration

### Custom Image Location

Modify `loadImages()` in `main.go`:

```go
imagesDir := "/custom/path/to/images"
```

### Custom ZPool Name

Modify pool name in `performInstallation()`:

```go
cfg := install.Config{
    ZpoolName: "mypool",
    ...
}
```

### Custom Partition Sizes

Modify `partitionDisk()` in `install.go`:

```go
// Increase EFI partition to 512 MB
{"gpart", "add", "-t", "efi", "-s", "512M", "-l", "efiboot0", disk},
```

### Disable Compression

Modify `createZFSPool()` in `install.go`:

```go
cmd := exec.Command("zpool", "create", "-f",
    "-o", "altroot=/mnt",
    // Remove compression line
    "-O", "atime=off",
    poolName, zfsPart)
```

## Development

### Code Structure

```
installer/
├── pgsd-inst/
│   └── main.go              # TUI implementation (Bubble Tea)
└── internal/
    └── install/
        └── install.go       # Installation pipeline
```

### Key Components

**TUI Model:**
- States: Welcome, ImageSelect, DiskSelect, Confirm, Installing, Complete, Error
- Message types: `installLogMsg`, `installCompleteMsg`, `installErrorMsg`
- Commands: Async installation using Bubble Tea command pattern

**Installation Pipeline:**
- Config validation
- Requirement checking
- Disk partitioning
- Filesystem creation
- Data extraction
- Bootloader installation
- Finalization

### Testing

**Manual Testing:**

```sh
# Build installer
make build-installer

# Create test image (if needed)
make build-image IMAGE=pgsd-desktop

# Run installer (requires root and real disk)
sudo bin/pgsd-inst
```

**Unit Tests:**

```sh
# Run all tests
make test

# Test specific package
go test -v ./installer/internal/install
```

**Integration Testing:**

Test in VM with minimal setup:

```sh
# Launch QEMU with virtual disk
qemu-system-x86_64 \
    -cdrom iso/pgsd-bootenv-minimal.iso \
    -boot d \
    -m 2G \
    -enable-kvm \
    -drive file=test-disk.img,format=raw,if=virtio
```

## Troubleshooting

### No Disks Found

**Problem:** Installer shows "no disks found"

**Solutions:**
- Check if disks are detected: `sysctl kern.disks`
- Verify geom works: `geom disk list`
- Check permissions: Run as root

### Installation Hangs at "Extracting root filesystem"

**Problem:** Installation appears frozen during ZFS extraction

**Cause:** Large ZFS stream (500MB-2GB) takes time to decompress and write

**Solution:** Wait patiently (2-10 minutes depending on disk speed)

### EFI Partition Verification Failed

**Problem:** "EFI partition verification failed"

**Cause:** Partition not created or not accessible

**Solutions:**
- Check partitions: `gpart show <disk>`
- Verify device exists: `ls -l /dev/<disk>p1`
- Check dmesg for errors: `dmesg | tail -20`

### Bootloader Installation Failed

**Problem:** "gpart bootcode failed"

**Cause:** Missing /boot/boot1.efifat file

**Solution:**
- Ensure boot environment has EFI boot files
- Check file exists: `ls -l /boot/boot1.efifat`
- Use full variant ISO if missing

### System Won't Boot After Installation

**Problem:** Installed system doesn't boot

**Checklist:**
1. Verify bootfs property: `zpool get bootfs pgsd`
2. Check EFI partition: `mount -t msdosfs /dev/<disk>p1 /mnt && ls /mnt/EFI`
3. Verify boot order in BIOS/UEFI settings
4. Try booting from USB to repair bootloader

## Future Enhancements

Planned improvements:

1. **Real-time Progress Updates** - Stream log messages during installation
2. **Post-Install Configuration** - Set hostname, root password, create users
3. **Network Configuration** - Configure network settings during install
4. **Custom Partitioning** - Allow manual partition layout
5. **Multi-Disk Support** - RAID-Z, mirror configurations
6. **Encryption Support** - GELI/ZFS encryption options
7. **Locale Selection** - Choose language/timezone during install
8. **Package Selection** - Customize installed package sets
9. **Rollback Support** - Undo failed installations automatically

## See Also

- [Build Pipeline Documentation](BUILD_PIPELINE.md)
- [Image Recipes](IMAGE_RECIPES.md)
- [Minimal Variant](MINIMAL_VARIANT.md)
- [Variants Documentation](VARIANTS.md)
- [FreeBSD ZFS Documentation](https://docs.freebsd.org/en/books/handbook/zfs/)
- [Bubble Tea Framework](https://github.com/charmbracelet/bubbletea)
