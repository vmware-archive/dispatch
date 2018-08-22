package istio

import (
	"context"
	"strings"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/trace"
)

const (
	// InternalGateway represents the gateway that we should redirect to
	// Should probably come from a config file
	InternalGateway = "knative-ingressgateway.istio-system.svc.cluster.local"
)

func NewClient() (*Client, error) {
	log.Infoln("Trying to create istio client")
	crdClient, err := crd.NewClient("", "", []model.ProtoSchema{model.VirtualService}, "")
	if err != nil {
		return nil, err
	}
	return &Client{
		istioClient: crdClient,
	}, nil
}

type Client struct {
	istioClient *crd.Client
}

func (c *Client) AddAPI(ctx context.Context, specs, name, org string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	vs := model.VirtualService
	spec, err := vs.FromYAML(specs)
	if err != nil {
		return ewrapper.Wrapf(err, "Unable to build virtual service from spec string")
	}

	cfg := model.Config{
		model.ConfigMeta{
			Type:      "virtual-service",
			Group:     "networking.istio.io",
			Version:   "v1alpha3",
			Name:      name,
			Namespace: org,
		},
		spec,
	}

	result, err := c.istioClient.Create(cfg)
	if err != nil {
		return ewrapper.Wrapf(err, "Unable to create virtual service")
	}
	log.Debugf("Created object with resourceVersion: %v", result)

	return nil
}

func (c *Client) GetAPI(ctx context.Context, name, org string) (*model.Config, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	cfg, found := c.istioClient.Get("virtual-service", name, org)
	if !found {
		return nil, ewrapper.Errorf("Istio couldn't located the api %s you requested", name)
	}
	return cfg, nil
}

func (c *Client) DeleteAPI(ctx context.Context, name, org string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	err := c.istioClient.Delete("virtual-service", name, org)
	if err != nil && !strings.HasSuffix(err.Error(), "not found") {
		return ewrapper.Wrapf(err, "Unable to delete api %s", name)
	}
	return nil
}

func (c *Client) ListAPI(ctx context.Context, org string) ([]model.Config, error) {
	return c.istioClient.List("virtual-service", org)
}
