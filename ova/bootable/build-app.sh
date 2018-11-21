#!/bin/bash
# Copyright 2017-2018 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu -o pipefail +h && [ -n "$DEBUG" ] && set -x
DIR=$(dirname "$(readlink -f "$0")")
# shellcheck source=log.sh
source "${DIR}/log.sh"

# Importing the pubkey
log2 "configuring os"
log3 "importing local gpg key"
rpm --import /etc/pki/rpm-gpg/VMWARE-RPM-GPG-KEY

log3 "setting umask to 022"
sed -i 's/umask 027/umask 022/' /etc/profile

log3 "setting root password"
echo 'root:Vmw@re!23' | chpasswd

log3 "configuring password expiration"
chage -I -1 -m 0 -M 99999 -E -1 root

log3 "configuring ${brprpl}UTC${reset} timezone"
ln --force --symbolic /usr/share/zoneinfo/UTC /etc/localtime

log3 "configuring ${brprpl}en_US.UTF-8${reset} locale"
/bin/echo "LANG=en_US.UTF-8" > /etc/locale.conf

log3 "configuring ${brprpl}haveged${reset}"
systemctl enable haveged


log3 "configuring ${brprpl}sshd${reset}"
echo "UseDNS no" >> /etc/ssh/sshd_config
systemctl enable sshd
systemctl enable sshd-keygen

log2 "running provisioners"
find script-provisioners -type f | sort -n | while read -r SCRIPT; do
  log3 "running ${brprpl}$SCRIPT${reset}"
  ./"$SCRIPT"
done;

log2 "cleaning up base os disk"
tdnf clean all

/sbin/ldconfig
/usr/sbin/pwconv
/usr/sbin/grpconv
/bin/systemd-machine-id-setup

rm /etc/resolv.conf
ln -sf ../run/systemd/resolve/resolv.conf /etc/resolv.conf

log3 "cleaning up tmp"
rm -rf /tmp/*

log3 "removing man pages"
rm -rf /usr/share/man/*
log3 "removing any docs"
rm -rf /usr/share/doc/*
log3 "removing caches"
find /var/cache -type f -exec rm -rf {} \;

log3 "removing bash history"
# Remove Bash history
unset HISTFILE
echo -n > /root/.bash_history

# Clean up log files
log3 "cleaning log files"
find /var/log -type f | while read -r f; do echo -ne '' > "$f"; done;

log3 "clearing last login information"
echo -ne '' >/var/log/lastlog
echo -ne '' >/var/log/wtmp
echo -ne '' >/var/log/btmp

log3 "resetting bashrs"
echo -ne '' > /root/.bashrc
echo -ne '' > /root/.bash_profile
echo 'shopt -s histappend' >> /root/.bash_profile
echo 'export PROMPT_COMMAND="history -a; history -c; history -r; $PROMPT_COMMAND"' >> /root/.bash_profile

# Clear SSH host keys
log3 "resetting ssh host keys"
rm -f /etc/ssh/{ssh_host_dsa_key,ssh_host_dsa_key.pub,ssh_host_ecdsa_key,ssh_host_ecdsa_key.pub,ssh_host_ed25519_key,ssh_host_ed25519_key.pub,ssh_host_rsa_key,ssh_host_rsa_key.pub}

# Zero out the free space to save space in the final image
log3 "zero out free space"
dd if=/dev/zero of=/EMPTY bs=1M  2>/dev/null || echo "dd exit code $? is suppressed"
rm -f /EMPTY

log3 "syncing fs"
sync

# seal the template
> /etc/machine-id
