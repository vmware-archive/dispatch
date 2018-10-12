#!/usr/bin/env bats

# Workaround for running resource setup bats files from another bats file

set -o pipefail

# Set this ourselves in case bats call fails
EXIT_STATUS=0

setup_entity() {
    bats_file="e2e/upgrade/setup/${1}.bats"
    echo "=> ${bats_file}" >&2
    bats "${DISPATCH_ROOT}/${bats_file}" >&2
    if [[ $? -ne 0 ]]; then
        EXIT_STATUS=1
    fi
    echo >&2
}

# Load resources in specified order
set +e
setup_entity images
setup_entity functions
setup_entity apis
setup_entity events
setup_entity organizations
setup_entity secrets
setup_entity services
set -e

exit ${EXIT_STATUS}