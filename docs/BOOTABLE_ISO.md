# Bootable ISO Documentation

This document explains how the PGSD bootable ISOs work and how to use them properly.

## Boot Modes Supported

PGSD ISOs support multiple boot modes:

1. **BIOS Boot (Legacy)**: Using El Torito boot standard with `boot/cdboot`
2. **USB Boot (BIOS)**: Hybrid MBR structure allowing USB boot in BIOS mode
3. **UEFI Boot**: Using EFI System Partition with `EFI/BOOT/BOOTX64.EFI`

## Boot Configuration

### Root Filesystem

The ISO uses a **CD9660** (ISO 9660) filesystem as the root filesystem. The boot loader is configured during ISO build to explicitly mount using the ISO9660 volume label:

- `overlays/bootenv/boot/loader.conf` - Boot environment configuration

During the ISO build process, the `vfs.root.mountfrom` setting is automatically configured with the ISO volume label (e.g., `vfs.root.mountfrom="cd9660:/dev/iso9660/PGSDBOOTENVARCAN"`). This explicit configuration ensures reliable BIOS boot on bare metal, where auto-detection can fail.

### Writable Filesystem Overlays

Since the ISO is read-only, the boot environment sets up tmpfs overlays for writable areas:

- `/tmp` - 512MB tmpfs
- `/var` - 256MB tmpfs
- User home directories

This is configured in `overlays/bootenv/etc/rc.conf.d/bootenv`.

## Creating Bootable USB Drives

### Using `dd` (All Platforms)

```bash
# FreeBSD
dd if=pgsd-bootenv-arcan.iso of=/dev/da0 bs=1M status=progress

# Linux
dd if=pgsd-bootenv-arcan.iso of=/dev/sdb bs=1M status=progress

# macOS
dd if=pgsd-bootenv-arcan.iso of=/dev/rdisk2 bs=1m
```

**⚠️ WARNING**: Replace `/dev/da0`, `/dev/sdb`, or `/dev/rdisk2` with your actual USB drive device. Using the wrong device will destroy data!

### GPT Table Structure

The ISO images include a complete hybrid GPT/MBR boot structure with both primary and secondary GPT tables. This ensures compatibility with both BIOS and UEFI boot modes.

**Note:** Older versions of the build system only wrote the primary GPT table, which caused GEOM warnings about corrupt or invalid secondary GPT tables. This has been fixed - both primary and secondary GPT tables are now written correctly.

### Alternative: Using Etcher or Rufus

For a GUI experience, you can use:

- **Etcher** (Linux/macOS/Windows): https://www.balena.io/etcher/
- **Rufus** (Windows): https://rufus.ie/
  - Select "DD Image" mode when prompted
  - Do NOT use "ISO Image" mode

## Booting the ISO

### BIOS Boot

1. Insert USB drive or load ISO in VM
2. Enter BIOS/Boot menu (usually F12, F2, or Del during startup)
3. Select "USB Drive" or "CD/DVD" (should show as "Legacy" or "BIOS" mode)
4. Boot should proceed to FreeBSD boot loader
5. System will auto-mount the CD9660 filesystem

### UEFI Boot

1. Insert USB drive or load ISO in VM
2. Enter UEFI/Boot menu (usually F12, F2, or Del during startup)
3. Select "UEFI: USB Drive" or "UEFI: CD/DVD"
4. Boot should proceed to FreeBSD boot loader
5. System will auto-mount the CD9660 filesystem

### Troubleshooting

#### Stuck at `mountroot>` prompt

This indicates the boot loader couldn't find the root filesystem. **This issue has been fixed** in recent builds by explicitly configuring the ISO9660 volume label in loader.conf during build.

If you're using an older ISO and encounter this:

1. At the `mountroot>` prompt, type: `?`
2. This will list available devices (look for `/dev/iso9660/<LABEL>`)
3. Try manually mounting: `cd9660:/dev/iso9660/<VOLUME_LABEL>`

**Volume label:**
- `PGSD` for all PGSD boot ISOs

**Example:**
```
mountroot> cd9660:/dev/iso9660/PGSD
```

If this happens with a recently built ISO, please report it as a bug.

**Fix for older ISOs:** Rebuild with the latest version which includes the configureBootLoader step that injects the correct volume label.

#### UEFI firmware doesn't detect the USB drive

Some UEFI implementations are very strict about EFI boot structures. If your system doesn't recognize the USB drive in UEFI mode:

1. Try updating your system firmware
2. Use BIOS/Legacy mode instead
3. Try a different USB port (some systems only boot from specific ports)
4. Ensure Secure Boot is disabled in UEFI settings
5. Check if the ISO tool used during build supports UEFI (xorriso is recommended)

#### Boot hangs or kernel panic

1. Try adding `boot_verbose="YES"` to `/boot/loader.conf` before building the ISO
2. Check system RAM (minimum 2GB recommended, 4GB for full environment)

## Building ISOs

ISOs are built using the `pgsdbuild` tool. The build system will **automatically download** FreeBSD base system archives if they're not already cached:

```bash
# Build Arcan boot environment ISO
# Archives will be auto-downloaded if not present
pgsdbuild iso pgsd-bootenv-arcan
```

**Auto-fetch Configuration:**

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

### Required Tools

For best results, build on FreeBSD with these tools:

- `makefs` - FreeBSD's ISO creation tool (preferred)
- `mkimg` - For hybrid GPT/MBR boot structures
- `mtools` - For creating EFI boot images (optional but recommended)

On Linux, these tools work:

- `xorriso` - Best Linux ISO tool with UEFI support
- `genisoimage` or `mkisofs` - Legacy tools (UEFI support added)
- `mtools` - For creating EFI boot images

### Cross-Building

To build ISOs on Linux using FreeBSD boot files:

```bash
export FREEBSD_ROOT=/path/to/freebsd/extracted
pgsdbuild iso pgsd-bootenv-arcan
```

The `FREEBSD_ROOT` should point to an extracted FreeBSD base system with boot files.

## Technical Details

### Hybrid Boot Structure

The ISOs use a hybrid boot structure that combines:

1. **ISO 9660 filesystem** with Rock Ridge extensions
2. **El Torito boot catalog** for CD/DVD boot
3. **MBR partition table** for USB boot detection
4. **GPT partition table** for UEFI boot (via mkimg or xorriso)

This allows a single ISO to boot on:
- CD/DVD drives (BIOS)
- USB drives (BIOS and UEFI)
- Virtual machines (BIOS and UEFI)

### Boot Sequence

#### BIOS Mode:
1. BIOS loads `boot/cdboot` (CD) or `boot/isoboot` (USB) from MBR
2. Boot loader loads `boot/loader`
3. `boot/loader` reads `boot/loader.conf`
4. Kernel is loaded from `boot/kernel/kernel`
5. Kernel mounts CD9660 root filesystem (auto-detected)
6. Init process runs, starting `rc.d` scripts
7. Boot environment setup creates tmpfs overlays
8. Arcan/Durden starts with installer registered

#### UEFI Mode:
1. UEFI firmware loads `EFI/BOOT/BOOTX64.EFI`
2. EFI boot loader loads FreeBSD `boot/loader.efi`
3. Rest is same as BIOS mode

### File Locations

- **Boot files**: `boot/` directory (cdboot, isoboot, loader, kernel)
- **EFI boot**: `EFI/BOOT/BOOTX64.EFI`
- **Embedded images**: `/usr/local/share/pgsd/images/`
- **Installer binary**: `/usr/local/bin/pgsd-inst`

## See Also

- [FreeBSD Handbook: Creating Bootable Media](https://docs.freebsd.org/en/books/handbook/bsdinstall/)
- [ISO 9660 Standard](https://en.wikipedia.org/wiki/ISO_9660)
- [El Torito Specification](https://en.wikipedia.org/wiki/El_Torito_(CD-ROM_standard))
- [UEFI Specification](https://uefi.org/specifications)
