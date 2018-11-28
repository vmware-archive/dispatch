#!/usr/bin/env bash
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
set -euf -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# shellcheck source=./ovfenv_wrapper.sh
source "${SCRIPT_DIR}/ovfenv_wrapper.sh"

declare -r mask="*******"

umask 077

ENV_FILE="/etc/vmware/environment"

HOST=""
IP_ADDRESS=""

# Keep as one string, formatted in dispatch-appliance-tls
APPLIANCE_TLS_CERT="$(ovfenv --key appliance.tls_cert | sed -E ':a;N;$!ba;s/\r{0,1}\n//g')"
APPLIANCE_TLS_PRIVATE_KEY="$(ovfenv --key appliance.tls_cert_key | sed -E ':a;N;$!ba;s/\r{0,1}\n//g')"
APPLIANCE_TLS_CA_CERT="$(ovfenv --key appliance.ca_cert | sed -E ':a;N;$!ba;s/\r{0,1}\n//g')"


# TODO split into separate unit to run before ovf-network and network-online
APPLIANCE_PERMIT_ROOT_LOGIN="$(ovfenv --key appliance.permit_root_login)"
NETWORK_FQDN="$(ovfenv --key network.fqdn)"
NETWORK_IP0="$(ovfenv --key network.ip0)"
NETWORK_NETMASK0="$(ovfenv --key network.netmask0)"
NETWORK_GATEWAY="$(ovfenv --key network.gateway)"
NETWORK_DNS="$(ovfenv --key network.DNS | sed 's/,/ /g' | tr -s ' ')"
NETWORK_SEARCHPATH="$(ovfenv --key network.searchpath)"



function detectHostname() {
  HOST=$(hostnamectl status --static) || true
  if [ -n "${HOST}" ]; then
    echo "Using hostname from 'hostnamectl status --static': ${HOST}"
    return
  fi
}

function firstboot() {
  set +e
  local tmp
  tmp="$(ovfenv --key appliance.root_pwd)"
  if [[ "$tmp" == "$mask" ]]; then
    return
  fi

  echo "root:$tmp" | chpasswd
  # Reset password expiration to 90 days by default
  chage -d $(date +"%Y-%m-%d") -m 0 -M 90 root
  set -e
}

function clearPrivate() {
  # We then obscure the root password, if the VM is reconfigured with another
  # password after deployment, we don't act on it and keep obscuring it.
  if [[ "$(ovfenv --key appliance.root_pwd)" != "$mask" ]]; then
    ovfenv --key appliance.root_pwd --set "$mask"
  fi
}

# Wait for IP addr to show up
retry=45
while [ $retry -gt 0 ]; do
  IP_ADDRESS=$(ip addr show dev eth0 | sed -nr 's/.*inet ([^ ]+)\/.*/\1/p')
  if [ -n "$IP_ADDRESS" ]; then
    break
  fi
  let retry=retry-1
  echo "IP address is null, retrying"
  sleep 1
done

detectHostname

# Modify hostname
if [ -z "${HOST}" ]; then
  echo "Hostname is null, using IP"
  HOST=${IP_ADDRESS}
fi
echo "Using hostname: ${HOST}"


{
  echo "APPLIANCE_SERVICE_UID=10000";
  echo "HOSTNAME=${HOST}";
  echo "IP_ADDRESS=${IP_ADDRESS}";
  echo "APPLIANCE_TLS_CERT=${APPLIANCE_TLS_CERT}";
  echo "APPLIANCE_TLS_PRIVATE_KEY=${APPLIANCE_TLS_PRIVATE_KEY}";
  echo "APPLIANCE_TLS_CA_CERT=${APPLIANCE_TLS_CA_CERT}";
  echo "APPLIANCE_PERMIT_ROOT_LOGIN=${APPLIANCE_PERMIT_ROOT_LOGIN}";
  echo "NETWORK_FQDN=${NETWORK_FQDN}";
  echo "NETWORK_IP0=${NETWORK_IP0}";
  echo "NETWORK_NETMASK0=${NETWORK_NETMASK0}";
  echo "NETWORK_GATEWAY=${NETWORK_GATEWAY}";
  echo "NETWORK_DNS=${NETWORK_DNS}";
  echo "NETWORK_SEARCHPATH=${NETWORK_SEARCHPATH}";
} > ${ENV_FILE}

# Only run on first boot
if [[ ! -f /etc/vmware/firstboot ]]; then
  firstboot
  date -u +"%Y-%m-%dT%H:%M:%SZ" > /etc/vmware/firstboot
fi
# Remove private values from ovfenv
clearPrivate
