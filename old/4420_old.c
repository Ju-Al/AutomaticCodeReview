# This file is overwritten during software install.
blacklist /var/lib/systemd
blacklist ${PATH}/systemctl
blacklist /etc/systemd/system
blacklist /etc/runlevels/
# Persistent customizations should go in a .local file.
include disable-common.local

# The following block breaks trash functionality in file managers
#read-only ${HOME}/.local
#read-write ${HOME}/.local/share
blacklist ${HOME}/.local/share/Trash

# History files in $HOME and clipboard managers
blacklist-nolog ${HOME}/.*_history
blacklist-nolog ${HOME}/.adobe
blacklist-nolog ${HOME}/.cache/greenclip*
blacklist-nolog ${HOME}/.histfile
blacklist-nolog ${HOME}/.history
blacklist-nolog ${HOME}/.kde/share/apps/klipper
blacklist-nolog ${HOME}/.kde4/share/apps/klipper
blacklist-nolog ${HOME}/.local/share/fish/fish_history
blacklist-nolog ${HOME}/.local/share/klipper
blacklist-nolog ${HOME}/.macromedia
blacklist-nolog ${HOME}/.mupdf.history
blacklist-nolog ${HOME}/.python-history
blacklist-nolog ${HOME}/.python_history
blacklist-nolog ${HOME}/.pythonhist
blacklist-nolog ${HOME}/.lesshst
blacklist-nolog ${HOME}/.viminfo
blacklist-nolog /tmp/clipmenu*

# X11 session autostart
# blacklist ${HOME}/.xpra - this will kill --x11=xpra cmdline option for all programs
blacklist ${HOME}/.Xsession
blacklist ${HOME}/.blackbox
blacklist ${HOME}/.config/autostart
blacklist ${HOME}/.config/autostart-scripts
blacklist ${HOME}/.config/awesome
blacklist ${HOME}/.config/i3
blacklist ${HOME}/.config/sway
blacklist ${HOME}/.config/lxsession/LXDE/autostart
blacklist ${HOME}/.config/openbox
blacklist ${HOME}/.config/plasma-workspace
blacklist ${HOME}/.config/startupconfig
blacklist ${HOME}/.config/startupconfigkeys
blacklist ${HOME}/.fluxbox
blacklist ${HOME}/.gnomerc
blacklist ${HOME}/.kde/Autostart
blacklist ${HOME}/.kde/env
blacklist ${HOME}/.kde/share/autostart
blacklist ${HOME}/.kde/share/config/startupconfig
blacklist ${HOME}/.kde/share/config/startupconfigkeys
blacklist ${HOME}/.kde/shutdown
blacklist ${HOME}/.kde4/env
blacklist ${HOME}/.kde4/Autostart
blacklist ${HOME}/.kde4/share/autostart
blacklist ${HOME}/.kde4/shutdown
blacklist ${HOME}/.kde4/share/config/startupconfig
blacklist ${HOME}/.kde4/share/config/startupconfigkeys
blacklist ${HOME}/.local/share/autostart
blacklist ${HOME}/.xinitrc
blacklist ${HOME}/.xprofile
blacklist ${HOME}/.xserverrc
blacklist ${HOME}/.xsession
blacklist ${HOME}/.xsessionrc
blacklist /etc/X11/Xsession.d
blacklist /etc/xdg/autostart
read-only ${HOME}/.Xauthority

# Session manager
# see #3358
#?HAS_X11: blacklist ${HOME}/.ICEauthority
#?HAS_X11: blacklist /tmp/.ICE-unix

# KDE config
blacklist ${HOME}/.cache/konsole
blacklist ${HOME}/.config/khotkeysrc
blacklist ${HOME}/.config/krunnerrc
blacklist ${HOME}/.config/kscreenlockerrc
blacklist ${HOME}/.config/ksslcertificatemanager
blacklist ${HOME}/.config/kwalletrc
blacklist ${HOME}/.config/kwinrc
blacklist ${HOME}/.config/kwinrulesrc
blacklist ${HOME}/.config/plasma-locale-settings.sh
blacklist ${HOME}/.config/plasma-org.kde.plasma.desktop-appletsrc
blacklist ${HOME}/.config/plasmashellrc
blacklist ${HOME}/.config/plasmavaultrc
blacklist ${HOME}/.kde/share/apps/kwin
blacklist ${HOME}/.kde/share/apps/plasma
blacklist ${HOME}/.kde/share/apps/solid
blacklist ${HOME}/.kde/share/config/khotkeysrc
blacklist ${HOME}/.kde/share/config/krunnerrc
blacklist ${HOME}/.kde/share/config/kscreensaverrc
blacklist ${HOME}/.kde/share/config/ksslcertificatemanager
blacklist ${HOME}/.kde/share/config/kwalletrc
blacklist ${HOME}/.kde/share/config/kwinrc
blacklist ${HOME}/.kde/share/config/kwinrulesrc
blacklist ${HOME}/.kde/share/config/plasma-desktop-appletsrc
blacklist ${HOME}/.kde4/share/apps/kwin
blacklist ${HOME}/.kde4/share/apps/plasma
blacklist ${HOME}/.kde4/share/apps/solid
blacklist ${HOME}/.kde4/share/config/khotkeysrc
blacklist ${HOME}/.kde4/share/config/krunnerrc
blacklist ${HOME}/.kde4/share/config/kscreensaverrc
blacklist ${HOME}/.kde4/share/config/ksslcertificatemanager
blacklist ${HOME}/.kde4/share/config/kwalletrc
blacklist ${HOME}/.kde4/share/config/kwinrc
blacklist ${HOME}/.kde4/share/config/kwinrulesrc
blacklist ${HOME}/.kde4/share/config/plasma-desktop-appletsrc
blacklist ${HOME}/.local/share/kglobalaccel
blacklist ${HOME}/.local/share/kwin
blacklist ${HOME}/.local/share/plasma
blacklist ${HOME}/.local/share/plasmashell
blacklist ${HOME}/.local/share/solid
blacklist /tmp/konsole-*.history
read-only ${HOME}/.cache/ksycoca5_*
read-only ${HOME}/.config/*notifyrc
read-only ${HOME}/.config/kdeglobals
read-only ${HOME}/.config/kio_httprc
read-only ${HOME}/.config/kiorc
read-only ${HOME}/.config/kioslaverc
read-only ${HOME}/.config/ksslcablacklist
read-only ${HOME}/.kde/share/apps/konsole
read-only ${HOME}/.kde/share/apps/kssl
read-only ${HOME}/.kde/share/config/*notifyrc
read-only ${HOME}/.kde/share/config/kdeglobals
read-only ${HOME}/.kde/share/config/kio_httprc
read-only ${HOME}/.kde/share/config/kioslaverc
read-only ${HOME}/.kde/share/config/ksslcablacklist
read-only ${HOME}/.kde/share/kde4/services
read-only ${HOME}/.kde4/share/apps/konsole
read-only ${HOME}/.kde4/share/apps/kssl
read-only ${HOME}/.kde4/share/config/*notifyrc
read-only ${HOME}/.kde4/share/config/kdeglobals
read-only ${HOME}/.kde4/share/config/kio_httprc
read-only ${HOME}/.kde4/share/config/kioslaverc
read-only ${HOME}/.kde4/share/config/ksslcablacklist
read-only ${HOME}/.kde4/share/kde4/services
read-only ${HOME}/.local/share/konsole
read-only ${HOME}/.local/share/kservices5
read-only ${HOME}/.local/share/kssl

# KDE sockets
blacklist ${RUNUSER}/*.slave-socket
blacklist ${RUNUSER}/kdeinit5__*
blacklist ${RUNUSER}/kdesud_*
# see #3358
#?HAS_NODBUS: blacklist ${RUNUSER}/ksocket-*
#?HAS_NODBUS: blacklist /tmp/ksocket-*

# gnome
# contains extensions, last used times of applications, and notifications
blacklist ${HOME}/.local/share/gnome-shell
# contains recently used files and serials of static/removable storage
blacklist ${HOME}/.local/share/gvfs-metadata
# no direct modification of dconf database
read-only ${HOME}/.config/dconf
blacklist ${RUNUSER}/gnome-session-leader-fifo
blacklist ${RUNUSER}/gnome-shell
blacklist ${RUNUSER}/gsconnect

# systemd
blacklist ${HOME}/.config/systemd
blacklist ${HOME}/.local/share/systemd
blacklist ${PATH}/systemctl
blacklist ${PATH}/systemd-run
blacklist ${RUNUSER}/systemd
blacklist /etc/systemd/network
blacklist /etc/systemd/system
blacklist /var/lib/systemd
# creates problems on Arch where /etc/resolv.conf is a symlink to /var/run/systemd/resolve/resolv.conf
#blacklist /var/run/systemd

# openrc
blacklist /etc/init.d/
blacklist /etc/rc.conf
blacklist /etc/runlevels/

# VirtualBox
blacklist ${HOME}/.config/VirtualBox
blacklist ${HOME}/.VirtualBox
blacklist ${HOME}/VirtualBox VMs

# GNOME Boxes
blacklist ${HOME}/.config/gnome-boxes
blacklist ${HOME}/.local/share/gnome-boxes

# libvirt
blacklist ${HOME}/.cache/libvirt
blacklist ${HOME}/.config/libvirt
blacklist ${RUNUSER}/libvirt
blacklist /var/cache/libvirt
blacklist /var/lib/libvirt
blacklist /var/log/libvirt

# OCI-Containers / Podman
blacklist ${RUNUSER}/containers
blacklist ${RUNUSER}/crun
blacklist ${RUNUSER}/libpod
blacklist ${RUNUSER}/runc
blacklist ${RUNUSER}/toolbox

# VeraCrypt
blacklist ${HOME}/.VeraCrypt
blacklist ${PATH}/veracrypt
blacklist ${PATH}/veracrypt-uninstall.sh
blacklist /usr/share/applications/veracrypt.*
blacklist /usr/share/pixmaps/veracrypt.*
blacklist /usr/share/veracrypt

# TrueCrypt
blacklist ${HOME}/.TrueCrypt
blacklist ${PATH}/truecrypt
blacklist ${PATH}/truecrypt-uninstall.sh
blacklist /usr/share/applications/truecrypt.*
blacklist /usr/share/pixmaps/truecrypt.*
blacklist /usr/share/truecrypt

# zuluCrypt
blacklist ${HOME}/.zuluCrypt
blacklist ${HOME}/.zuluCrypt-socket
blacklist ${PATH}/zuluCrypt-cli
blacklist ${PATH}/zuluMount-cli

# var
blacklist /var/cache/apt
blacklist /var/cache/pacman
blacklist /var/lib/apt
blacklist /var/lib/clamav
blacklist /var/lib/dkms
blacklist /var/lib/mysql/mysql.sock
blacklist /var/lib/mysqld/mysql.sock
blacklist /var/lib/pacman
blacklist /var/lib/upower
# blacklist /var/log - a virtual /var/log directory (mostly empty) is build up by default for
# every sandbox, unless --writable-var-log switch is activated
blacklist /var/mail
blacklist /var/opt
blacklist /var/run/acpid.socket
blacklist /var/run/docker.sock
blacklist /var/run/minissdpd.sock
blacklist /var/run/mysql/mysqld.sock
blacklist /var/run/mysqld/mysqld.sock
blacklist /var/run/rpcbind.sock
blacklist /var/run/screens
blacklist /var/spool/anacron
blacklist /var/spool/cron
blacklist /var/spool/mail

# etc
blacklist /etc/adduser.conf
blacklist /etc/anacrontab
blacklist /etc/apparmor*
blacklist /etc/cron*
blacklist /etc/default
blacklist /etc/dkms
blacklist /etc/grub*
blacklist /etc/kernel*
blacklist /etc/logrotate*
blacklist /etc/modules*
blacklist /etc/profile.d
blacklist /etc/rc.local
# rc1.d, rc2.d, ...
blacklist /etc/rc?.d
blacklist /etc/sysconfig

# hide config for various intrusion detection systems
blacklist /etc/aide
blacklist /etc/aide.conf
blacklist /etc/chkrootkit.conf
blacklist /etc/fail2ban.conf
blacklist /etc/logcheck
blacklist /etc/lynis
blacklist /etc/rkhunter.*
blacklist /etc/snort
blacklist /etc/suricata
blacklist /etc/tripwire
blacklist /var/lib/rkhunter

# Startup files
read-only ${HOME}/.antigen
read-only ${HOME}/.bash_aliases
read-only ${HOME}/.bash_login
read-only ${HOME}/.bash_logout
read-only ${HOME}/.bash_profile
read-only ${HOME}/.bashrc
read-only ${HOME}/.config/environment.d
read-only ${HOME}/.config/fish
read-only ${HOME}/.csh_files
read-only ${HOME}/.cshrc
read-only ${HOME}/.forward
read-only ${HOME}/.kshrc
read-only ${HOME}/.local/share/fish
read-only ${HOME}/.login
read-only ${HOME}/.logout
read-only ${HOME}/.mkshrc
read-only ${HOME}/.oh-my-zsh
read-only ${HOME}/.pam_environment
read-only ${HOME}/.pgpkey
read-only ${HOME}/.plan
read-only ${HOME}/.profile
read-only ${HOME}/.project
read-only ${HOME}/.tcshrc
read-only ${HOME}/.zfunc
read-only ${HOME}/.zlogin
read-only ${HOME}/.zlogout
read-only ${HOME}/.zprofile
read-only ${HOME}/.zsh.d
read-only ${HOME}/.zsh_files
read-only ${HOME}/.zshenv
read-only ${HOME}/.zshrc
read-only ${HOME}/.zshrc.local

# Remote access
blacklist ${HOME}/.rhosts
blacklist ${HOME}/.shosts
blacklist ${HOME}/.ssh/authorized_keys
blacklist ${HOME}/.ssh/authorized_keys2
blacklist ${HOME}/.ssh/environment
blacklist ${HOME}/.ssh/rc
blacklist /etc/hosts.equiv
read-only ${HOME}/.ssh/config
read-only ${HOME}/.ssh/config.d

# Initialization files that allow arbitrary command execution
read-only ${HOME}/.caffrc
read-only ${HOME}/.cargo/env
read-only ${HOME}/.dotfiles
read-only ${HOME}/.emacs
read-only ${HOME}/.emacs.d
read-only ${HOME}/.exrc
read-only ${HOME}/.gvimrc
read-only ${HOME}/.homesick
read-only ${HOME}/.iscreenrc
read-only ${HOME}/.local/lib
read-only ${HOME}/.local/share/cool-retro-term
read-only ${HOME}/.mailcap
read-only ${HOME}/.msmtprc
read-only ${HOME}/.mutt/muttrc
read-only ${HOME}/.muttrc
read-only ${HOME}/.nano
read-only ${HOME}/.npmrc
read-only ${HOME}/.pythonrc.py
read-only ${HOME}/.reportbugrc
read-only ${HOME}/.tmux.conf
read-only ${HOME}/.vim
read-only ${HOME}/.viminfo
read-only ${HOME}/.vimrc
read-only ${HOME}/.xmonad
read-only ${HOME}/.xscreensaver
read-only ${HOME}/.yarnrc
read-only ${HOME}/_exrc
read-only ${HOME}/_gvimrc
read-only ${HOME}/_vimrc
read-only ${HOME}/dotfiles

# Make directories commonly found in $PATH read-only
read-only ${HOME}/.bin
read-only ${HOME}/.cargo/bin
read-only ${HOME}/.gem
read-only ${HOME}/.local/bin
read-only ${HOME}/.luarocks
read-only ${HOME}/.npm-packages
read-only ${HOME}/.nvm
read-only ${HOME}/.rustup
read-only ${HOME}/bin

# Write-protection for desktop entries
read-only ${HOME}/.config/menus
read-only ${HOME}/.gnome/apps
read-only ${HOME}/.local/share/applications

read-only ${HOME}/.config/mimeapps.list
read-only ${HOME}/.config/user-dirs.dirs
read-only ${HOME}/.config/user-dirs.locale
read-only ${HOME}/.local/share/mime

# Write-protection for thumbnailer dir
read-only ${HOME}/.local/share/thumbnailers

# prevent access to ssh-agent
blacklist /tmp/ssh-*

# top secret
blacklist /.fscrypt
blacklist /etc/davfs2/secrets
blacklist /etc/group+
blacklist /etc/group-
blacklist /etc/gshadow
blacklist /etc/gshadow+
blacklist /etc/gshadow-
blacklist /etc/passwd+
blacklist /etc/passwd-
blacklist /etc/shadow
blacklist /etc/shadow+
blacklist /etc/shadow-
blacklist /etc/ssh
blacklist /etc/ssh/*
blacklist /home/.ecryptfs
blacklist /home/.fscrypt
blacklist ${HOME}/*.kdb
blacklist ${HOME}/*.kdbx
blacklist ${HOME}/*.key
blacklist ${HOME}/.Private
blacklist ${HOME}/.caff
blacklist ${HOME}/.cargo/credentials
blacklist ${HOME}/.cargo/credentials.toml
blacklist ${HOME}/.cert
blacklist ${HOME}/.config/hub
blacklist ${HOME}/.config/keybase
blacklist ${HOME}/.davfs2/secrets
blacklist ${HOME}/.ecryptfs
blacklist ${HOME}/.fetchmailrc
blacklist ${HOME}/.fscrypt
blacklist ${HOME}/.git-credential-cache
blacklist ${HOME}/.git-credentials
blacklist ${HOME}/.gnome2/keyrings
blacklist ${HOME}/.gnupg
blacklist ${HOME}/.kde/share/apps/kwallet
blacklist ${HOME}/.kde4/share/apps/kwallet
blacklist ${HOME}/.local/share/keyrings
blacklist ${HOME}/.local/share/kwalletd
blacklist ${HOME}/.local/share/pki
blacklist ${HOME}/.local/share/plasma-vault
blacklist ${HOME}/.msmtprc
blacklist ${HOME}/.mutt
blacklist ${HOME}/.muttrc
blacklist ${HOME}/.netrc
blacklist ${HOME}/.nyx
blacklist ${HOME}/.pki
blacklist ${HOME}/.smbcredentials
blacklist ${HOME}/.ssh
blacklist ${HOME}/.vaults
blacklist /var/backup

# cloud provider configuration
blacklist ${HOME}/.aws
blacklist ${HOME}/.boto
blacklist ${HOME}/.config/gcloud
blacklist ${HOME}/.kube
blacklist ${HOME}/.passwd-s3fs
blacklist ${HOME}/.s3cmd
blacklist /etc/boto.cfg

# system directories
blacklist /sbin
blacklist /usr/local/sbin
blacklist /usr/sbin

# system management
blacklist ${PATH}/at
blacklist ${PATH}/busybox
blacklist ${PATH}/chage
blacklist ${PATH}/chfn
blacklist ${PATH}/chsh
blacklist ${PATH}/crontab
blacklist ${PATH}/evtest
blacklist ${PATH}/expiry
blacklist ${PATH}/fusermount
blacklist ${PATH}/gksu
blacklist ${PATH}/gksudo
blacklist ${PATH}/gpasswd
blacklist ${PATH}/kdesudo
blacklist ${PATH}/ksu
blacklist ${PATH}/mount
blacklist ${PATH}/mount.ecryptfs_private
blacklist ${PATH}/nc
blacklist ${PATH}/ncat
blacklist ${PATH}/nmap
blacklist ${PATH}/newgidmap
blacklist ${PATH}/newgrp
blacklist ${PATH}/newuidmap
blacklist ${PATH}/ntfs-3g
blacklist ${PATH}/pkexec
blacklist ${PATH}/procmail
blacklist ${PATH}/sg
blacklist ${PATH}/strace
blacklist ${PATH}/su
blacklist ${PATH}/sudo
blacklist ${PATH}/tcpdump
blacklist ${PATH}/umount
blacklist ${PATH}/unix_chkpwd
blacklist ${PATH}/xev
blacklist ${PATH}/xinput

# other SUID binaries
blacklist /usr/lib/virtualbox
blacklist /usr/lib64/virtualbox

# prevent lxterminal connecting to an existing lxterminal session
blacklist /tmp/.lxterminal-socket*
# prevent tmux connecting to an existing session
blacklist /tmp/tmux-*

# disable terminals running as server resulting in sandbox escape
blacklist ${PATH}/gnome-terminal
blacklist ${PATH}/gnome-terminal.wrapper
# blacklist ${PATH}/konsole
# konsole doesn't seem to have this problem - last tested on Ubuntu 16.04
blacklist ${PATH}/lilyterm
blacklist ${PATH}/lxterminal
blacklist ${PATH}/mate-terminal
blacklist ${PATH}/mate-terminal.wrapper
blacklist ${PATH}/pantheon-terminal
blacklist ${PATH}/roxterm
blacklist ${PATH}/roxterm-config
blacklist ${PATH}/terminix
blacklist ${PATH}/tilix
blacklist ${PATH}/urxvtc
blacklist ${PATH}/urxvtcd
blacklist ${PATH}/xfce4-terminal
blacklist ${PATH}/xfce4-terminal.wrapper

# kernel files
blacklist /initrd*
blacklist /vmlinuz*

# snapshot files
blacklist /.snapshots

# flatpak
blacklist ${HOME}/.cache/flatpak
blacklist ${HOME}/.config/flatpak
noblacklist ${HOME}/.local/share/flatpak/exports
read-only ${HOME}/.local/share/flatpak/exports
blacklist ${HOME}/.local/share/flatpak/*
blacklist ${HOME}/.var
# most of the time bwrap is SUID binary
blacklist ${PATH}/bwrap
blacklist ${RUNUSER}/.dbus-proxy
blacklist ${RUNUSER}/.flatpak
blacklist ${RUNUSER}/.flatpak-cache
blacklist ${RUNUSER}/.flatpak-helper
blacklist ${RUNUSER}/app
blacklist ${RUNUSER}/doc
blacklist /usr/share/flatpak
noblacklist /var/lib/flatpak/exports
blacklist /var/lib/flatpak/*

# snap
blacklist ${RUNUSER}/snapd-session-agent.socket

# mail directories used by mutt
blacklist ${HOME}/.Mail
blacklist ${HOME}/.mail
blacklist ${HOME}/.signature
blacklist ${HOME}/Mail
blacklist ${HOME}/mail
blacklist ${HOME}/postponed
blacklist ${HOME}/sent

# kernel configuration
blacklist /proc/config.gz

# prevent DNS malware attempting to communicate with the server
# using regular DNS tools
blacklist ${PATH}/dig
blacklist ${PATH}/dlint
blacklist ${PATH}/dns2tcp
blacklist ${PATH}/dnssec-*
blacklist ${PATH}/dnswalk
blacklist ${PATH}/drill
blacklist ${PATH}/host
blacklist ${PATH}/iodine
blacklist ${PATH}/kdig
blacklist ${PATH}/khost
blacklist ${PATH}/knsupdate
blacklist ${PATH}/ldns-*
blacklist ${PATH}/ldnsd
blacklist ${PATH}/nslookup
blacklist ${PATH}/resolvectl
blacklist ${PATH}/unbound-host

# rest of ${RUNUSER}
blacklist ${RUNUSER}/*.lock
blacklist ${RUNUSER}/inaccessible
blacklist ${RUNUSER}/pk-debconf-socket
blacklist ${RUNUSER}/update-notifier.pid
