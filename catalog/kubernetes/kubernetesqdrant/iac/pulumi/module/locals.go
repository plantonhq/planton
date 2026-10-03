package module

import (
	"fmt"
	"strconv"

	kubernetesqdrantv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesqdrant/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesqdrantv1alpha1.KubernetesQdrantSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Qdrant installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Qdrant clusters coexist in one Kubernetes cluster.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Name of the main ClusterIP Service (http 6333, grpc 6334) — the
	// module pins fullnameOverride to metadata.name (the 63-char
	// child-name discipline), so qdrant.fullname and every chart child
	// derive from the resource name deterministically.
	ServiceName string

	// Name of the chart-owned Secret carrying the API key material
	// (`<name>-apikey`, keys api-key / read-only-api-key). Populated for
	// the generate arm and the existing-secret arm alike — the chart
	// copies referenced keys into it at install. Empty when no key is
	// declared (unauthenticated).
	ApiKeySecretName string

	// Same chart-owned Secret when a read-only key is declared, else "".
	ReadOnlyApiKeySecretName string

	// In-cluster endpoints. Scheme follows the tls block.
	HttpEndpoint string
	GrpcEndpoint string

	// kubectl one-liner for reaching REST from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesqdrantv1alpha1.KubernetesQdrantIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesQdrant.String(),
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

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	// fullnameOverride is pinned to metadata.name (values.go), so the
	// chart's main Service is exactly the resource name.
	serviceName := releaseName

	apiKeySecretName := ""
	if spec.GetApiKey() != nil {
		apiKeySecretName = releaseName + "-apikey"
	}
	readOnlyApiKeySecretName := ""
	if spec.GetReadOnlyApiKey() != nil {
		readOnlyApiKeySecretName = releaseName + "-apikey"
	}

	httpScheme := "http"
	if spec.GetTls() != nil {
		httpScheme = "https"
	}

	return &Locals{
		Spec:                     spec,
		Labels:                   labels,
		Namespace:                namespace,
		ReleaseName:              releaseName,
		ChartVersion:             chartVersion,
		ServiceName:              serviceName,
		ApiKeySecretName:         apiKeySecretName,
		ReadOnlyApiKeySecretName: readOnlyApiKeySecretName,
		HttpEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local:6333",
			httpScheme, serviceName, namespace),
		GrpcEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:6334",
			serviceName, namespace),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 6333:6333",
			serviceName, namespace),
	}
}
