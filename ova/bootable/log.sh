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

# common logging functions for builder

set -eu -o pipefail +h

escape="\033[";
reset="${escape}0m";
brprpl="${reset}${escape}0;00;95m"

declare -A text_colors=(
    [default]="0;39"
    [red]="0;31"
    [green]="0;92"
    [yellow]="0;93"
    [blue]="0;94"
    [purple]="0;95"
    [cyan]="0;96"
    [white]="0;97"
)

declare -A text_styles=(
    [default]="0;0"
    [bold]="0;1"
)

function log () {
    local COLOR STYLE OPTIND
    COLOR="default"
    STYLE="default"
    break="\n"
    while getopts "c:s:n" OPTION; do
        case "${OPTION}" in
        c)
            COLOR=${OPTARG}
            ;;
        s)
            STYLE=${OPTARG}
            ;;
        n)
            break=""
            ;;

        esac
    done

    shift $((OPTIND-1))

    echo -ne "${escape}${text_styles[$STYLE]}${text_colors[$COLOR]}m$*${reset}${break}" | tee /dev/fd/3 2>/dev/null || true;
}

function log1 () {
    log -nc default -s bold "$(date +"%Y-%m-%d %H:%M:%S") "
    log -c green -s bold "[=] $*"
}

function log2 () {
    log -nc default -s bold "$(date +"%Y-%m-%d %H:%M:%S") "
    log -c yellow -s bold " [==] $*"
}

function log3() {
    log -nc default -s bold "$(date +"%Y-%m-%d %H:%M:%S") "
    log -c blue  "  [===] $*"
}
