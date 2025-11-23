# PGSD ZSH Configuration
# ~/.zshrc

# History
HISTFILE=~/.zsh_history
HISTSIZE=10000
SAVEHIST=10000
setopt SHARE_HISTORY
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_SPACE

# Completion
autoload -Uz compinit
compinit

# Colors
autoload -U colors && colors

# Prompt
PS1="%{$fg[cyan]%}%n@%m%{$reset_color%}:%{$fg[blue]%}%~%{$reset_color%}$ "

# Aliases
alias ls='ls -G'
alias ll='ls -lh'
alias la='ls -lhA'
alias grep='grep --color=auto'
alias df='df -h'
alias du='du -h'

# ZFS aliases
alias zls='zfs list'
alias zsnap='zfs snapshot'
alias zclone='zfs clone'

# PGSD specific
alias pgsd-update='sudo pkg update && sudo pkg upgrade'
alias pgsd-clean='sudo pkg autoremove && sudo pkg clean'

# Functions
beadm-create() {
    if [ -z "$1" ]; then
        echo "Usage: beadm-create <name>"
        return 1
    fi
    sudo beadm create "$1"
    beadm list
}

# Key bindings
bindkey -e  # Emacs-style key bindings
bindkey "^[[3~" delete-char  # Delete key
bindkey "^[[H" beginning-of-line  # Home key
bindkey "^[[F" end-of-line  # End key
