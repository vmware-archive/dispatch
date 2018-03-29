///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package internal

import (
	"github.com/pkg/errors"
	"github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	riffcs "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"
	riffv1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned/typed/projectriff/v1"
	log "github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type RiffTalk struct {
	topics    riffv1.TopicInterface
	functions riffv1.FunctionInterface
}

func NewRiffTalk(k8sConfig, funcNamespace string) *RiffTalk {
	riffClient := newRiffClient(k8sConfig)
	return &RiffTalk{
		topics:    riffClient.ProjectriffV1().Topics(funcNamespace),
		functions: riffClient.ProjectriffV1().Functions(funcNamespace),
	}
}

func (d RiffTalk) Create(fnName, image string) error {
	topic := &v1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Name: fnName,
		},
	}
	function := &v1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: fnName,
		},
		Spec: v1.FunctionSpec{
			Protocol: "http",
			Input:    fnName,
			Container: kapi.Container{
				Image: image,
			},
		},
	}

	if _, err := d.topics.Create(topic); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "error creating topic '%s'", fnName)
		}
	}

	if _, err := d.functions.Create(function); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "error creating function '%s'", fnName)
		}
	}

	return nil
}

func (d RiffTalk) Delete(fnName string) error {
	if err := d.functions.Delete(fnName, nil); err != nil {
		if !kerrors.IsNotFound(err) {
			return errors.Wrapf(err, "error deleting function '%s'", fnName)
		}
	}

	if err := d.topics.Delete(fnName, nil); err != nil {
		if !kerrors.IsNotFound(err) {
			return errors.Wrapf(err, "error deleting topic '%s'", fnName)
		}
	}

	return nil
}

func newRiffClient(kubeconfPath string) riffcs.Interface {
	config, err := kubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return riffcs.NewForConfigOrDie(config)
}

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfPath)
	}
	return rest.InClusterConfig()
}
