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

set -euf -o pipefail

OVF_CACHE_DIR="/etc/vmware/ovfcache"

mkdir -p ${OVF_CACHE_DIR}

function ovfenv() {
    # wrapper around ovfenv to cache values if ovfenv returns empty responses
    local key=${2}
    local ovf_available=""
    local cache_file="${OVF_CACHE_DIR}/${key}"

    if /usr/bin/ovfenv >/dev/null 2>&1; then
        ovf_available="yes"
    fi

    if [[ $# -eq 4 ]]; then
        # Write mode: --key key --set value
        local value=${4}
        if [[ -n ${ovf_available} ]]; then
            /usr/bin/ovfenv --key "${key}" --set "${value}"
        fi
        echo "${value}" > "${cache_file}"
    else
        # read mode: --key key
        if [[ -n ${ovf_available} ]]; then
            # ovfenv works, read the value, cache it and return it
            local value
            value=$(/usr/bin/ovfenv --key "${key}")
            echo "${value}" > "${cache_file}"
            echo "${value}"
        elif [[ -e ${cache_file} ]]; then
            # ovfenv borked, use cache
            cat "${cache_file}"
        else
            # oh well
            echo ""
        fi
    fi
}