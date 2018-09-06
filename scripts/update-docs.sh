#!/bin/bash

# Simple script for updating the docs on release
# Usage: ./update-docs [release]

DOCS_DIR=$(mktemp -d)
pushd $DOCS_DIR
mkdir gh-pages
pushd gh-pages
git clone git@github.com:vmware/dispatch.git
cd dispatch && git checkout origin/gh-pages
popd
mkdir master
pushd master
git clone git@github.com:vmware/dispatch.git
cd dispatch && git checkout origin/master
popd
rsync -av --exclude '\.git*' --exclude 'README.md' master/dispatch/docs/ gh-pages/dispatch --delete
cd gh-pages/dispatch
git add .
git commit -m "updating docs for release ${1}"
git checkout -b gh-pages
git push --force-with-lease origin gh-pages
popd
rm -rf $DOCS_DIR
