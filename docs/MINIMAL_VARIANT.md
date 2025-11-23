# PGSD Minimal Boot Environment

This document describes the minimal boot environment variant designed for fast, lightweight installation.

## Overview

The `pgsd-bootenv-minimal` variant is a stripped-down bootable ISO optimized for:

- **Small ISO Size**: ~500-800 MB (vs ~1.5 GB for full variant)
- **Fast Boot Time**: Minimal services, reduced boot delay
- **Low Memory Usage**: 1GB RAM minimum (vs 2GB for full variant)
- **Installer-Only**: No extra utilities, browsers, or applications
- **Wired Network Only**: No WiFi support to reduce size

## Use Cases

Perfect for:
- Virtual machine installations
- Server deployments
- Systems with limited RAM (1-2 GB)
- Fast network installations
- Automated deployments
- Testing and development

**Not suitable for:**
- WiFi-only systems (use full variant)
- Trying PGSD before installing (use full variant)
- Systems requiring rescue utilities (use full variant)

## Differences from Full Variant

| Feature | Full Variant | Minimal Variant |
|---------|--------------|-----------------|
| ISO Size | ~1.5 GB | ~600 MB |
| RAM Required | 2 GB | 1 GB |
| Boot Time | ~30-60s | ~15-30s |
| Terminal | foot + others | foot only |
| Browser | Firefox | None |
| File Manager | pcmanfm | None |
| WiFi Support | Yes | No (wired only) |
| SSH Server | Yes | No |
| Audio | Yes | No |
| Remote Access | Yes | No |
| Time Sync | Yes (ntpd) | No |
| Desktop Services | dbus | None |
| Logging | Full | Disabled |
| Auto-launch | Manual | Installer auto-starts |

## Package Lists

### Minimal Package Lists

**bootenv/minimal-base** (~250 MB):
- FreeBSD kernel (essential only)
- Minimal userland (no extras)
- pkg (package manager)
- sh (shell - no zsh/bash)
- Basic utilities only

**bootenv/minimal-arcan** (~150 MB):
- Arcan display server (core only)
- Essential libraries (mesa, vulkan)
- No Wayland compatibility layer
- No extra tools

**bootenv/minimal-durden** (~30 MB):
- Durden window manager (minimal config)
- Lua runtime (minimal)
- No extra applets

**bootenv/minimal-terminal** (~10 MB):
- foot terminal emulator only
- No alternatives

**installer/pgsd-inst** (~20 MB):
- TUI installer binary
- Installation dependencies

**bootenv/minimal-disk-tools** (~30 MB):
- gpart (partitioning)
- zfs (filesystem tools)
- Basic disk utilities
- No beadm, no geli

**bootenv/minimal-network** (~20 MB):
- dhcpcd (DHCP client)
- Basic network tools
- No WiFi (wpa_supplicant excluded)
- No NetworkManager

**bootenv/minimal-graphics** (~40 MB):
- Generic framebuffer/VESA only
- Or single driver detection
- No firmware bundles

**Total: ~550 MB installed**

## Configuration

### Boot Loader

**Auto-boot**: 1 second delay (vs 3 seconds)
**Resolution**: 1024x768 default (lower than full)
**Messages**: Muted (quiet boot)
**Timer Frequency**: 100 Hz (vs 1000 Hz for responsiveness)

### Services

**Enabled:**
- dhcpcd (wired network only)

**Disabled:**
- sshd (no remote access)
- ntpd (no time sync)
- dbus (no desktop services)
- sndiod (no audio)
- powerd (no power management)
- syslogd (no logging)

### Memory Filesystems

**tmpfs /tmp**: 256 MB (vs 512 MB)
**tmpfs /var**: 128 MB (vs 256 MB)

**Total RAM for filesystems**: ~400 MB (vs ~800 MB)

### Arcan/Durden Configuration

**Theme**: Minimal (no effects, black background)
**Animations**: Disabled
**Shadows**: Disabled
**Transparency**: Disabled
**Compositor**: Disabled
**Taskbar**: Disabled
**Menu**: Disabled
**Notifications**: Disabled

**Auto-launch**: Installer starts automatically 2 seconds after Durden loads

## User Experience

### Boot Sequence

1. **Boot** → 1 second delay
2. **Kernel Load** → Fast (minimal drivers)
3. **Auto-login** → Live user 'pgsd'
4. **Arcan Start** → Minimal configuration
5. **Installer Launch** → Automatic (2 second delay)

### Terminal Access

Press `Meta+Enter` (Windows/Super+Enter) to open terminal.

From terminal:
```bash
# Installer is already running in auto-launched window
# Or start manually:
sudo pgsd-inst

# Check network (wired only):
ifconfig
dhclient em0

# View disks:
gpart show
diskinfo -v

# Manual Arcan restart (if needed):
arcan durden
```

### Network Configuration

**Wired (DHCP)**: Automatic

**Manual IP** (if needed):
```bash
ifconfig em0 192.168.1.100 netmask 255.255.255.0
route add default 192.168.1.1
echo "nameserver 8.8.8.8" > /etc/resolv.conf
```

**WiFi**: Not supported (use full variant)

## Building

### Prerequisites

Build system image first:
```bash
pgsdbuild image pgsd-desktop
```

### Build Minimal ISO

```bash
# Build minimal ISO
pgsdbuild iso pgsd-bootenv-minimal

# Verbose output
pgsdbuild -v iso pgsd-bootenv-minimal

# Result
ls -lh iso/pgsd-bootenv-minimal.iso
```

### Build Time

Minimal variant builds faster:
- Full variant: ~5-10 minutes
- Minimal variant: ~2-5 minutes

(Times vary based on system performance)

## Testing

### QEMU

```bash
qemu-system-x86_64 \
    -cdrom iso/pgsd-bootenv-minimal.iso \
    -boot d \
    -m 1G \
    -enable-kvm \
    -net nic -net user
```

### VirtualBox

```bash
VBoxManage createvm --name "PGSD-Minimal" --ostype FreeBSD_64 --register
VBoxManage modifyvm "PGSD-Minimal" --memory 1024 --vram 16
VBoxManage storagectl "PGSD-Minimal" --name "IDE" --add ide
VBoxManage storageattach "PGSD-Minimal" \
    --storagectl "IDE" --port 0 --device 0 \
    --type dvddrive --medium iso/pgsd-bootenv-minimal.iso
VBoxManage startvm "PGSD-Minimal"
```

### Real Hardware

1. Write ISO to USB:
   ```bash
   # FreeBSD
   dd if=iso/pgsd-bootenv-minimal.iso of=/dev/da0 bs=1M status=progress

   # Linux
   sudo dd if=iso/pgsd-bootenv-minimal.iso of=/dev/sdb bs=1M status=progress
   ```

2. Boot from USB
3. Wait for auto-login
4. Installer starts automatically
5. Follow installation prompts

## Performance Comparison

### Boot Time

| Environment | Full Variant | Minimal Variant |
|-------------|--------------|-----------------|
| BIOS POST | ~5s | ~5s |
| Boot Loader | 3s | 1s |
| Kernel Load | ~10s | ~8s |
| Services | ~15s | ~5s |
| Arcan Start | ~10s | ~8s |
| **Total** | **~43s** | **~27s** |

### Memory Usage

| Component | Full Variant | Minimal Variant |
|-----------|--------------|-----------------|
| Kernel | ~200 MB | ~150 MB |
| Filesystems | ~800 MB | ~400 MB |
| Arcan/Durden | ~300 MB | ~200 MB |
| Services | ~200 MB | ~50 MB |
| **Total** | **~1.5 GB** | **~800 MB** |

## Troubleshooting

### Installer Doesn't Auto-Launch

Manual start:
```bash
# Press Meta+Enter for terminal
sudo pgsd-inst
```

### Network Not Working

Check interface:
```bash
# List interfaces
ifconfig

# Start DHCP manually
dhclient em0  # Replace em0 with your interface
```

### Graphics Issues

The minimal variant uses generic framebuffer/VESA. If graphics fail:

1. Boot into single-user mode
2. Use console installer (future enhancement)
3. Or use full variant with specific drivers

### Not Enough RAM

Minimum 1GB required. If system has less:
- Reduce tmpfs sizes in /etc/rc.conf.d/minimal
- Disable tmpfs entirely (warning: slower)

### Need WiFi

Use full variant (`pgsd-bootenv-arcan`) which includes WiFi support.

## Customization

### Change Auto-Launch Delay

Edit `overlays/bootenv-minimal/usr/local/share/arcan/appl/durden/durden_minimal.lua`:
```lua
auto_launch = {
    delay = 5,  -- Change from 2 to 5 seconds
    ...
}
```

### Disable Auto-Launch

```lua
auto_launch = {
    enabled = false,  -- Change to false
    ...
}
```

### Add More Memory to Filesystems

Edit `overlays/bootenv-minimal/etc/rc.conf.d/minimal`:
```sh
tmpsize="512m"  # Increase from 256m
varsize="256m"  # Increase from 128m
```

### Enable SSH Access

Edit `overlays/bootenv-minimal/etc/rc.conf.d/minimal`:
```sh
sshd_enable="YES"
```

Rebuild ISO.

## When to Use Each Variant

### Use Minimal Variant When:
- Installing in VM with limited RAM
- Installing server systems (no desktop needed during install)
- Fast installation is priority
- Wired network available
- Familiar with FreeBSD/PGSD
- Automated/scripted installations

### Use Full Variant When:
- WiFi-only system
- Want to try PGSD before installing
- Need system rescue tools
- Unfamiliar with FreeBSD
- Want remote access during installation (SSH)
- Need browser for documentation lookup
- First-time installation

## Future Enhancements

Planned improvements for minimal variant:

1. **Console-only mode**: No Arcan/Durden, pure console installer
2. **Network detection**: Auto-detect and use available drivers
3. **Driver detection**: Detect GPU and load only needed driver
4. **Locale selection**: Choose language at boot
5. **ISO customization**: Create custom minimal ISOs with specific drivers

## See Also

- [Full Variant Documentation](VARIANTS.md)
- [Image Recipes](IMAGE_RECIPES.md)
- [Package Lists](PACKAGE_LISTS.md)
- [Build Pipeline](BUILD_PIPELINE.md)
