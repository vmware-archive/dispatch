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
set -euf -o pipefail

if [ -z "${TAG}" ]; then
  echo "TAG must be set"
  exit 1
fi

mkdir -p /data/dispatch
mkdir -p /root/.ssh

mkdir -p /usr/lib/systemd/system

mkdir -p /var/lib/dispatch

# Write version files
echo "appliance=${TAG}" > /data/version
echo "appliance=${TAG}" > /etc/vmware/version
