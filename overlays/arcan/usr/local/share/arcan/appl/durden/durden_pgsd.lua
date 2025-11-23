-- PGSD Durden Configuration Override
-- This file provides PGSD-specific defaults for Durden
-- Location: /usr/local/share/arcan/appl/durden/durden_pgsd.lua

-- This is a placeholder for PGSD-specific Durden customizations
-- In a production system, this would include:
--   - Default keybindings
--   - Theme settings (colors, fonts, etc.)
--   - Workspace configuration
--   - Display settings
--   - Audio routing
--   - Application launchers

return {
    -- Visual theme
    theme = {
        name = "pgsd-default",
        background = "#1a1a1a",
        foreground = "#e0e0e0",
        accent = "#4a90e2",
        font = "Terminus",
        font_size = 12,
    },

    -- Default keybindings (Meta = Super/Windows key)
    bindings = {
        terminal = "Meta+Return",      -- Open terminal (foot)
        launcher = "Meta+D",            -- Application launcher
        close = "Meta+Shift+Q",         -- Close window
        fullscreen = "Meta+F",          -- Toggle fullscreen
        workspace_next = "Meta+Right",  -- Next workspace
        workspace_prev = "Meta+Left",   -- Previous workspace
    },

    -- Default applications
    applications = {
        terminal = "foot",
        browser = "firefox",
        file_manager = "pcmanfm",
        editor = "vim",
    },

    -- Workspace configuration
    workspaces = {
        count = 4,
        names = { "Main", "Web", "Dev", "Media" },
    },

    -- Display settings
    display = {
        scaling = 1.0,                  -- UI scaling factor
        vsync = true,                   -- Enable vsync
        compositor = true,              -- Enable compositing
    },

    -- Audio settings
    audio = {
        backend = "sndio",              -- Use sndio
        device = "snd/0",               -- Default device
    },
}
