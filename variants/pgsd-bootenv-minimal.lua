-- PGSD Minimal Boot Environment Variant
--
-- This is a stripped-down boot environment with:
-- - Minimal FreeBSD base
-- - Arcan/Durden (minimal configuration)
-- - pgsd-inst installer only
-- - Essential drivers only
--
-- Optimized for:
-- - Small ISO size (~500-800 MB vs ~1.5 GB)
-- - Fast boot time
-- - Installation purpose only (not a full live environment)
--
-- Usage:
--   1. Boot from ISO
--   2. Arcan/Durden auto-starts
--   3. Installer launches automatically or via terminal
--   4. Install and reboot

return {
  -- Variant identification
  id = "pgsd-bootenv-minimal",
  name = "PGSD Minimal Boot Environment",
  version = "0.1.0",

  -- Minimal package lists for installer-only environment
  pkg_lists = {
    -- Absolute minimum FreeBSD base
    -- Only kernel, essential userland, pkg, shell
    "bootenv/minimal-base",

    -- Arcan (minimal - display server only, no extras)
    "bootenv/minimal-arcan",

    -- Durden (minimal - window manager only)
    "bootenv/minimal-durden",

    -- Single terminal emulator (foot - lightweight)
    "bootenv/minimal-terminal",

    -- The installer
    "installer/pgsd-inst",

    -- Only essential disk tools
    -- gpart, zfs tools, basic disk utilities
    "bootenv/minimal-disk-tools",

    -- Minimal network (DHCP only, no WiFi to reduce size)
    "bootenv/minimal-network",

    -- Only current system graphics driver (detect and install one)
    -- Or generic VESA/framebuffer fallback
    "bootenv/minimal-graphics",
  },

  -- Minimal overlays
  overlays = {
    -- Common base configuration (required)
    "common",

    -- Boot environment setup (required for auto-login)
    "bootenv",

    -- Minimal Arcan config (installer launcher only)
    "bootenv-minimal",
  },

  -- Where system images are embedded
  images_dir = "/usr/local/share/pgsd/images",

  -- Boot environment configuration
  bootenv = {
    -- Live user (same as full version for consistency)
    live_user = {
      username = "pgsd",
      password = "pgsd",
      shell = "/bin/sh",              -- Use sh instead of zsh (smaller)
      groups = { "wheel", "operator" },
      auto_login = true,
    },

    -- ISO configuration
    iso = {
      volume_id = "PGSD_MIN",          -- Shorter label
      publisher = "PGSD Foundation",
      boot_mode = "uefi",
      legacy_boot = true,
    },

    -- Minimal services only
    services = {
      "dhcpcd",                        -- Network only if wired
      -- No sshd, ntpd, dbus to reduce memory and boot time
    },

    -- Arcan target for installer
    arcan_target = {
      name = "pgsd-installer",
      type = "BINARY",
      path = "/usr/local/bin/pgsd-inst",
      description = "PGSD Installer",
      auto_launch = true,              -- Auto-launch on Durden start
    },
  },

  -- Minimal system requirements
  system_requirements = {
    ram_min = "1GB",                   -- Lower than full version
    ram_recommended = "2GB",
    disk_space = "8GB",                -- Minimum installation space
  },

  -- Embedded images
  embedded_images = {
    "pgsd-desktop",                    -- Only the main image
  },

  -- Build configuration
  build = {
    iso_size = "600MB",                -- Target minimal size
    compression = "xz",
    compression_level = 9,             -- Maximum compression for smaller ISO
    include_sources = false,

    -- Optimization flags
    strip_binaries = true,             -- Strip debug symbols
    compress_files = true,             -- Compress files on ISO
    exclude_docs = true,               -- Exclude man pages, docs
    exclude_locales = true,            -- English only
  },

  -- Minimal mode specific settings
  minimal = {
    -- Auto-launch installer after Arcan starts
    auto_launch_installer = true,

    -- Reduce logging
    quiet_boot = true,

    -- No extra utilities
    no_browser = true,
    no_file_manager = true,
    no_text_editor = true,

    -- Disable unnecessary features
    no_audio = true,                   -- No audio support needed for installer
    no_bluetooth = true,
    no_printing = true,
  },
}
