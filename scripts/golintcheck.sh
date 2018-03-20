#!/bin/bash

golint -set_exit_status ./cmd/... ./pkg/...
exit $?
