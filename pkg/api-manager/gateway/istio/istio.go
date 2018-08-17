package istio

import (
	"context"
	"fmt"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
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

func (c *Client) buildVirtualService(gateway, service, destination string, cors bool, methods ...string) (string, error) {
	match := httpMatch{
		URI: uri{
			Prefix: fmt.Sprintf("/%v", destination),
		},
	}
	destinationRoute := route{
		Destination: routeDest{
			Host: InternalGateway,
		},
		Weight: 100,
	}

	matcher := httpMatcher{
		Match: []httpMatch{match},
		Rewrite: rewrite{
			Authority: fmt.Sprintf("%v.default.example.com", service),
		},
		Route: []route{destinationRoute},
	}
	if cors {
		matcher.CORS = corsPolicy{
			AllowHeaders: []string{"*"},
			AllowMethods: methods,
		}
	}
	newRoute := reRouterSpec{
		Gateways: []string{gateway},
		Hosts:    []string{"*"},
		HTTP:     []httpMatcher{matcher},
	}
	bytes, err := yaml.Marshal(newRoute)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c *Client) AddAPI(ctx context.Context, api *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	vsSpecs, err := c.buildVirtualService(
		"knative-shared-gateway.knative-serving.svc.cluster.local",
		fmt.Sprintf("%v.default.example.com", api.Function),
		fmt.Sprintf("/%v", api.Name),
		api.CORS,
		api.Methods...,
	)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to build the specs for virtual service")
	}

	vs := model.VirtualService
	spec, err := vs.FromYAML(vsSpecs)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to build virtual service from spec string")
	}

	cfg := model.Config{
		model.ConfigMeta{
			Type:    "virtual-service",
			Group:   "networking.istio.io",
			Version: "v1alpha3",
			Name:    api.Name,
		},
		spec,
	}

	result, err := c.istioClient.Create(cfg)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to create virtual service")
	}
	log.Debugf("Created object with resourceVersion: %v", result)

	return api, nil
}

func (c *Client) GetAPI(ctx context.Context, name string) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) UpdateAPI(ctx context.Context, name string, api *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) DeleteAPI(ctx context.Context, api *gateway.API) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return ewrapper.Errorf("Not implemented yet dummy")
}
