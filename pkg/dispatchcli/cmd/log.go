///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	logShort = i18n.T(`Get logs for Dispatch service`)
	logLong  = i18n.T(`Get logs for Dispatch service. K8s access required.`)

	logExample = i18n.T(`
# Display event manager log and keep following
dispatch log event-manager -f

# Display last 10 lines of event manager log and keep following
dispatch log event-manager -f -t 10

# Display last 100 lines of "identity-manager" container in identity manager
dispatch log identity-manager identity-manager -t 100
`)

	components = map[string]bool{
		"api-manager":         true,
		"application-manager": true,
		"event-manager":       true,
		"function-manager":    true,
		"identity-manager":    true,
		"image-manager":       true,
		"service-manager":     true,
		// TODO, some sort of regexp to pass in the driver type and name
		"event-driver-vcenter": true,
	}

	namespace string
	follow    bool
	tail      int
)

// NewCmdLog creates a command object for Dispatch component logging
func NewCmdLog(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T(`log COMPONENT_NAME [CONTAINER_NAME] [-f] [-t <TAIL_LINES>]`),
		Short:   logShort,
		Long:    logLong,
		Example: logExample,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"logs"},
		Run: func(cmd *cobra.Command, args []string) {
			err := runLog(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Following logs")
	cmd.Flags().IntVarP(&tail, "tail", "t", -1, "Lines of recent log file to display. -1 means display all")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "dispatch", "Namespace of Dispatch install.")
	return cmd
}

func getPodName(namespace, component string) (string, error) {
	k8sGetPods := exec.Command(
		"kubectl", "-n", namespace,
		"get", "pods",
	)
	k8sGetPodsOut, err := k8sGetPods.CombinedOutput()
	if err != nil {
		return "", err
	}
	pods := strings.Split(string(k8sGetPodsOut), "\n")
	for _, podOutput := range pods {
		if strings.Contains(podOutput, component) {
			ind := strings.Index(podOutput, " ")
			if ind != -1 {
				return podOutput[:ind], nil
			}
		}
	}
	return "", errors.Errorf("Pod not found")
}

func streamPodLogs(logLine chan string, namespace, pName, cName string) {

	defer func() { close(logLine) }()

	cmdArgs := []string{
		"-n", namespace,
		"logs",
		"--tail=" + strconv.Itoa(tail),
		pName,
	}
	if cName != "" {
		cmdArgs = append(cmdArgs, cName)
	}
	if follow {
		cmdArgs = append(cmdArgs, "-f")
	}

	k8sCmd := exec.Command(
		"kubectl",
		cmdArgs...,
	)
	cmdReader, err := k8sCmd.StdoutPipe()
	if err != nil {
		return
	}
	cmdErr, err := k8sCmd.StderrPipe()
	if err != nil {
		return
	}

	logScanner := bufio.NewScanner(cmdReader)
	go func() {
		for logScanner.Scan() {
			logLine <- logScanner.Text()
		}
	}()

	k8sCmd.Start()
	errBytes, _ := ioutil.ReadAll(cmdErr)
	if errs := string(errBytes); errs != "" {
		logLine <- errs
	}

	if err = k8sCmd.Wait(); err != nil {
		logLine <- err.Error()
	}
}

func runLog(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	component := args[0]
	_, ok := components[component]
	if !ok {
		return errors.Errorf("Dispatch component `%s` not available", component)
	}

	pName, err := getPodName(namespace, component)
	if err != nil {
		return err
	}

	cName := ""
	if len(args) > 1 {
		cName = args[1]
	}

	cLogLine := make(chan string)

	go streamPodLogs(cLogLine, namespace, pName, cName)
	for line := range cLogLine {
		fmt.Println(line)
	}

	return nil
}
