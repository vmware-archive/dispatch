#!/bin/bash

# Use with a systemd serviceType=oneshot  RemainAfterExit=yes

file=$1
intervalInSeconds=30 # 1 minute
while [[ ! -e $file ]]; do
    echo "File $file does not yet exist. sleeping $intervalInSeconds seconds..."
    sleep $intervalInSeconds
done