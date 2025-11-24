# Bootable ISO Documentation

This document explains how the PGSD bootable ISOs work and how to use them properly.

## Boot Modes Supported

PGSD ISOs support multiple boot modes:

1. **BIOS Boot (Legacy)**: Using El Torito boot standard with `boot/cdboot`
2. **USB Boot (BIOS)**: Hybrid MBR structure allowing USB boot in BIOS mode
3. **UEFI Boot**: Using EFI System Partition with `EFI/BOOT/BOOTX64.EFI`

## Boot Configuration

### Root Filesystem

The ISO uses a **CD9660** (ISO 9660) filesystem as the root filesystem. The boot loader configuration files have been set up to auto-detect the boot device:

- `overlays/bootenv/boot/loader.conf` - For full boot environment
- `overlays/bootenv-minimal/boot/loader.conf` - For minimal installer

These configurations use `vfs.root.mountfrom=""` to let FreeBSD auto-detect the CD9660 filesystem.

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

### Expected Warning: GPT Header

When using `dd` to write the ISO to USB media, you may see this warning:

```
GEOM: da0: the secondary GPT header is not in the last LBA.
```

**This is expected behavior and does not prevent booting!**

#### Why This Happens

The ISO image has a GPT table with a backup header at the end of the ISO file. When you write the ISO to a larger USB drive using `dd`, the backup header is no longer at the end of the physical disk, but rather at the end of the ISO image within the disk.

#### Is This a Problem?

No! The ISO will still boot correctly in both BIOS and UEFI modes. The primary GPT header at the beginning of the disk is sufficient for booting.

#### How to Fix (Optional)

If you want to eliminate the warning, you can recover the GPT table after writing:

**On FreeBSD:**
```bash
gpart recover da0
```

**On Linux:**
```bash
sgdisk -e /dev/sdb
```

This moves the backup GPT header to the actual end of the disk.

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

This indicates the boot loader couldn't auto-detect the root filesystem. This should be fixed in the latest version, but if you encounter it:

1. At the `mountroot>` prompt, type: `?`
2. This will list available devices
3. Try manually mounting: `cd9660:/dev/cd0` or `cd9660:iso9660/<VOLUME_LABEL>`

**Volume labels:**
- `PGSDBOOTENVARCAN` for pgsd-bootenv-arcan
- `PGSDBOOTENVMINIMAL` for pgsd-bootenv-minimal

If this happens, please report it as a bug.

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
3. Try the minimal ISO variant (`pgsd-bootenv-minimal`) which has lower requirements

## Building ISOs

ISOs are built using the `pgsdbuild` tool:

```bash
# Build full Arcan boot environment ISO
pgsdbuild iso pgsd-bootenv-arcan

# Build minimal installer ISO
pgsdbuild iso pgsd-bootenv-minimal
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
