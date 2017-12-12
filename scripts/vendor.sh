#!/bin/bash
set -e -x

dep ensure -v -vendor-only
prune
