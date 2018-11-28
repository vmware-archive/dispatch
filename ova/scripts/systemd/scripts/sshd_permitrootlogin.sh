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

# remove all default settings
sed -i "/^PermitRootLogin.*/d" /etc/ssh/sshd_config

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# shellcheck source=./ovfenv_wrapper.sh
source "${SCRIPT_DIR}/ovfenv_wrapper.sh"

# TODO use environment file
PERMIT="$(ovfenv --key appliance.permit_root_login)"

# Only accepts True as yes
if [ "${PERMIT,,}" == "true" ]; then
  echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
else
  echo "PermitRootLogin no" >> /etc/ssh/sshd_config
fi
