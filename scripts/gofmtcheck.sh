#!/bin/bash

# Check gofmt
echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(gofmt -l . | grep -v -E 'vendor|gen')
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` to reformat code, or use"
    echo "the command: \`make difffmt\` to view the expected changes."
    exit 1
fi

exit 0
