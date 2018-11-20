#!/usr/bin/bash
# Copyright 2017 VMware, Inc. All Rights Reserved.
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
set -euf -o pipefail

# Enable systemd services
systemctl enable dispatch-mounts.target repartition.service resizefs.service
systemctl enable dispatch-appliance-environment.service
systemctl enable dispatch-appliance-ready.target
systemctl enable dispatch-appliance-tls.service
systemctl enable sshd_permitrootlogin.service
systemctl enable ovf-network.service ova-firewall.service

# Enable systemd component services
systemctl enable docker
systemctl enable dispatch-server.service

# Set our dispatch target as the default boot target
systemctl set-default dispatch-appliance.target

# Clean up temporary directories
rm -rf /tmp/* /var/tmp/*
tdnf clean all

# Warning message for client ssh
message="##########################################################################
##  Please refer to Dispatch documentation before making changes to     ##
##  the Dispatch Appliance via SSH; a better avenue may exist for the   ##
##  change you wish to make.                                            ##
##########################################################################"

# Modify ssh config to display warning message before log on
echo "$message" > "/etc/issue.net"
banner=$(grep "Banner" /etc/ssh/sshd_config)
if [ -z "$banner" ]; then
    echo "Banner /etc/issue.net" >> "/etc/ssh/sshd_config"
else
    sed -i "s/.*Banner.*/Banner\ \/etc\/issue\.net/g" /etc/ssh/sshd_config
fi

# Overwirte /etc/motd to display warning message after log on
echo "$message" > "/etc/motd"

# Disable IPv6 redirection and router advertisements in kernel settings
settings="net.ipv6.conf.all.accept_redirects = 0
net.ipv6.conf.default.accept_redirects = 0
net.ipv6.conf.all.accept_ra = 0
net.ipv6.conf.default.accept_ra = 0"
echo "$settings" > "/etc/sysctl.d/40-ipv6.conf"

# Hardening SSH configuration
afsetting=$(grep "AllowAgentForwarding" /etc/ssh/sshd_config)
if [ -z "$afsetting" ]; then
    echo "AllowAgentForwarding no" >> "/etc/ssh/sshd_config"
else
    sed -i "s/.*AllowAgentForwarding.*/AllowAgentForwarding\ no/g" /etc/ssh/sshd_config
fi

tcpfsetting=$(grep "AllowTcpForwarding" /etc/ssh/sshd_config)
if [ -z "$tcpfsetting" ]; then
    echo "AllowTcpForwarding no" >> "/etc/ssh/sshd_config"
else
    sed -i "s/.*AllowTcpForwarding.*/AllowTcpForwarding\ no/g" /etc/ssh/sshd_config
fi
