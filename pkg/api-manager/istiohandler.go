package apimanager

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	sharedclientset "github.com/knative/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	// InternalGateway represents the gateway that we should redirect to
	// Should probably come from a config file
	InternalGateway = "knative-ingressgateway.istio-system.svc.cluster.local"
)

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath == "" {
		userKubeConfig := filepath.Join(os.Getenv("HOME"), ".kube/config")
		if _, err := os.Stat(userKubeConfig); err == nil {
			kubeconfPath = userKubeConfig
		}
	}
	if kubeconfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfPath)
	}
	return rest.InClusterConfig()
}

func knClient(kubeconfPath string) sharedclientset.Interface {
	config, err := kubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return sharedclientset.NewForConfigOrDie(config)
}

func makeMethodRegex(methods []string) string {
	return strings.Join(methods, "|") + "/g"
}

func makeMethodsFromRegex(regex string) []string {
	split := strings.Split(regex, "|")
	if len(split) == 1 {
		return []string{strings.Split(split[0], "/g")[0]}
	}
	return split[0 : len(split)-1]
}

type IstioHandlers struct {
	knClient sharedclientset.Interface
}

func NewIstioHandlers() *IstioHandlers {
	return &IstioHandlers{
		// TODO: Pass kubeconfigPath through config (see Ivan's function manager work)
		knClient: knClient(""),
	}
}

func apiModelOntoIstioEntity(m *v1.API) v1alpha3.VirtualServiceSpec {
	var result v1alpha3.VirtualServiceSpec
	hosts := m.Hosts
	if len(m.Hosts) == 0 {
		hosts = append(hosts, "*")
	}
	result.Hosts = hosts
	result.Gateways = []string{"knative-shared-gateway.knative-serving.svc.cluster.local"}

	var routes []v1alpha3.HTTPRoute
	for _, prefix := range m.Uris {
		if prefix == "/" {
			prefix = fmt.Sprintf("/%v", *m.Function)
		}
		match := v1alpha3.HTTPMatchRequest{
			Uri:    &v1alpha1.StringMatch{Prefix: prefix},
			Method: &v1alpha1.StringMatch{Regex: makeMethodRegex(m.Methods)},
		}
		route := v1alpha3.HTTPRoute{
			Match: []v1alpha3.HTTPMatchRequest{match},
			Route: []v1alpha3.DestinationWeight{
				v1alpha3.DestinationWeight{
					Destination: v1alpha3.Destination{
						Host: InternalGateway,
					},
					Weight: 100,
				},
			},
			Rewrite: &v1alpha3.HTTPRewrite{
				Authority: fmt.Sprintf("%v.default.example.com", *m.Function),
			},
		}
		routes = append(routes, route)
	}
	log.Infof("Routes: %+v", routes)
	result.Http = routes
	return result
}

func istioEntityOntoAPIModel(vs v1alpha3.VirtualServiceSpec) *v1.API {
	var result v1.API
	var uris []string
	var methods []string
	var function string
	hosts := vs.Hosts
	for _, route := range vs.Http {
		uris = append(uris, route.Match[0].Uri.Prefix)
		methods = makeMethodsFromRegex(route.Match[0].Method.Regex)
		function = strings.Split(route.Rewrite.Authority, ".")[0]
	}
	result.Enabled = true
	result.Methods = methods
	result.Uris = uris
	result.Hosts = hosts
	result.Function = &function
	result.Protocols = []string{"Http"}
	return &result
}

func (cl *IstioHandlers) AddAPI(params endpoint.AddAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to add istio api: %+v", params)

	ns := params.XDispatchOrg

	specs := apiModelOntoIstioEntity(params.Body)

	virtualService := v1alpha3.VirtualService{
		metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		metav1.ObjectMeta{
			Name:      *params.Body.Name,
			Namespace: ns,
			Labels: map[string]string{
				"Organization": ns,
			},
		},
		specs,
	}

	log.Infof("Trying to create VirtualService: %+v", virtualService)
	newVirtualService, err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).Create(&virtualService)
	if err != nil {
		log.Errorf("Couldn't Create API: %v", err)
		return endpoint.NewAddAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to create api"),
			})
	}
	log.Infof("Created Virtual Service: %+v", newVirtualService)
	return endpoint.NewAddAPIOK().WithPayload(params.Body)
}

func (cl *IstioHandlers) GetAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	ns := params.XDispatchOrg
	log.Infof("Trying to get an istio api: %+v", params)
	result, err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).Get(params.API, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
	})
	if err != nil {
		log.Infof("Couldn't find api: %v", err)
		return endpoint.NewGetAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to find"),
			})
	}
	converted := istioEntityOntoAPIModel(result.Spec)
	converted.Name = &result.Name
	return endpoint.NewGetAPIOK().WithPayload(converted)
}

func (cl *IstioHandlers) GetAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	ns := params.XDispatchOrg
	log.Infof("Getting apis: %+v", params)
	allVs, err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		LabelSelector: "Organization",
	})
	if err != nil {
		log.Infof("Couldn't find api: %v", err)
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to find"),
			})
	}
	var results []*v1.API
	for _, vs := range allVs.Items {
		converted := istioEntityOntoAPIModel(vs.Spec)
		converted.Name = &vs.Name
		results = append(results, converted)
	}
	return endpoint.NewGetApisOK().WithPayload(results)
}

func (cl *IstioHandlers) UpdateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to add istio api: %+v", params)

	ns := params.XDispatchOrg

	specs := apiModelOntoIstioEntity(params.Body)

	virtualService := v1alpha3.VirtualService{
		metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		metav1.ObjectMeta{
			Name:      *params.Body.Name,
			Namespace: ns,
		},
		specs,
	}

	log.Infof("Trying to create VirtualService: %+v", virtualService)
	newVirtualService, err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).Update(&virtualService)
	if err != nil {
		log.Errorf("Couldn't Create API: %v", err)
		return endpoint.NewUpdateAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to update api"),
			})
	}
	log.Infof("Created Virtual Service: %+v", newVirtualService)
	return endpoint.NewUpdateAPIOK().WithPayload(params.Body)
}

func (cl *IstioHandlers) DeleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	ns := params.XDispatchOrg
	log.Infof("Trying to get an istio api: %+v", params)
	err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).Delete(params.API, &metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1alpha3",
		},
	})
	if err != nil {
		log.Infof("Couldn't delete api: %v", err)
		return endpoint.NewDeleteAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to delete api"),
			})
	}
	return endpoint.NewDeleteAPIOK().WithPayload(&v1.API{
		Name: &params.API,
	})
}
