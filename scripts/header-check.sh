#!/bin/bash
# Easy & Dumb header check for CI jobs, currently checks ".go" files only.
#
# This will be called by the CI system (with no args) to perform checking and
# fail the job if headers are not correctly set. It can also be called with the
# 'fix' argument to automatically add headers to the missing files.
#
# Check if headers are fine:
#   $ ./hack/header-check.sh
# Check and fix headers:
#   $ ./hack/header-check.sh fix

set -e -o pipefail

# Header check, starts from 1, evaluated as regex, change at will.
HEADER[1]="^\/{71}$"
HEADER[2]="^\/\/ Copyright \(c\) [0-9]{4} VMware, Inc\. All Rights Reserved\.$"
HEADER[3]="^\/\/ SPDX\-License\-Identifier\: Apache\-2\.0$"
HEADER[4]="^\/{71}$"

# Initialize vars
ERR=false
FAIL=false

for file in $(git ls-files | grep "\.go$" | grep -v vendor/ | grep -v mocks/); do
  echo -n "Header check: $file... "
  for count in $(seq 1 ${#HEADER[@]}); do
    if [[ ! $(sed ${count}q\;d ${file}) =~ ${HEADER[$count]} ]]; then
      ERR=true
    fi
  done
  if [ $ERR == true ]; then
    if [[ $# -gt 0 && $1 =~ [[:upper:]fix] ]]; then
      cat ./scripts/copyright-header.txt ${file} > ${file}.new
      mv ${file}.new ${file}
      echo "$(tput -T xterm setaf 3)FIXING$(tput -T xterm sgr0)"
      git add ${file}
      ERR=false
    else
      echo "$(tput -T xterm setaf 1)FAIL$(tput -T xterm sgr0)"
      ERR=false
      FAIL=true
    fi
  else
    echo "$(tput -T xterm setaf 2)OK$(tput -T xterm sgr0)"
  fi
done

# If we failed one check, return 1
[ $FAIL == true ] && exit 1 || exit 0
