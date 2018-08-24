package apimanager

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

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

func (cl *IstioHandlers) AddAPI(params endpoint.AddAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to add istio api: %+v", params)

	// api := apiModelOntoIstioEntity(params.Body)
	// virtualServiceSpec, err := yaml.Marshal(api)
	// if err != nil {
	// 	log.Errorf("Failed to marshal virtualservicspec: %+v", api)
	// }
	// err = cl.client.AddAPI(ctx, string(virtualServiceSpec), api.Name, params.XDispatchOrg)
	// if err != nil {
	// 	log.Errorf("Istio failed to create the api: %v", err)
	// }
	// m := istioEntityToModel(api)
	// log.Infof("Added api: %+v", m)
	// return endpoint.NewAddAPIOK().WithPayload(m)

	ns := params.XDispatchOrg

	sampleVs := v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *params.Body.Name,
			Namespace: ns,
		},
	}
	virtualService, err := cl.knClient.NetworkingV1alpha3().VirtualServices(ns).Create(&sampleVs)
	if err != nil {
		log.Errorf("Couldn't Create API: %v", err)
		return endpoint.NewAddAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("Failed to create api"),
			})
	}
	log.Infof("Created Virtual Service: %+v", virtualService)
	return endpoint.NewAddAPIOK().WithPayload(params.Body)
}

func (cl *IstioHandlers) GetAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to get an istio api: %+v", params)
	return nil
}

func (cl *IstioHandlers) GetAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Getting apis: %+v", params)

	return nil
}

func (cl *IstioHandlers) UpdateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	return nil
}

func (cl *IstioHandlers) DeleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to get an istio api: %+v", params)
	return nil
}
