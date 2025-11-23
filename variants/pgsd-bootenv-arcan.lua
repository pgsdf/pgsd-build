-- PGSD Boot Environment Variant
--
-- This variant creates a bootable ISO with:
-- - Live Arcan/Durden desktop environment
-- - pgsd-inst installer (TUI)
-- - Embedded system images for installation
-- - Installer registered as Arcan target
--
-- The resulting ISO will be:
-- - Bootable ISO: iso/pgsd-bootenv-arcan.iso
-- - Contains: installer, images, boot environment
--
-- Usage:
--   1. Boot from ISO
--   2. Arcan/Durden launches automatically
--   3. Launch "PGSD Installer" from Durden menu (or from terminal)
--   4. Select image, disk, and install

return {
  -- Variant identification
  id = "pgsd-bootenv-arcan",
  name = "PGSD Boot Environment (Arcan)",
  version = "0.1.0",

  -- Package lists to install in the boot environment
  -- These packages create a minimal live environment with Arcan/Durden
  pkg_lists = {
    -- Base FreeBSD system for boot environment
    -- Maps to: FreeBSD-kernel, FreeBSD-runtime, pkg, sudo, zsh
    "bootenv/base",

    -- Arcan display server (live environment)
    -- Maps to: arcan, arcan-wayland, mesa-libs, vulkan-loader
    "bootenv/arcan",

    -- Durden window manager (live environment)
    -- Maps to: durden, lua54
    "bootenv/durden",

    -- Essential live environment utilities
    -- Maps to: foot (terminal), firefox (browser), file manager
    "bootenv/utils",

    -- Installer binary
    -- The pgsd-inst TUI installer
    "installer/pgsd-inst",

    -- Network support for boot environment
    -- Maps to: wpa_supplicant, dhcpcd, curl
    "bootenv/network",

    -- Graphics drivers for live boot
    -- Maps to: drm-kmod, gpu-firmware-*
    "bootenv/graphics",

    -- Disk management tools
    -- Maps to: gpart, zfs, beadm, geli
    "system/disk-tools",
  },

  -- Filesystem overlays to apply to boot environment
  overlays = {
    -- Common PGSD configurations
    -- Contains: /etc/motd, /etc/rc.conf.d/*, /boot/loader.conf
    "common",

    -- Desktop configurations for live environment
    -- Contains: user environment, auto-start scripts
    "desktop",

    -- Arcan/Durden configurations
    -- Contains: Durden theme, Arcan settings
    "arcan",

    -- Boot environment specific overlay
    -- Contains: installer registration, live user setup, auto-login
    "bootenv",
  },

  -- Directory where system images are embedded in the ISO
  -- The installer will read images from this location
  images_dir = "/usr/local/share/pgsd/images",

  -- Boot environment configuration (metadata)
  bootenv = {
    -- Live user configuration
    live_user = {
      username = "pgsd",
      password = "pgsd",          -- Default password for live user
      shell = "/usr/local/bin/zsh",
      groups = { "wheel", "operator", "video", "audio" },
      auto_login = true,           -- Auto-login on tty1
    },

    -- ISO configuration
    iso = {
      volume_id = "PGSD_BOOT",
      publisher = "Pacific Grove Software Distribution Foundation",
      boot_mode = "uefi",          -- UEFI boot support
      legacy_boot = true,          -- Also support BIOS boot
    },

    -- Services to enable in boot environment
    services = {
      "sshd",                      -- SSH access to live environment
      "ntpd",                      -- Time synchronization
      "dbus",                      -- Desktop services
      "sndiod",                    -- Audio
      "wpa_supplicant",            -- WiFi
      "dhcpcd",                    -- DHCP client for network
    },

    -- Arcan/Durden launcher configuration
    arcan_target = {
      name = "pgsd-installer",
      type = "BINARY",
      path = "/usr/local/bin/pgsd-inst",
      icon = "/usr/local/share/pgsd/icons/installer.png",
      description = "PGSD System Installer",
      categories = { "System", "Installation" },
    },
  },

  -- Memory requirements for boot environment
  system_requirements = {
    ram_min = "2GB",               -- Minimum RAM for live environment
    ram_recommended = "4GB",       -- Recommended for smooth experience
    disk_space = "10GB",           -- Minimum disk for installation
  },

  -- Images to include in ISO
  -- These are automatically copied from artifacts/ directory
  embedded_images = {
    "pgsd-desktop",                -- Main desktop image
    -- Add more images as needed:
    -- "pgsd-server",
    -- "pgsd-minimal",
  },

  -- Build configuration hints
  build = {
    -- ISO size estimate
    iso_size = "1.5GB",            -- Approximate ISO size

    -- Compression settings
    compression = "xz",            -- Compression algorithm for ISO
    compression_level = 6,         -- Balance between size and build time

    -- Include source for offline installation
    include_sources = false,       -- Set to true to include package sources
  },
}
