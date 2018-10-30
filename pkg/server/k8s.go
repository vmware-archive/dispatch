///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/dispatch/pkg/utils"
)

func k8sClient(kubeconfPath string) kubernetes.Interface {
	config, err := utils.KubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "error configuring k8s API client"))
	}
	return kubernetes.NewForConfigOrDie(config)
}

// Helper function to resolve the image registry which may be described as a kubernetes service name.  Service names
// must be resolved to their cluster IP
func registryURL(client kubernetes.Interface, imageRegistry, namespace string) (string, error) {
	regURL, err := url.Parse(imageRegistry)
	if err != nil {
		log.Fatalf("cannot parse image registry %s", imageRegistry)
	}
	// Really scheme!? (since we don't have an actual scheme the host gets put there...)
	hostParts := strings.Split(regURL.Scheme, ".")
	// If the image is hosted locally, we need the cluster IP
	if len(hostParts) == 1 || hostParts[len(hostParts)-1] == "local" {
		namespace := namespace
		if len(hostParts) > 1 {
			namespace = hostParts[1]
		}
		service, err := client.CoreV1().Services(namespace).Get(hostParts[0], metav1.GetOptions{})
		if err == nil {
			regURL.Scheme = service.Spec.ClusterIP
		}
		log.Infof("Failed to get cluster ip, assume resolvable: %v", err)
		// Assume the url is resolvable externally if no matching service is found
	}
	log.Infof("Using docker registry %s", regURL.String())
	return regURL.String(), nil
}
