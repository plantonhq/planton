package module

import (
	"fmt"
	"strconv"

	kubernetestempov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestempo/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetestempov1alpha1.KubernetesTempoSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources).
	Labels map[string]string

	// Namespace Tempo installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name. fullnameOverride pins every chart
	// child name to it.
	ReleaseName string

	// Chart version resolved to the pinned default when unset.
	ChartVersion string

	// Name of the Tempo Service (= the release name via fullnameOverride).
	ServiceName string

	// In-cluster HTTP endpoint (port 3200) — the Grafana `tempo`
	// datasource / TraceQL URL.
	HttpEndpoint string

	// In-cluster OTLP ingest endpoints (gRPC 4317, HTTP 4318).
	OtlpGrpcEndpoint string
	OtlpHttpEndpoint string

	// kubectl one-liner for reaching the Tempo API from a workstation.
	PortForwardCommand string

	// The r2 arm's module-owned credentials Secret (`<name>-r2-credentials`)
	// and its rendering; nil unless storage.r2 is declared.
	R2CredentialsSecretName string
	R2                      *r2Storage
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetestempov1alpha1.KubernetesTempoIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesTempo.String(),
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

	return &Locals{
		Spec:         spec,
		Labels:       labels,
		Namespace:    namespace,
		ReleaseName:  releaseName,
		ChartVersion: chartVersion,
		ServiceName:  releaseName,

		R2CredentialsSecretName: releaseName + vars.R2CredentialsSecretSuffix,
		R2:                      buildR2Storage(spec.GetStorage()),

		HttpEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			releaseName, namespace, vars.HttpPort),
		OtlpGrpcEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			releaseName, namespace, vars.OtlpGrpc),
		OtlpHttpEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			releaseName, namespace, vars.OtlpHttp),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			releaseName, namespace, vars.HttpPort, vars.HttpPort),
	}
}
