#!/usr/bin/env bash

set -e

# Wrapper script to run bats tests for various drivers.
# Usage: ./run-e2e.sh [subtest]

function quiet_run () {
    if [[ "$VERBOSE" == "1" ]]; then
        "$@"
    else
        "$@" &>/dev/null
    fi
}

function bring_dispatch_up {
  OS=`uname`
  echo "   Installing dispatch..."
  dispatch install --file ${DISPATCH_CONFIG} --charts-dir ${DISPATCH_ROOT}/charts --destination ${DISPATCH_ROOT} --debug
}

function cleanup_dispatch {
  echo " - Cleaning up dispatch"
}

function dispatch() {
    OS=`uname`
    ${DISPATCH_ROOT}/bin/${DISPATCH_BIN_NAME}-${OS,,} --config "${DISPATCH_CLI_CONFIG}" "$@"
}

function run_bats() {
    if [[ $INSTALL_DISPATCH == 1 ]]; then
        bring_dispatch_up
    fi

    echo "=> running clean first"
    # BATS returns non-zero to indicate the tests have failed, we shouldn't
    # necessarily bail in this case, so that's the reason for the e toggle.
    set +e
    bats "e2e/tests/clean.bats"

    for bats_file in $(find "$1" -name \*.bats | grep -v clean); do
        echo "=> $bats_file"
        bats "$bats_file"
        if [[ $? -ne 0 ]]; then
            EXIT_STATUS=1
        fi
        echo
    done
    set -e

    if [[ $INSTALL_DISPATCH == 1 ]]; then
        cleanup_dispatch
    fi
}

# Set this ourselves in case bats call fails
EXIT_STATUS=0
export BATS_FILE="$1"
export DISPATCH_CONFIG="$2"
: ${INSTALL_DISPATCH:=1}

# Check we're not running bash 3.x
if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
    echo "Bash 4.1 or later is required to run these tests"
    exit 1
fi

# If bash 4.x, check the minor version is 1 or later
if [ "${BASH_VERSINFO[0]}" -eq 4 ] && [ "${BASH_VERSINFO[1]}" -lt 1 ]; then
    echo "Bash 4.1 or later is required to run these tests"
    exit 1
fi

if [[ -z "$BATS_FILE" ]]; then
    echo "You must specify a bats test to run."
    exit 1
fi

if [[ ! -e "$BATS_FILE" ]]; then
    echo "Requested bats file or directory not found: $BATS_FILE"
    exit 1
fi

if [[ -z "$DISPATCH_CONFIG" && $INSTALL_DISPATCH == 1 ]]; then
    echo "You must specify a dispatch config file if installing dispatch (default)."
    exit 1
fi

if [[ ! -e "$DISPATCH_CONFIG" && $INSTALL_DISPATCH == 1 ]]; then
    echo "Requested dispatch config file not found: $DISPATCH_CONFIG"
    exit 1
fi

# TODO: Should the script bail out if these are set already?
export BASE_TEST_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
export DISPATCH_ROOT="$BASE_TEST_DIR/../.."
export DISPATCH_BIN_NAME=dispatch
# By default the CLI config is the generated .dispatch.json in the DISPATCH_ROOT.
# If running e2e against a cluster brought up separately, you can point to a
# different location (i.e. $HOME/.dispatch.json)
: ${DISPATCH_CLI_CONFIG:="${DISPATCH_ROOT}/.dispatch/config.json"}

export RUN_ID=$RANDOM

if [ "${CI}" = true ] ; then
  export TEST_ID="d${IMAGE_TAG}"
else
  export TEST_ID="t-$(date +%s | shasum | base64 | head -c 5)"
  export BATS_LOG="$DISPATCH_ROOT/bats.log"
fi


if [ -z "$CI" ] ; then
    # This function gets used in the integration tests, so export it.
    export -f dispatch
fi

> "$BATS_LOG"
run_bats "$BATS_FILE"
echo "finished"

exit ${EXIT_STATUS}
