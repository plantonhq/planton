package module

import (
	"fmt"
	"strconv"

	kuberneteskeycloakv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloak/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep: same rendered CR body, same
// outputs.
type Locals struct {
	KubernetesKeycloak *kuberneteskeycloakv1alpha1.KubernetesKeycloak
	Spec               *kuberneteskeycloakv1alpha1.KubernetesKeycloakSpec

	// ResourceName is metadata.name — the CR name and the operator's
	// naming root for every derived object (StatefulSet, Services,
	// generated Secret).
	ResourceName string

	// Namespace the server deploys into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Labels tie the module-created objects back to the Planton
	// resource.
	Labels map[string]string

	// TlsSecretName is the resolved TLS Secret name; empty means the
	// HTTPS listener is off and the server runs plain HTTP (the spec's
	// TLS-or-HTTP validation guarantees http_enabled in that case).
	TlsSecretName string

	// Resolved listener ports (spec defaults 8443/8080/9000).
	HttpPort       int
	HttpsPort      int
	ManagementPort int

	// Output handles per the operator naming contract.
	StatefulSet            string
	ServiceName            string
	DiscoveryService       string
	ApiEndpoint            string
	ManagementEndpoint     string
	InitialAdminSecretName string
	PortForwardCommand     string
}

func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskeycloakv1alpha1.KubernetesKeycloakIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKeycloak.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	resourceName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	tlsSecretName := spec.GetHttp().GetTlsSecretName().GetValue()

	httpPort := vars.DefaultHttpPort
	httpsPort := vars.DefaultHttpsPort
	if http := spec.GetHttp(); http != nil {
		if http.HttpPort != nil {
			httpPort = int(http.GetHttpPort())
		}
		if http.HttpsPort != nil {
			httpsPort = int(http.GetHttpsPort())
		}
	}
	managementPort := vars.DefaultManagementPort
	if spec.HttpManagementPort != nil {
		managementPort = int(spec.GetHttpManagementPort())
	}

	// TLS configured means the API speaks https on the https port;
	// otherwise the TLS-or-HTTP spec rule guarantees the plain-HTTP
	// listener is on.
	scheme := "https"
	apiPort := httpsPort
	if tlsSecretName == "" {
		scheme = "http"
		apiPort = httpPort
	}

	serviceName := resourceName + vars.ServiceSuffix

	// The bootstrap-admin credential handle: the user-provided Secret
	// when declared, else the operator-generated create-once
	// `<name>-initial-admin` (username "temp-admin") — seeded at FIRST
	// start only, never rotated by the operator.
	initialAdminSecretName := spec.GetBootstrapAdminSecretName()
	if initialAdminSecretName == "" {
		initialAdminSecretName = resourceName + vars.InitialAdminSecretSuffix
	}

	return &Locals{
		KubernetesKeycloak: target,
		Spec:               spec,
		ResourceName:       resourceName,
		Namespace:          namespace,
		Labels:             labels,
		TlsSecretName:      tlsSecretName,
		HttpPort:           httpPort,
		HttpsPort:          httpsPort,
		ManagementPort:     managementPort,
		StatefulSet:        resourceName,
		ServiceName:        serviceName,
		DiscoveryService:   resourceName + vars.DiscoveryServiceSuffix,
		ApiEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
			scheme, serviceName, namespace, apiPort),
		ManagementEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
			scheme, serviceName, namespace, managementPort),
		InitialAdminSecretName: initialAdminSecretName,
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d",
			namespace, serviceName, apiPort, apiPort),
	}
}
