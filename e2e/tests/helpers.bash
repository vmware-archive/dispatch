#!/bin/bash

: ${RET_NR:=60}
: ${WAIT_T:=2}

### COMMON FUNCTIONS ###

# retry_simple takes 3 args: command [interval (secs)] [attempts (num)]
retry_simple () {
  count=0
  while [[ $count < ${3} ]]; do
    run bash -c "${1}"
    echo_to_log
    echo "$1 returned status: $status" >> ${BATS_LOG}
    if [[ $status == 0 ]]; then
      return $status
    fi
    sleep ${2}
    count=$((count+1))
  done
  flunk "command failed with exit status $status: $output after ${3} attempts"
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

export -f min_value_met

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
# usage: run_with_retry <bash command> <success output> [retries] [sleep interval]
#
# example: validate the number of subnets for a given cell ID (success output is 1, retry is 2, sleep is 30s)
#   run_with_retry "aws ec2 describe-subnets --region=$DEFAULT_AWS_REGION --filters 'Name=tag:CellID,Values=${TEST_ID}' | jq '.Subnets[].CidrBlock' | wc -l" 1 2 30
run_with_retry() {
  retries=${3:-${RET_NR}}
  wait_time=${4:-${WAIT_T}}
  for i in $(seq 1 ${retries}); do
      run bash -c "${1}"
      assert_success
      echo_to_log
      if [[ ${output} == ${2} ]]; then
          return 0
      fi
      if [[ ${output} =~ ${2} ]]; then
          return 0
      fi
      sleep ${wait_time}
  done
  echo ${1}
  echo $output
  return 1
}

# Run an arbitrary bash command until it succeeds
#
# usage: run_with_success <bash command> [retries] [sleep interval]
#
run_until_success() {
  retries=${2:-${RET_NR}}
  wait_time=${3:-${WAIT_T}}
  for i in $(seq 1 ${retries}); do
      run bash -c "${1}"
      if [[ ${status} -eq 0 ]]; then
        return 0
      fi
      sleep ${wait_time}
  done
  echo ${1}
  echo $output
  return 1
}

batch_create_images() {
  run dispatch create seed-images
  assert_success
  run bash -c "dispatch get images --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run_with_retry "dispatch get image $i --json | jq -r .status" "READY" 60 2
  done
}

batch_delete_images() {
  run dispatch delete --file ${1}
  assert_success
  run_with_retry "dispatch get base-images --json | jq '. | length'" 0 20 2
}

delete_entities(){
  echo "deleting entity type: ${1}"
  run bash -c "dispatch get ${1} --json | jq -r .[].name"
  assert_success
  for i in $output; do
      run bash -c "dispatch delete ${1} ${i}"
      echo "deleting ${1} ${i}"
  done
  run_with_retry "dispatch get ${1} --json | jq '. | length'" 0 10 2
}

create_test_svc_account(){
  name=${1}
  keydir=${2}
  openssl genrsa -out ${keydir}/private.key 4096
  openssl rsa -in ${keydir}/private.key -pubout -outform PEM -out ${keydir}/public.key

  # Create test service account
  run dispatch iam create serviceaccount \
    ${name}-user \
    --public-key ${keydir}/public.key
  echo_to_log
  assert_success

  # Create test policy for the service account
  run dispatch iam create policy \
    ${name}-policy \
    --subject ${name}-user --action "*" --resource "*"
  echo_to_log
  assert_success
}

delete_test_svc_account(){
  name=${1}

  run dispatch iam delete serviceaccount ${name}-user
  echo_to_log
  assert_success

  run dispatch iam delete policy ${name}-policy
  echo_to_log
  assert_success
}

setup_test_org(){
  org_name=${1}
  keydir=${2}
  # Create test organization
  run dispatch iam create organization ${org_name}
  echo_to_log
  assert_success

  openssl genrsa -out ${keydir}/private.key 4096
  openssl rsa -in ${keydir}/private.key -pubout -outform PEM -out ${keydir}/public.key

  # Create test service account
  run dispatch iam create serviceaccount \
    ${org_name}-user \
    --public-key ${keydir}/public.key \
    --organization ${org_name}
  echo_to_log
  assert_success

  # Create test policy for the service account
  run dispatch iam create policy \
    ${org_name}-policy \
    --subject ${org_name}-user --action "*" --resource "*" \
    --organization ${org_name}
  echo_to_log
  assert_success
}

delete_test_org(){
  org_name=${1}

  run dispatch iam delete serviceaccount ${org_name}-user --organization ${org_name}
  echo_to_log
  assert_success

  run dispatch iam delete policy ${org_name}-policy --organization ${org_name}
  echo_to_log
  assert_success

  run dispatch iam delete organization ${org_name}
}

cleanup() {
  delete_entities subscription
  delete_entities eventdriver
  delete_entities eventdrivertype
  delete_entities api
  delete_entities function
  delete_entities secret
  delete_entities image
  delete_entities base-image
}
