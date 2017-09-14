#!/bin/bash
# Easy & Dumb test check for CI jobs, currently checks ".go" files only.
#
# This will be called by the CI system (with no args) to perform checking and
# fail the job if tests are missing.
#
# Check if test files exist in all packages are fine:
#   $ ./hack/test-check.sh

set -e -o pipefail

# File patterns to exclude from check
EXCLUDE_FILE="^vendor/|^examples/|/gen/.*.go$|/mocks/.*.go$|/main.go$"
# File content which would exclude the file from the test
EXCLUDE_CONTENT="^//.*NO TEST"

# Initialize vars
ERR=false
FAIL=false

# Try and grep directorys containing go file (packages)
for pkg in $(git ls-files | egrep "\.go$" | egrep -v "${EXCLUDE_FILE}" | xargs -I {} dirname {} | sort | uniq); do
  echo "Checking package: $pkg"
  testFiles=$(find $pkg -name '*test.go')
  if [ -z "$testFiles" ]; then
    # exclude whitelisted files from the package
    whiteList=$(egrep -l -d skip "${EXCLUDE_CONTENT}" $pkg/* | sort || true)
    # find all files under source control in this package (but don't look deeper)
    all=$(find $pkg -mindepth 1 -maxdepth 1 -type f -name '*.go' | grep -f <(git ls-files $pkg) | sort || true)
    # if all files in the package are whitelisted, this package doesn't require test files
    if [ "$whiteList" = "$all" ]; then
      echo "- Whitelisted $pkg"
      continue
    fi

   echo "$(tput -T xterm setaf 1)Package $pkg missing tests$(tput -T xterm sgr0)"
    ERR=false
    FAIL=true
  fi
done

# If we failed one check, return 1
[ $FAIL == true ] && exit 1 || exit 0
