# PGSD Desktop User Profile
# ~/.profile - sourced for login shells

# Environment variables
export EDITOR=vim
export PAGER=less
export VISUAL=vim

# XDG Base Directory Specification
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_DATA_HOME="$HOME/.local/share"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_STATE_HOME="$HOME/.local/state"

# Arcan environment
export ARCAN_RESOURCEPATH="/usr/local/share/arcan"
export ARCAN_STATEPATH="$XDG_DATA_HOME/arcan"
export ARCAN_FONTPATH="/usr/local/share/fonts"

# Wayland
export XDG_SESSION_TYPE=wayland
export XDG_CURRENT_DESKTOP=arcan

# Qt Wayland
export QT_QPA_PLATFORM=wayland
export QT_WAYLAND_DISABLE_WINDOWDECORATION=1

# GTK settings
export GDK_BACKEND=wayland
export MOZ_ENABLE_WAYLAND=1

# Audio
export AUDIODEVICE=snd/0               # Default audio device

# Path
export PATH="$HOME/.local/bin:$PATH"

# Source shell-specific configs
if [ -n "$BASH_VERSION" ]; then
    [ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"
elif [ -n "$ZSH_VERSION" ]; then
    [ -f "$HOME/.zshrc" ] && . "$HOME/.zshrc"
fi

# Auto-start Arcan/Durden on login (tty1 only)
if [ "$(tty)" = "/dev/ttyv0" ] && [ -z "$DISPLAY" ] && [ -z "$WAYLAND_DISPLAY" ]; then
    echo "Starting Arcan/Durden..."
    exec arcan durden
fi
