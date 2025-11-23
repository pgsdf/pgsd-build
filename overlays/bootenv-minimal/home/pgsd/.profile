# PGSD Minimal Boot Environment Profile
# Bare minimum for installer-only environment

# Clear screen
clear

# Display minimal welcome
cat <<'EOF'
╔════════════════════════════════════════╗
║   PGSD Installer                       ║
║   Minimal Boot Environment             ║
╚════════════════════════════════════════╝

Starting Arcan/Durden...
Installer will launch automatically.

To manually start: sudo pgsd-inst

EOF

# Environment for installer
export PGSD_INSTALLER="/usr/local/bin/pgsd-inst"
export PGSD_IMAGES_DIR="/usr/local/share/pgsd/images"
export PGSD_MINIMAL_MODE=1

# Minimal PATH
export PATH="/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin"

# Launch Arcan/Durden on tty1
if [ "$(tty)" = "/dev/ttyv0" ] && [ -z "$DISPLAY" ] && [ -z "$WAYLAND_DISPLAY" ]; then
    # Launch Arcan with Durden
    exec arcan durden
fi
