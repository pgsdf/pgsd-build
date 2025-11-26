# PGSD Variant Reference

This document explains PGSD variants and how to create bootable ISO images.

## Overview

Variants are configurations for bootable installation media (ISO images). Unlike image recipes which define installed systems, variants define the live boot environment used for installation.

A variant contains:
- Live boot environment packages
- Installer application (pgsd-inst)
- Embedded system images for installation
- Boot environment configuration
- Auto-login and live user setup

## Variant vs Image

**Image Recipe** (`images/*.lua`):
- Defines the **installed system**
- Output: ZFS stream + EFI partition + manifest
- Example: `pgsd-desktop` (the system you install)

**Variant** (`variants/*.lua`):
- Defines the **boot environment** (live ISO)
- Output: Bootable ISO file
- Example: `pgsd-bootenv-arcan` (the installer ISO)

## Variant Structure

### Required Fields

```lua
return {
  id = "pgsd-bootenv-arcan",        -- Unique identifier
  name = "PGSD Boot Environment",    -- Human-readable name
  version = "0.1.0",                 -- Version string
  pkg_lists = { "bootenv/base", ... }, -- Packages for live environment
  overlays = { "common", ... },      -- Overlays for live environment
  images_dir = "/usr/local/share/pgsd/images", -- Where images are embedded
}
```

### Optional Fields

```lua
  -- Boot environment configuration
  bootenv = {
    live_user = {
      username = "pgsd",
      password = "pgsd",
      auto_login = true,
    },

    iso = {
      volume_id = "PGSD_BOOT",
      publisher = "PGSD Foundation",
    },

    services = { "sshd", "ntpd", "dbus" },

    arcan_target = {
      name = "pgsd-installer",
      path = "/usr/local/bin/pgsd-inst",
    },
  },

  -- System requirements
  system_requirements = {
    ram_min = "2GB",
    disk_space = "10GB",
  },

  -- Images to embed
  embedded_images = { "pgsd-desktop" },
}
```

## Package Lists for Boot Environments

Boot environment package lists are prefixed with `bootenv/` to distinguish them from installed system packages.

### Standard Boot Environment Lists

**bootenv/base**:
- Minimal FreeBSD system for booting
- ~400 MB (lighter than full system)
- Includes: kernel, basic userland, pkg

**bootenv/arcan**:
- Arcan display server for live environment
- Similar to desktop/arcan but may exclude unnecessary libs

**bootenv/durden**:
- Durden window manager for live environment

**bootenv/utils**:
- Live environment utilities
- Includes: terminal emulator, browser, file manager

**bootenv/network**:
- WiFi and network tools
- Includes: wpa_supplicant, dhcpcd, curl

**bootenv/graphics**:
- Graphics drivers for live boot
- Essential drivers only (reduce ISO size)

**installer/pgsd-inst**:
- The TUI installer binary
- Dependencies for installation (zfs, gpart, etc.)

**system/disk-tools**:
- Disk management utilities
- Includes: gpart, zfs, beadm, geli, diskinfo

## Overlays for Boot Environments

**common**: Base system configuration (shared with images)

**desktop**: Desktop environment settings (shared with images)

**arcan**: Arcan/Durden configuration (shared with images)

**bootenv**: Boot environment specific overlay
  - Live user setup scripts
  - Auto-login configuration
  - Welcome scripts
  - Installer launchers
  - Boot environment initialization

## Boot Environment Features

### Live User Setup

The bootenv overlay creates a live user with:
- Username/password: `pgsd/pgsd`
- Auto-login on console (tty1)
- Groups: wheel, operator, video, audio
- Shell: ZSH with helpful aliases
- Home directory with welcome scripts

### Installer Integration

The installer is integrated into the boot environment:

1. **Command Line**: `sudo pgsd-install` or `sudo pgsd-inst`
2. **Desktop Menu**: "PGSD Installer" application (Arcan target)
3. **Launcher Scripts**: Helper scripts for easy access

### Embedded Images

System images are embedded in the ISO at `/usr/local/share/pgsd/images`:

```
/usr/local/share/pgsd/images/
  pgsd-desktop/
    root.zfs.xz
    efi.img
    manifest.toml
```

The installer reads from this directory to present available images.

### Network Configuration

Boot environment includes network support:
- DHCP auto-configuration for wired
- WiFi tools (wpa_supplicant)
- Helper scripts for network setup

### Memory Filesystems

Live environment uses tmpfs for writable areas:
- `/tmp` - 512MB tmpfs
- `/var` - 256MB tmpfs
- Changes don't persist (live environment)

## Build Process

### Prerequisites

Images must be built before creating ISOs:

```bash
# Build system images first
pgsdbuild image pgsd-desktop

# Verify artifacts exist
ls artifacts/pgsd-desktop/
# Should show: root.zfs.xz, efi.img, manifest.toml
```

### Building ISOs

```bash
# Build bootable ISO
pgsdbuild iso pgsd-bootenv-arcan

# With verbose output
pgsdbuild -v iso pgsd-bootenv-arcan

# Keep work directory
pgsdbuild --keep-work iso pgsd-bootenv-arcan

# Custom ISO output directory
pgsdbuild --iso-dir /custom/path iso pgsd-bootenv-arcan
```

### Build Output

ISO artifact location: `iso/pgsd-bootenv-arcan.iso`

ISO contains:
- Boot environment root filesystem
- Bootloader (UEFI + BIOS)
- Embedded system images
- Installer application
- Live user environment

## ISO Boot Process

1. **Boot from ISO**
   - UEFI or BIOS boot
   - FreeBSD kernel loads
   - Boot environment services start

2. **Auto-Login**
   - Live user `pgsd` auto-login on tty1
   - Welcome message displayed
   - Shell profile loads

3. **Arcan/Durden Launch**
   - Arcan starts automatically from `.profile`
   - Durden window manager initializes
   - Desktop environment ready

4. **Installer Access**
   - Desktop menu: "PGSD Installer"
   - Terminal: `sudo pgsd-install`
   - Direct: `sudo pgsd-inst`

5. **Installation**
   - Select system image
   - Select target disk
   - Confirm and install
   - Reboot to installed system

## Example: Multi-Image ISO

```lua
-- variants/pgsd-bootenv-complete.lua
return {
  id = "pgsd-bootenv-complete",
  name = "PGSD Complete Installation Media",
  version = "0.1.0",

  pkg_lists = {
    "bootenv/base",
    "bootenv/arcan",
    "bootenv/durden",
    "bootenv/utils",
    "installer/pgsd-inst",
    "bootenv/network",
    "bootenv/graphics",
    "system/disk-tools",
  },

  overlays = {
    "common",
    "desktop",
    "arcan",
    "bootenv",
  },

  images_dir = "/usr/local/share/pgsd/images",

  -- Multiple images embedded
  embedded_images = {
    "pgsd-desktop",
    "pgsd-server",
    "pgsd-minimal",
    "pgsd-workstation",
  },

  bootenv = {
    iso = {
      volume_id = "PGSD_COMPLETE",
    },
  },
}
```

## Customization

### Custom Live User

Modify `overlays/bootenv/usr/local/etc/rc.d/pgsd_bootenv_init`:

```sh
# Create live user with custom settings
pw useradd myuser -u 1000 -g wheel \
    -s /bin/sh -m -h - <<EOF
mypassword
EOF
```

### Custom Welcome Message

Edit `overlays/bootenv/usr/local/bin/pgsd-welcome`

### Custom Installer Launcher

Edit `overlays/bootenv/usr/local/bin/pgsd-install`

### Additional Services

Edit `overlays/bootenv/etc/rc.conf.d/bootenv`:

```sh
# Enable additional service
myservice_enable="YES"
```

## Boot Environment Package Lists

Create `pkg-lists/bootenv/` (future enhancement):

```
pkg-lists/bootenv/
  base.txt          # Minimal FreeBSD base
  arcan.txt         # Arcan for live environment
  durden.txt        # Durden for live environment
  utils.txt         # Live environment utilities
  network.txt       # Network tools
  graphics.txt      # Graphics drivers
```

## Testing

### Test in VM

```bash
# Build ISO
pgsdbuild iso pgsd-bootenv-arcan

# Test with QEMU
qemu-system-x86_64 \
    -cdrom iso/pgsd-bootenv-arcan.iso \
    -boot d \
    -m 4G \
    -enable-kvm

# Or VirtualBox
VBoxManage createvm --name "PGSD-Test" --ostype FreeBSD_64 --register
VBoxManage modifyvm "PGSD-Test" --memory 4096 --vram 128
VBoxManage storagectl "PGSD-Test" --name "IDE" --add ide
VBoxManage storageattach "PGSD-Test" --storagectl "IDE" \
    --port 0 --device 0 --type dvddrive \
    --medium iso/pgsd-bootenv-arcan.iso
VBoxManage startvm "PGSD-Test"
```

### Verify ISO Contents

```bash
# Mount ISO (FreeBSD)
sudo mount -t cd9660 /dev/cd0 /mnt

# Or on Linux
sudo mount -o loop iso/pgsd-bootenv-arcan.iso /mnt

# Check contents
ls /mnt
ls /mnt/usr/local/share/pgsd/images
ls /mnt/usr/local/bin/pgsd-inst
```

## Troubleshooting

### ISO Won't Boot

- Check BIOS/UEFI boot order
- Verify ISO is properly written to USB/DVD
- Try legacy BIOS mode if UEFI fails

### Network Not Working

```bash
# In live environment
ifconfig                     # List interfaces
dhclient em0                # DHCP on interface em0
wpa_supplicant -i wlan0 -c /etc/wpa_supplicant.conf
```

### Arcan Won't Start

```bash
# Check logs
dmesg | less
cat /var/log/Xorg.0.log     # If applicable

# Try manual start
arcan durden
```

### Installer Not Found

```bash
# Verify installer exists
ls -l /usr/local/bin/pgsd-inst
which pgsd-install

# Check images directory
ls /usr/local/share/pgsd/images/
```

## Best Practices

1. **Keep Boot Environment Minimal**: Only include packages needed for installation
2. **Test Before Release**: Always test ISOs in VMs before distribution
3. **Document Requirements**: Clearly state RAM, disk, and network requirements
4. **Version Everything**: Use semantic versioning for variants
5. **Include Rescue Tools**: Add utilities for system recovery

## See Also

- [Image Recipes Guide](IMAGE_RECIPES.md)
- [Package Lists Reference](PACKAGE_LISTS.md)
- [Build Pipeline](BUILD_PIPELINE.md)
- [Architecture](ARCHITECTURE.md)
