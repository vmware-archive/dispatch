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

# this file generates disks and grub configs for the appliance
set -eu -o pipefail +h && [ -n "$DEBUG" ] && set -x
DIR=$(dirname "$(readlink -f "$0")")
# shellcheck source=log.sh
source "${DIR}/log.sh"

function setup_grub() {
  disk=$1
  device="${1}p2"
  root=$2

  log3 "install grub to ${brprpl}${root}/boot${reset} on ${brprpl}${disk}${reset}"
  mkdir -p "${root}/boot/grub2"
  ln -sfv grub2 "${root}/boot/grub"
  grub2-install --target=i386-pc --modules "part_gpt gfxterm vbe tga png ext2" --no-floppy --force --boot-directory="${root}/boot" "$disk"

  PARTUUID=$(blkid -s PARTUUID -o value "${device}")
  BOOT_UUID=$(blkid -s UUID -o value "${device}")
  BOOT_DIRECTORY=/boot/

  log3 "configure grub"
  rm -rf "${root}/boot/grub2/fonts"
  cp "${DIR}/boot/ascii.pf2" "${root}/boot/grub2/"
  mkdir -p "${root}/boot/grub2/themes/photon"
  cp "${DIR}"/boot/splash.png "${root}/boot/grub2/themes/photon/photon.png"
  cp "${DIR}"/boot/terminal_*.tga "${root}/boot/grub2/themes/photon/"
  cp "${DIR}"/boot//theme.txt "${root}/boot/grub2/themes/photon/"
  # linux-esx tries to mount rootfs even before nvme got initialized.
  # rootwait fixes this issue
  EXTRA_PARAMS=""
  if [[ "$1" == *"nvme"* ]]; then
      EXTRA_PARAMS=rootwait
  fi

  cat > "${root}/boot/grub2/grub.cfg" << EOF
# Begin /boot/grub2/grub.cfg
set default=0
set timeout=5
search -n -u $BOOT_UUID -s
loadfont ${BOOT_DIRECTORY}grub2/ascii.pf2
insmod gfxterm
insmod vbe
insmod tga
insmod png
insmod ext2
insmod part_gpt
set gfxmode="640x480"
gfxpayload=keep
terminal_output gfxterm
set theme=${BOOT_DIRECTORY}grub2/themes/photon/theme.txt
load_env -f ${BOOT_DIRECTORY}photon.cfg
if [ -f  ${BOOT_DIRECTORY}systemd.cfg ]; then
    load_env -f ${BOOT_DIRECTORY}systemd.cfg
else
    set systemd_cmdline=net.ifnames=0
fi
set rootpartition=PARTUUID=$PARTUUID
menuentry "Photon" {
    linux ${BOOT_DIRECTORY}\$photon_linux root=\$rootpartition \$photon_cmdline \$systemd_cmdline $EXTRA_PARAMS
    if [ -f ${BOOT_DIRECTORY}\$photon_initrd ]; then
        initrd ${BOOT_DIRECTORY}\$photon_initrd
    fi
}
# End /boot/grub2/grub.cfg
EOF
}

function create_disk() {
  local img="$1"
  local disk_size="$2"
  local mp="$3"
  local boot="${4:-}"
  local ci_mode="${CI_MODE:-}"
  cd "${PACKAGE}"

  losetup -f || ( echo "Cannot setup loop devices" && exit 1 )

  log3 "allocating raw image of ${brprpl}${disk_size}${reset}"
  fallocate -l "$disk_size" -o 1024 "$img"

  log3 "wiping existing filesystems"
  sgdisk -Z -og "$img"

  part_num=1
  if [[ -n $boot ]]; then
    log3 "creating bios boot partition"
    sgdisk -n $part_num:2048:+2M -c $part_num:"BIOS Boot" -t $part_num:ef02 "$img"
    part_num=$((part_num+1))
  fi

  log3 "creating linux partition"
  sgdisk -N $part_num -c $part_num:"Linux system" -t $part_num:8300 "$img"

  log3 "reloading loop devices"
  disk=$(losetup --show -f -P "$img")

  if [[ ${ci_mode} ]]; then
    # In concourse we don't have direct access to /dev, so partitions for raw image mounted as loop device
    # are not automatically created.
    # instead, we need to manually create them using mknod.

    # drop the first line, as this is our LOOPDEV itself, but we only what the child partitions
    partitions=$(lsblk --raw --output "MAJ:MIN" --noheadings ${disk} | tail -n +2)
    counter=1
    for part in ${partitions}; do
        major=$(echo ${part} | cut -d: -f1)
        minor=$(echo ${part} | cut -d: -f2)
        if [ ! -e "${disk}p${counter}" ]; then mknod ${disk}p${counter} b ${major} ${minor}; fi
        counter=$((counter + 1))
    done
  fi

  log3 "formatting linux partition"
  mkfs.ext4 -F "${disk}p$part_num"

  log3 "mounting partition ${brprpl}${disk}p$part_num${reset} at ${brprpl}${mp}${reset}"
  mkdir -p "$mp"
  mount "${disk}p$part_num" "$mp"

    if [[ -n $boot ]]; then
    log3 "setup grup on boot disk"
    setup_grub "$disk" "$mp"
  fi
  
}

function convert() {
  local dev=$1
  local mount=$2
  local raw=$3
  local vmdk=$4
  cd "${PACKAGE}"
  log3 "unmount ${brprpl}${mount}${reset}"
  if mountpoint "${mount}" >/dev/null 2>&1; then
    umount -R "${mount}/" >/dev/null 2>&1
  fi

  log3 "release loopback device ${brprpl}${dev}${reset}"
  losetup -d "$dev"

  log3 "converting raw image ${brprpl}${raw}${reset} into ${brprpl}${vmdk}${reset}"
  qemu-img convert -f raw -O vmdk -o 'compat6,adapter_type=lsilogic,subformat=streamOptimized' "$raw" "$vmdk"
  rm "$raw"
}

function usage() {
  echo "Usage: $0 -p package-location -a [create|export] -i NAME -s SIZE -r ROOT [-i NAME -s SIZE -r ROOT]..."
  echo "  -p package-location   the working directory to use"
  echo "  -a action             the action to perform (create or export)"
  echo "  -i name               the name of an image"
  echo "  -s size               the size of an image"
  echo "  -r root               the mount point for the root of an image, relative to the package-location"
  echo "Example: $0 -p /tmp -a create -i disk1.vmdk -s 6GiB -r mnt/root -i disk2.vmdk -s 2GiB -r mnt/data"
  echo "Example: $0 -p /tmp -a create -i disk1.vmdk -i disk2.vmdk -s 6GiB -s 2GiB -r mnt/root -r mnt/data"
  exit 1
}

while getopts "p:a:i:s:r:" flag
do
    case $flag in

        p)
            # Required. Package name
            PACKAGE="$OPTARG"
            ;;

        a)
            # Required. Action: create or export
            ACTION="$OPTARG"
            ;;

        i)
            # Required, multi. Ordered list of image names
            IMAGES+=("$OPTARG")
            ;;

        s)
            # Required, multi. Ordered list of image sizes
            IMAGESIZES+=("$OPTARG")
            ;;

        r)
            # Required, multi. Ordered list of image roots
            IMAGEROOTS+=("$OPTARG")
            ;;

        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

# check there were no extra args, the required ones are set, and an equal number of each disk argument were supplied
if [[ -n "$*" || -z "${PACKAGE}" || -z "${ACTION}" || ${#IMAGES[@]} -eq 0 || ${#IMAGES[@]} -ne ${#IMAGESIZES[@]} || ${#IMAGES[@]} -ne ${#IMAGEROOTS[@]} ]]; then
    usage
fi

if [ "${ACTION}" == "create" ]; then
  log1 "create disk images"
  for i in "${!IMAGES[@]}"; do
    BOOT=""
    [ "$i" == "0" ] && BOOT="1"
    log2 "creating ${IMAGES[$i]}.img"
    create_disk "${IMAGES[$i]}.img" "${IMAGESIZES[$i]}" "${PACKAGE}/${IMAGEROOTS[$i]}" $BOOT
  done

elif [ "${ACTION}" == "export" ]; then
  log1 "export images to VMDKs"
  for i in "${!IMAGES[@]}"; do
    log2 "exporting ${IMAGES[$i]}.img to ${IMAGES[$i]}.vmdk"
    echo "export ${PACKAGE}/${IMAGES[$i]}"
    DEV=$(losetup -l -O NAME,BACK-FILE -a | tail -n +2 | grep "${PACKAGE}/${IMAGES[$i]}" | awk '{print $1}')
    convert "${DEV}" "${PACKAGE}/${IMAGEROOTS[$i]}" "${IMAGES[$i]}.img" "${IMAGES[$i]}.vmdk"
  done

  log2 "VMDK Sizes"
  log2 "$(du -h ./*.vmdk)"

else
  usage

fi
