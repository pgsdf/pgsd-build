-- PGSD Desktop Image Recipe
--
-- This image provides a complete PGSD desktop environment with:
-- - FreeBSD base system
-- - Arcan display server
-- - Durden window manager/compositor
-- - Essential desktop utilities
--
-- The resulting image will be:
-- - Compressed ZFS stream: artifacts/pgsd-desktop/root.zfs.xz
-- - EFI boot partition: artifacts/pgsd-desktop/efi.img
-- - Build manifest: artifacts/pgsd-desktop/manifest.toml
--
-- Installation target: ZFS pool with atomic snapshots and rollback

return {
  -- Image identification
  id = "pgsd-desktop",
  version = "0.1.0",

  -- ZFS configuration
  -- Pool name for the installed system
  zpool_name = "pgsd",

  -- Root dataset path (will be used for snapshots and boot environments)
  root_dataset = "pgsd/ROOT/default",

  -- Package lists to install
  -- Each entry represents a logical package set that maps to FreeBSD packages
  pkg_lists = {
    -- Base FreeBSD system packages
    -- Maps to: FreeBSD-kernel, FreeBSD-userland, pkg, sudo, zsh, etc.
    "base",

    -- Arcan display server and core libraries
    -- Maps to: arcan, arcan-wayland, arcan-tools, mesa-libs, vulkan-loader, etc.
    "desktop/arcan",

    -- Durden window manager/compositor
    -- Maps to: durden (custom package), arcan-durden, etc.
    "desktop/durden",

    -- Desktop utilities and applications
    -- Maps to: firefox, weston, foot (terminal), etc.
    "desktop/apps",

    -- Development tools (optional, can be removed for minimal installs)
    -- Maps to: git, vim, tmux, build-essential, etc.
    "dev/tools",

    -- Network and system utilities
    -- Maps to: NetworkManager, wpa_supplicant, dhclient, etc.
    "system/network",

    -- Audio subsystem (OSS/SNDIO)
    -- Maps to: sndio, sndiod, virtual_oss, etc.
    "system/audio",

    -- Graphics drivers
    -- Maps to: drm-kmod, gpu-firmware-*, xf86-video-*, etc.
    "system/graphics",
  },

  -- Filesystem overlays to apply
  -- These contain configuration files, scripts, and customizations
  overlays = {
    -- Common configurations for all PGSD systems
    -- Contains: /etc/motd, /etc/rc.conf.d/*, basic system configs
    "common",

    -- Desktop-specific configurations
    -- Contains: Arcan configs, Durden settings, user environment
    "desktop",

    -- Arcan-specific overlay
    -- Contains: /usr/local/share/arcan/appl configs, display settings
    "arcan",
  },

  -- Boot configuration (optional metadata for documentation)
  boot = {
    -- EFI loader configuration
    loader_conf = {
      -- Enable ZFS root
      ["vfs.root.mountfrom"] = "zfs:pgsd/ROOT/default",
      -- Console settings
      ["boot_multicons"] = "YES",
      ["boot_serial"] = "NO",
      -- Load required kernel modules
      ["zfs_load"] = "YES",
      ["if_wlan_load"] = "YES",
      ["wlan_wep_load"] = "YES",
      ["wlan_ccmp_load"] = "YES",
      ["wlan_tkip_load"] = "YES",
    },
  },

  -- ZFS dataset layout (optional metadata for documentation)
  datasets = {
    -- Root dataset (boot environment)
    { name = "pgsd/ROOT/default", mountpoint = "/", canmount = "noauto" },

    -- User data (not in boot environment, preserved across BEs)
    { name = "pgsd/home", mountpoint = "/home", canmount = "on" },

    -- Variable data
    { name = "pgsd/var", mountpoint = "/var", canmount = "on" },
    { name = "pgsd/var/log", mountpoint = "/var/log", canmount = "on" },
    { name = "pgsd/var/tmp", mountpoint = "/var/tmp", canmount = "on" },

    -- Optional data directories
    { name = "pgsd/usr/local", mountpoint = "/usr/local", canmount = "on" },
  },

  -- System configuration hints (optional metadata)
  system = {
    hostname = "pgsd-desktop",
    timezone = "UTC",

    -- Default services to enable
    services = {
      "sshd",
      "ntpd",
      "dbus",
      "sndiod",
      "wpa_supplicant",
    },

    -- Default users to create (handled by overlay scripts)
    users = {
      { name = "pgsd", groups = { "wheel", "operator", "video", "audio" } },
    },
  },
}
