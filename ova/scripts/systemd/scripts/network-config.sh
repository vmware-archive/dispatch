#!/usr/bin/bash
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

network_conf_file=/etc/systemd/network/09-dispatch.network

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# shellcheck source=./ovfenv_wrapper.sh
source "${SCRIPT_DIR}/ovfenv_wrapper.sh"

mask2cdr () {
  set -- 0^^^128^192^224^240^248^252^254^ ${#1} ${1##*255.}
  set -- $(( ($2 - ${#3})*2 )) ${1%%${3%%.*}*}
  echo $(( $1 + (${#2}/4) ))
}

netConfig=''
dhcpOpts=''

# TODO use environment file
fqdn="$(ovfenv --key network.fqdn)"
network_address="$(ovfenv --key network.ip0)"
netmask="$(ovfenv --key network.netmask0)"
gateway="$(ovfenv --key network.gateway)"
dns="$(ovfenv --key network.DNS | sed 's/,/ /g' | tr -s ' ')"
domains="$(ovfenv --key network.searchpath)"

# static OR DHCP options.
if [[ -n $network_address || -n $netmask || -n $gateway ]]; then
  netConfig="Address=${network_address}/$(mask2cdr "$netmask")\n\
         Gateway=$gateway\n"
else
  netConfig="DHCP=ipv4\n"
fi

# always set dns if it exists
if [[ -n $dns ]]; then
  netConfig+="DNS=$dns\n"
  dhcpOpts+="UseDNS=false\n"
fi

# always set the domain if it exists
if [[ -n $domains ]]; then
  netConfig+="Domains=$domains\n"
  dhcpOpts+="UseDomains=false\n"
fi

# always set the fqdn if it exists using hostnamectl
if [[ -n $fqdn ]]; then
  hostnamectl set-hostname "$fqdn"
  dhcpOpts+="UseHostname=false\n"
fi

cat <<EOF | tee ${network_conf_file}
[Match]
Name=eth0

[Network]
$(echo -e "$netConfig")

[DHCP]
$(echo -e $dhcpOpts)
EOF

chmod 644 ${network_conf_file}
