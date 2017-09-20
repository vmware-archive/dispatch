///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"log"
	"os"
)

const (
	defaultErrorExitCode = 1
)

var fatalErrHandler = fatal

// fatal prints the message (if provided) and then exits.
func fatal(msg string, code int) {
	log.SetOutput(os.Stderr)
	if msg != "" {
		log.Println(msg)
	}
	os.Exit(code)
}

// ErrExit may be passed to CheckError to instruct it to output nothing but exit with
// status code 1.
var ErrExit = fmt.Errorf("exit")

// CheckErr prints a user friendly error to STDERR and exits with a non-zero
// exit code.
func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

// checkErr formats a given error as a string and calls the passed handleErr
// func with that string and an exit code.
func checkErr(err error, handleErr func(string, int)) {
	if err == nil {
		return
	}
	// TODO: Extend with additional errors if needed:
	switch {
	case err == ErrExit:
		handleErr("", defaultErrorExitCode)
	default:
		handleErr(err.Error(), defaultErrorExitCode)
	}
}
