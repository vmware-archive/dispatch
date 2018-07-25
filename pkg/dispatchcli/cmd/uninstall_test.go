///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type TestUninstallRunner struct {
	output []string
}

func (r *TestUninstallRunner) Run(name string, args ...string) ([]byte, error) {
	out := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	r.output = append(r.output, out)
	return []byte(""), nil
}

func (r *TestUninstallRunner) GetOutput() []string {
	return r.output
}

func testUninstall(t *testing.T, args []string, expOut []string) {
	var buf bytes.Buffer
	cli := NewCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)

	cli.SetArgs(append([]string{"uninstall"}, args...))
	testUninstallRunner := TestUninstallRunner{}
	execRunner = &testUninstallRunner
	err := cli.Execute()
	assert.Equal(t, expOut, testUninstallRunner.GetOutput())
	CheckErr(err)
}

func TestRunUninstall(t *testing.T) {
	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n dispatch",
		"kubectl delete secret tls api-dispatch-tls -n kong",
		"helm delete --tiller-namespace kube-system --purge dispatch-certificate",
		"helm delete --tiller-namespace kube-system --purge ingress",
		"helm delete --tiller-namespace kube-system --purge postgres",
		"helm delete --tiller-namespace kube-system --purge docker-registry",
		"helm delete --tiller-namespace kube-system --purge openfaas",
		"kubectl delete namespace openfaas",
		"helm delete --tiller-namespace kube-system --purge jaeger",
		"kubectl delete namespace jaeger",
		"helm delete --tiller-namespace kube-system --purge riff",
		"kubectl delete namespace riff",
		"helm delete --tiller-namespace kube-system --purge transport",
		"helm delete --tiller-namespace kube-system --purge rabbitmq",
		"helm delete --tiller-namespace kube-system --purge api-gateway",
		"kubectl delete namespace kong",
		"helm delete --tiller-namespace kube-system --purge zookeeper",
		"kubectl delete namespace zookeeper",
		"helm delete --tiller-namespace kube-system --purge dispatch",
		"kubectl delete namespace dispatch",
	}
	testUninstall(t, []string{}, expOut)
}

func TestRunUninstallSingleNamespace(t *testing.T) {
	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n test-single-namespace",
		"kubectl delete secret tls api-dispatch-tls -n test-single-namespace",
		"helm delete --tiller-namespace kube-system --purge dispatch-certificate",
		"helm delete --tiller-namespace kube-system --purge ingress",
		"helm delete --tiller-namespace kube-system --purge postgres",
		"helm delete --tiller-namespace kube-system --purge docker-registry",
		"helm delete --tiller-namespace kube-system --purge openfaas",
		"helm delete --tiller-namespace kube-system --purge jaeger",
		"helm delete --tiller-namespace kube-system --purge riff",
		"helm delete --tiller-namespace kube-system --purge transport",
		"helm delete --tiller-namespace kube-system --purge rabbitmq",
		"helm delete --tiller-namespace kube-system --purge api-gateway",
		"helm delete --tiller-namespace kube-system --purge zookeeper",
		"helm delete --tiller-namespace kube-system --purge dispatch",
		"kubectl delete namespace test-single-namespace",
	}
	testUninstall(t, []string{"--single-namespace", "test-single-namespace"}, expOut)
}

func TestRunUninstallKeepNamespace(t *testing.T) {
	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n test-namespace",
		"kubectl delete secret tls api-dispatch-tls -n test-namespace",
		"helm delete --tiller-namespace kube-system --purge dispatch-certificate",
		"helm delete --tiller-namespace kube-system --purge ingress",
		"helm delete --tiller-namespace kube-system --purge postgres",
		"helm delete --tiller-namespace kube-system --purge docker-registry",
		"helm delete --tiller-namespace kube-system --purge openfaas",
		"helm delete --tiller-namespace kube-system --purge jaeger",
		"helm delete --tiller-namespace kube-system --purge riff",
		"helm delete --tiller-namespace kube-system --purge transport",
		"helm delete --tiller-namespace kube-system --purge rabbitmq",
		"helm delete --tiller-namespace kube-system --purge api-gateway",
		"helm delete --tiller-namespace kube-system --purge zookeeper",
		"helm delete --tiller-namespace kube-system --purge dispatch",
	}
	testUninstall(t, []string{"--single-namespace", "test-namespace", "--keep-namespaces"}, expOut)
}

func TestRunUninstallTillerNamespace(t *testing.T) {
	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n dispatch",
		"kubectl delete secret tls api-dispatch-tls -n kong",
		"helm delete --tiller-namespace test-tiller-namespace --purge dispatch-certificate",
		"helm delete --tiller-namespace test-tiller-namespace --purge ingress",
		"helm delete --tiller-namespace test-tiller-namespace --purge postgres",
		"helm delete --tiller-namespace test-tiller-namespace --purge docker-registry",
		"helm delete --tiller-namespace test-tiller-namespace --purge openfaas",
		"kubectl delete namespace openfaas",
		"helm delete --tiller-namespace test-tiller-namespace --purge jaeger",
		"kubectl delete namespace jaeger",
		"helm delete --tiller-namespace test-tiller-namespace --purge riff",
		"kubectl delete namespace riff",
		"helm delete --tiller-namespace test-tiller-namespace --purge transport",
		"helm delete --tiller-namespace test-tiller-namespace --purge rabbitmq",
		"helm delete --tiller-namespace test-tiller-namespace --purge api-gateway",
		"kubectl delete namespace kong",
		"helm delete --tiller-namespace test-tiller-namespace --purge zookeeper",
		"kubectl delete namespace zookeeper",
		"helm delete --tiller-namespace test-tiller-namespace --purge dispatch",
		"kubectl delete namespace dispatch",
	}
	testUninstall(t, []string{"--tiller-namespace", "test-tiller-namespace"}, expOut)
}

func TestRunUninstallWithConfigFile(t *testing.T) {

	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n dispatch",
		"kubectl delete secret tls api-dispatch-tls -n kong",
		"helm delete --tiller-namespace kube-system --purge dispatch-certificate",
		"helm delete --tiller-namespace kube-system --purge ingress",
		"helm delete --tiller-namespace kube-system --purge postgres",
		"helm delete --tiller-namespace kube-system --purge docker-registry",
		"helm delete --tiller-namespace kube-system --purge test-openfaas-release",
		"kubectl delete namespace test-openfaas-namespace",
		"helm delete --tiller-namespace kube-system --purge jaeger",
		"kubectl delete namespace jaeger",
		"helm delete --tiller-namespace kube-system --purge riff",
		"kubectl delete namespace riff",
		"helm delete --tiller-namespace kube-system --purge transport",
		"helm delete --tiller-namespace kube-system --purge rabbitmq",
		"helm delete --tiller-namespace kube-system --purge api-gateway",
		"kubectl delete namespace kong",
		"helm delete --tiller-namespace kube-system --purge zookeeper",
		"kubectl delete namespace zookeeper",
		"helm delete --tiller-namespace kube-system --purge dispatch",
		"kubectl delete namespace dispatch",
	}
	var uninstallInstallConfigYaml = `
openfaas:
  chart:
    namespace: test-openfaas-namespace
    release: test-openfaas-release
`
	f, _ := ioutil.TempFile(os.TempDir(), "test")
	defer f.Close()
	f.Write([]byte(uninstallInstallConfigYaml))
	testUninstall(t, []string{"-f", f.Name()}, expOut)
}

func TestRunUninstallServices(t *testing.T) {
	expOut := []string{
		"kubectl delete secret tls dispatch-tls -n dispatch",
		"kubectl delete secret tls api-dispatch-tls -n kong",
		"helm delete --tiller-namespace kube-system --purge dispatch-certificate",
		"helm delete --tiller-namespace kube-system --purge jaeger",
		"kubectl delete namespace jaeger",
	}
	testUninstall(t, []string{"--service", "dispatch-certificate", "--service", "jaeger"}, expOut)
}
