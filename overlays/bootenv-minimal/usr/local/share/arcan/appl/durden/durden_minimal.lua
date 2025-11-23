-- PGSD Minimal Durden Configuration
-- Ultra-minimal configuration for installer-only environment
--
-- This configuration:
-- - Disables all non-essential features
-- - Provides minimal window manager functionality
-- - Auto-launches the installer
-- - Reduces memory and CPU usage

return {
    -- Minimal theme (no fancy graphics)
    theme = {
        name = "pgsd-minimal",
        background = "#000000",        -- Black background
        foreground = "#ffffff",        -- White text
        accent = "#4a90e2",            -- Blue accent
        font = "monospace",            -- System monospace font
        font_size = 12,

        -- Disable effects
        animations = false,
        shadows = false,
        transparency = false,
        blur = false,
    },

    -- Minimal keybindings
    bindings = {
        terminal = "Meta+Return",      -- Open terminal (foot)
        close = "Meta+Shift+Q",        -- Close window

        -- Disable other bindings to reduce complexity
        launcher = nil,                -- No launcher menu
        fullscreen = nil,
        workspace_next = nil,
        workspace_prev = nil,
    },

    -- Single application: terminal
    applications = {
        terminal = "foot",             -- Only terminal available
    },

    -- Single workspace (no workspace switching)
    workspaces = {
        count = 1,
        names = { "Installer" },
    },

    -- Minimal display settings
    display = {
        scaling = 1.0,
        vsync = false,                 -- Disable vsync for speed
        compositor = false,            -- No compositing effects
    },

    -- No audio
    audio = {
        backend = "none",
    },

    -- Auto-launch configuration
    auto_launch = {
        -- Launch installer terminal on startup
        enabled = true,
        delay = 2,                     -- 2 second delay
        command = "foot -e sudo /usr/local/bin/pgsd-inst",
    },

    -- Minimal mode flags
    minimal_mode = {
        no_animations = true,
        no_shadows = true,
        no_transparency = true,
        no_compositor = true,
        no_menu = true,                -- No application menu
        no_taskbar = true,             -- No taskbar
        no_notifications = true,       -- No notification system

        -- Performance optimizations
        low_memory = true,
        reduce_cpu = true,
    },
}
