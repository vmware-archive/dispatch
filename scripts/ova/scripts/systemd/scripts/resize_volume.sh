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

# Size threshold upon which we trigger automatic expansion of the partition,
# currently set at 10MiB
device_size_threshold=10485760

# Size threshold upon which we trigger automatic expansion of the filesystem,
# currently set at 4KiB
fs_size_threshold=4096

function check_integer {
  [[ $1 =~ ^-?[0-9]+$ ]] || ( echo "check did not return an integer, failing"; exit 1 )
}

function repartition {
  block_device=$1
  data_partition=$2
  blkdev_size=`blockdev --getsize64 ${block_device}`
  check_integer $blkdev_size

  set +e
  partition_size=`blockdev --getsize64 ${block_device}${data_partition}`
  partition_size=${partition_size:-0}
  check_integer $partition_size
  device_size_difference=`expr ${blkdev_size} - ${partition_size}`
  set -e
  check_integer $device_size_difference

  if [ $partition_size -eq 0 ]; then
    # partition doesn't exists. disk is raw, id db or log disk.
    echo "Partitioning ${block_device}"
    sgdisk -N $data_partition -c $data_partition:"Linux system" -t $data_partition:8300 $block_device
    # Reload partition table of the device
    echo "Reloading partition table"
    partprobe ${block_device}
    sleep 3
    
  elif [ $device_size_difference -gt $device_size_threshold ]; then
    # Resize Partition to use 100% of the block device
    # Keep UUIDs consistent after repartition
    echo "Repartitioning ${block_device}"
    PARTUUID=$(blkid -s PARTUUID -o value "${block_device}${data_partition}")
    sgdisk -e $block_device
    sgdisk -d $data_partition $block_device
    sgdisk -N $data_partition -c $data_partition:"Linux system" -u $data_partition:$PARTUUID -t $data_partition:8300 $block_device
    # Reload partition table of the device
    echo "Reloading partition table"
    partprobe ${block_device}
    sleep 3

  else
    echo "No repartition performed, size threshold not met"
  fi
}

function resize {
  data_partition=$1$2
  set +e
  fs_size=`dumpe2fs -h ${data_partition} |& gawk -F: '/Block count/{count=$2} /Block size/{size=$2} END{print count*size}'`
  fs_size=${fs_size:-0}
  check_integer $fs_size
  partition_size=`blockdev --getsize64 ${data_partition}`
  check_integer $partition_size
  fs_size_difference=`expr ${partition_size} - ${fs_size}`
  set -e
  check_integer $fs_size_difference

  if [ $fs_size -eq 0 ]; then
    # filesytem does not exist - must be a new partition.
    echo "Make filesystem on ${data_partition}"
    mkfs.ext4 ${data_partition}

  elif [ $fs_size_difference -gt $fs_size_threshold ]; then
    # Force a filesystem check on the data partition
    echo "Force filesystem check on ${data_partition}"
    # check will fail for root disk
    set +e
    e2fsck -pf ${data_partition}
    set -e
    # Resize the filesystem
    echo "Resize filesystem on ${data_partition}"
    resize2fs ${data_partition}

  else
    echo "No resize performed, size threshold not met"
  fi
}

function usage {
  echo -e $"Usage: $0 {repartition|resize} /dev/disk partition\nie. resize_data_volume.sh /dev/sdb 1"
  exit 1
}

if [ $# -gt 2 ]; then
  case "$1" in
    repartition)
      repartition $2 $3
      ;;
    resize)
      resize $2 $3
      ;;
    *)
      usage
  esac
else
  usage
fi
