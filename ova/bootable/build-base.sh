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

function set_base() {
  src="${1}"
  rt="${2}"

  log2 "preparing install stage"
  log3 "configuring ${brprpl}tdnf${reset}"
  install -D --mode=0644 --owner=root --group=root /etc/pki/rpm-gpg/VMWARE-RPM-GPG-KEY "${rt}/etc/pki/rpm-gpg/VMWARE-RPM-GPG-KEY"
  mkdir -p "${rt}/var/lib/rpm"
  mkdir -p "${rt}/cache/tdnf"
  log3 "initializing ${brprpl}rpm db${reset}"
  rpm --root "${rt}/" --initdb
  rpm --root "${rt}/" --import "${rt}/etc/pki/rpm-gpg/VMWARE-RPM-GPG-KEY"

  log3 "configuring ${brprpl}yum repos${reset}"
  mkdir -p "${rt}/etc/yum.repos.d/"
  rm /etc/yum.repos.d/{photon,photon-updates}.repo
  cp "${DIR}"/repo/*-remote.repo /etc/yum.repos.d/
  cp -a /etc/yum.repos.d/ "${rt}/etc/"
  cp /etc/resolv.conf "${rt}/etc/"

  log3 "verifying yum and tdnf setup"
  tdnf repolist --refresh

  log3 "installing ${brprpl}photon-release-2.0${reset}"
  tdnf install --releasever 2.0 --installroot "${rt}/" --refresh -y \
    photon-release

  log3 "installing ${brprpl}filesystem bash shadow coreutils findutils${reset}"
  tdnf install --installroot "${rt}/" --refresh -y \
    filesystem bash shadow coreutils findutils

  log3 "installing ${brprpl}systemd linux-esx tdnf ca-certificates sed gzip tar${reset}"
  tdnf install --installroot "${rt}/" --refresh -y \
    systemd util-linux \
    pkgconfig dbus cpio \
    rpm tdnf \
    openssh linux-esx sed \
    gzip zip tar xz bzip2 \
    iana-etc ca-certificates \
    curl which initramfs \
    krb5 motd procps-ng \
    bc kmod libdb

  log3 "installing ${brprpl}tzdata glibc-lang vim${reset}"
  tdnf install --installroot "${rt}/" --refresh -y \
    tzdata glibc-lang vim

  log3 "installing system dependencies"
  tdnf install --installroot "${rt}/" --refresh -y \
    haveged ethtool gawk \
    socat git nfs-utils \
    cifs-utils ebtables \
    iproute2 iptables iputils \
    cdrkit xfsprogs sudo \
    lvm2 parted gptfdisk \
    docker-17.12.1-1.ph1 \
    net-tools logrotate sshpass \
    e2fsprogs open-vm-tools

  log3 "installing package dependencies"
  tdnf install --installroot "${rt}/" --refresh -y \
    docker

  log3 "installing ${brprpl}root${reset}"
  cp -a "${src}/root/." "${rt}/"
}

function usage() {
  echo "Usage: $0 -r root-location 1>&2"
  exit 1
}

while getopts "r:" flag
do
    case $flag in

        r)
            # Required. Package name
            ROOT="$OPTARG"
            ;;

        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))
if [[ -n "$*" || -z "${ROOT}" ]]; then
    usage
fi

log2 "install OS to ${ROOT}"

set_base "${DIR}" "${ROOT}"
