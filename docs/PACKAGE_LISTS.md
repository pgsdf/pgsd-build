# PGSD Package Lists Reference

This document defines the mapping from logical package list identifiers to actual FreeBSD packages for the PGSD build system.

## Overview

Package lists are referenced in image recipes (`.lua` files) and represent logical groups of related packages. During image building, each package list is resolved to a set of FreeBSD packages that are installed via `pkg`.

## Package List Format

In the prototype, package lists are represented as string identifiers. In a production system, these would map to:

1. **Package files**: Text files listing specific pkg names
2. **Package sets**: Meta-packages bundling related functionality
3. **Build scripts**: Dynamic package resolution based on hardware/features

## Standard Package Lists

### `base`

**Purpose**: Core FreeBSD system packages required for boot and basic operation

**FreeBSD Packages**:
```
FreeBSD-kernel-generic          # FreeBSD kernel
FreeBSD-runtime                 # Base userland runtime
FreeBSD-utilities               # System utilities
pkg                             # Package manager
sudo                            # Administrative access
zsh                             # Z shell (default)
bash                            # Bourne-again shell
vim                             # Vi improved editor
tmux                            # Terminal multiplexer
openssh-portable                # SSH client/server
ca_root_nss                     # Root CA certificates
```

**Size**: ~600 MB installed

---

### `desktop/arcan`

**Purpose**: Arcan display server and core graphics stack

**FreeBSD Packages**:
```
arcan                           # Arcan display server
arcan-wayland                   # Wayland compatibility layer
arcan-tools                     # Arcan utilities
mesa-libs                       # Mesa 3D graphics libraries
mesa-dri                        # Mesa DRI drivers
vulkan-loader                   # Vulkan graphics API loader
libdrm                          # Direct Rendering Manager library
libinput                        # Input device library
libxkbcommon                    # XKB keyboard handling
pixman                          # Pixel manipulation library
cairo                           # 2D graphics library
freetype2                       # Font rendering engine
fontconfig                      # Font configuration
harfbuzz                        # Text shaping engine
```

**Size**: ~350 MB installed

**Notes**:
- Arcan may need to be built from ports or custom package repository
- Version: arcan >= 0.6.x recommended

---

### `desktop/durden`

**Purpose**: Durden window manager/compositor for Arcan

**FreeBSD Packages**:
```
durden                          # Durden window manager (custom)
arcan-durden                    # Durden integration scripts
lua54                           # Lua runtime (Durden scripting)
luarocks                        # Lua package manager
```

**Size**: ~50 MB installed

**Notes**:
- Durden is distributed with Arcan but may be packaged separately
- Located at `/usr/local/share/arcan/appl/durden`

---

### `desktop/apps`

**Purpose**: Essential desktop applications and utilities

**FreeBSD Packages**:
```
firefox                         # Web browser
chromium                        # Alternative browser (optional)
foot                            # Wayland-native terminal emulator
weston                          # Reference Wayland compositor/apps
mpv                             # Media player
imv                             # Image viewer
zathura                         # PDF viewer
zathura-pdf-mupdf              # PDF backend for zathura
pcmanfm                         # File manager
```

**Size**: ~800 MB installed

---

### `dev/tools`

**Purpose**: Development tools and build environment

**FreeBSD Packages**:
```
git                             # Version control
subversion                      # SVN (if needed)
gcc                             # GNU C compiler
llvm                            # LLVM compiler infrastructure
gmake                           # GNU make
cmake                           # CMake build system
ninja                           # Ninja build system
pkgconf                         # pkg-config replacement
gdb                             # GNU debugger
lldb                            # LLVM debugger
python3                         # Python 3 interpreter
perl5                           # Perl interpreter
```

**Size**: ~1.5 GB installed

---

### `system/network`

**Purpose**: Network management and connectivity

**FreeBSD Packages**:
```
wpa_supplicant                  # WiFi authentication
NetworkManager                  # Network management daemon (optional)
dhcpcd                          # DHCP client
openresolv                      # DNS resolver management
curl                            # HTTP client library
wget                            # Web retrieval utility
rsync                           # File synchronization
openssh-portable                # SSH (if not in base)
```

**Size**: ~100 MB installed

---

### `system/audio`

**Purpose**: Audio subsystem (SNDIO-based)

**FreeBSD Packages**:
```
sndio                           # SNDIO audio library
sndiod                          # SNDIO daemon
virtual_oss                     # Virtual OSS device
oss                             # Open Sound System
alsa-lib                        # ALSA compatibility (optional)
alsa-plugins                    # ALSA plugins
pulseaudio                      # PulseAudio (alternative, optional)
```

**Size**: ~80 MB installed

**Notes**:
- PGSD prefers SNDIO over PulseAudio for simplicity
- PulseAudio can be included for application compatibility

---

### `system/graphics`

**Purpose**: Graphics drivers and firmware

**FreeBSD Packages**:
```
drm-kmod                        # DRM kernel module
gpu-firmware-amd-kmod           # AMD GPU firmware
gpu-firmware-intel-kmod         # Intel GPU firmware
gpu-firmware-nvidia-kmod        # NVIDIA GPU firmware (optional)
xf86-video-amdgpu              # AMD open source driver
xf86-video-intel               # Intel graphics driver
libva-intel-driver             # Intel VA-API driver
mesa-dri                       # Mesa DRI drivers
```

**Size**: ~200 MB installed

**Notes**:
- Firmware packages are kernel version specific
- NVIDIA proprietary drivers available separately

---

## Custom Package Lists

Projects can define custom package lists by creating files in `pkg-lists/` (future enhancement):

```
pkg-lists/
  base.txt
  desktop/
    arcan.txt
    durden.txt
    apps.txt
  system/
    network.txt
    audio.txt
    graphics.txt
```

Each `.txt` file contains one package name per line, with comments starting with `#`.

## Package Resolution Process

1. Image recipe specifies `pkg_lists = { "base", "desktop/arcan", ... }`
2. Build system resolves each list ID to package names
3. Packages are installed via `pkg -r <root> install -y <packages>`
4. Dependencies are automatically resolved by `pkg`

## Creating New Package Lists

To add a new package list:

1. Add entry to this documentation
2. (Future) Create corresponding `.txt` file in `pkg-lists/`
3. Reference in image recipe `.lua` file
4. Build and test

## Size Estimates

Typical installation sizes:

- **Minimal** (base only): ~600 MB
- **Desktop** (base + arcan + durden + apps): ~2.5 GB
- **Full** (including dev tools): ~4.5 GB

These are uncompressed sizes. ZFS compression typically achieves 1.5-2x reduction.

## See Also

- [Image Recipe Format](ARCHITECTURE.md)
- [Build Pipeline](BUILD_PIPELINE.md)
- [FreeBSD pkg documentation](https://docs.freebsd.org/en/books/handbook/ports/)
