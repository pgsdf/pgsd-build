# PGSD Boot Environment User Profile
# ~/.profile for the live user in boot environment

# Show welcome message on first login
if [ ! -f ~/.pgsd_welcome_shown ]; then
    pgsd-welcome
    touch ~/.pgsd_welcome_shown
fi

# Source the standard profile
if [ -f /etc/skel/.profile ]; then
    . /etc/skel/.profile
fi

# Boot environment specific settings
export PGSD_BOOTENV=1
export PGSD_LIVE_MODE=1

# Installer is available at:
export PGSD_INSTALLER="/usr/local/bin/pgsd-inst"
export PGSD_IMAGES_DIR="/usr/local/share/pgsd/images"

# Add helpful alias for installation
alias install-pgsd='sudo pgsd-install'
alias installer='sudo pgsd-inst'

# Network helpers
alias wifi-connect='sudo wpa_passphrase'
alias wifi-restart='sudo service wpa_supplicant restart && sudo service dhcpcd restart'
alias net-status='ifconfig; echo ""; netstat -rn'

# Disk helpers
alias disks='diskinfo -v'
alias partitions='gpart show'

# System info
alias sysinfo='uname -a; sysctl hw.model hw.ncpu hw.physmem'

# Clear screen and show welcome on new terminal
clear
echo "PGSD Boot Environment - Live System"
echo "Type 'pgsd-welcome' for detailed information"
echo "Type 'sudo pgsd-install' to begin installation"
echo ""
