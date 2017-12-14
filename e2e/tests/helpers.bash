#!/bin/bash

### COMMON FUNCTIONS ###

# retry_timeout takes 2 args: command [timeout (secs)]
retry_timeout () {
  count=0
  while [[ ! `eval $1` ]]; do
    sleep 1
    count=$((count+1))
    if [[ "$count" -gt $2 ]]; then
      return 1
    fi
  done
}

# values_equal takes 2 values, both must be non-null and equal
values_equal () {
  if [[ "X$1" != "X" ]] || [[ "X$2" != "X" ]] && [[ $1 == $2 ]]; then
    return 0
  else
    return 1
  fi
}

# min_value_met takes 2 values, both must be non-null and 2 must be equal or greater than 1
min_value_met () {
  if [[ "X$1" != "X" ]] || [[ "X$2" != "X" ]] && [[ $2 -ge $1 ]]; then
    return 0
  else
    return 1
  fi
}

function echo_to_log {
  echo "$output" >> ${BATS_LOG}
}

function setup {
  echo "$BATS_TEST_NAME
----------
" >> ${BATS_LOG}
}

function teardown {
  echo "
----------

" >> ${BATS_LOG}
}

function errecho {
  >&2 echo "$@"
}

function only_if_env {
  if [[ ${!1} != "$2" ]]; then
      errecho "This test requires the $1 environment variable to be set to $2. Skipping..."
      skip
  fi
}

function require_env {
  if [[ -z ${!1} ]]; then
      errecho "This test requires the $1 environment variable to be set in order to run."
      exit 1
  fi
}

flunk() {
  { if [ "$#" -eq 0 ]; then cat -
    else echo "$@"
    fi
  } >&2
  return 1
}

fatal() {
  { if [ "$#" -eq 0 ]; then cat -
    else echo "$@"
    fi
  } >&2
  exit 1
}

assert_success() {
  if [ "$status" -ne 0 ]; then
    flunk "command failed with exit status $status: $output"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_success_or_fatal() {
  if [ "$status" -ne 0 ]; then
    fatal "command failed with exit status $status: $output"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_failure() {
  if [ "$status" -ne 1 ]; then
    flunk $(printf "expected failed exit status=1, got status=%d" $status)
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_equal() {
  if [ "$1" != "$2" ]; then
    { echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

assert_output() {
  local expected
  if [ $# -eq 0 ]; then expected="$(cat -)"
  else expected="$1"
  fi
  assert_equal "$expected" "$output"
}

assert_matches() {
  local pattern="${1}"
  local actual="${2}"

  if [ $# -eq 1 ]; then
    actual="$(cat -)"
  fi

  if ! grep -q "${pattern}" <<<"${actual}"; then
    { echo "pattern: ${pattern}"
      echo "actual:  ${actual}"
    } | flunk
  fi
}

assert_not_matches() {
  local pattern="${1}"
  local actual="${2}"

  if [ $# -eq 1 ]; then
    actual="$(cat -)"
  fi

  if grep -q "${pattern}" <<<"${actual}"; then
    { echo "pattern: ${pattern}"
      echo "actual:  ${actual}"
    } | flunk
  fi
}

assert_empty() {
  local actual="${1}"

  if [ $# -eq 0 ]; then
    actual="$(cat -)"
  fi

  if [ -n "${actual}" ]; then
    { echo "actual: ${actual}"
    } | flunk
  fi
}

assert_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    assert_equal "$2" "$(collapse_ws ${lines[$1]})"
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$(collapse_ws $line)" = "$1" ]; then return 0; fi
    done
    flunk "expected line \`$1'"
  fi
}

refute_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    local num_lines="${#lines[@]}"
    if [ "$1" -lt "$num_lines" ]; then
      flunk "output has $num_lines lines"
    fi
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$line" = "$1" ]; then
        flunk "expected to not find line \`$line'"
      fi
    done
  fi
}

assert() {
  if ! "$@"; then
    flunk "failed: $@"
  fi
}

# Run an arbitrary bash command and validate the output.
#
# usage: run_with_retry <bash command> <success output> <retries> <sleep interval>
#
# example: validate the number of subnets for a given cell ID (success output is 1, retry is 2, sleep is 30s)
#   run_with_retry "aws ec2 describe-subnets --region=$DEFAULT_AWS_REGION --filters 'Name=tag:CellID,Values=${TEST_ID}' | jq '.Subnets[].CidrBlock' | wc -l" 1 2 30
run_with_retry() {
  for i in $(seq 1 ${3}); do
      run bash -c "${1}"
      assert_success
      echo_to_log
      if [[ ${output} =~ "${2}" ]]; then
          break
      fi
      sleep ${4}
  done
  echo ${1}
  echo $output
  [[ ${output} =~ "${2}" ]]
}

batch_create_images() {
  run dispatch create --file ${BATS_TEST_DIRNAME}/${1}
  assert_success
  run bash -c "dispatch get images --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run_with_retry "dispatch get image $i --json | jq -r .status" "READY" 6 30
  done
}

batch_delete_images() {
  run dispatch delete --file ${BATS_TEST_DIRNAME}/${1}
  assert_success
  run_with_retry "dispatch get base-images --json | jq '. | length'" 0 4 5
}

delete_base_images() {
  run bash -c "dispatch get base-images --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run dispatch delete base-image $i
  done
  run_with_retry "dispatch get base-images --json | jq '. | length'" 0 4 5
}

delete_images() {
  run bash -c "dispatch get images --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run dispatch delete image $i
  done
  run_with_retry "dispatch get images --json | jq '. | length'" 0 4 5
}

delete_functions() {
  run bash -c "dispatch get functions --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run dispatch delete function $i
  done
  run_with_retry "dispatch get functions --json | jq '. | length'" 0 4 5
}

delete_secrets() {
  run bash -c "dispatch get secrets --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run dispatch delete secret $i
  done
  run_with_retry "dispatch get secrets --json | jq '. | length'" 0 4 5
}

delete_subscriptions() {
  run bash -c "dispatch get subscriptions --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run dispatch delete subscription $i
  done
  run_with_retry "dispatch get subscriptions --json | jq '. | length'" 0 4 5
}

cleanup() {
  delete_subscriptions
  delete_secrets
  delete_functions
  delete_images
  delete_base_images
}
