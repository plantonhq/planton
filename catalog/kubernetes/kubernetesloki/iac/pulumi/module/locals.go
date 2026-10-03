package module

import (
	"fmt"
	"strconv"

	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteslokiv1alpha1.KubernetesLokiSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Loki installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Loki instances can coexist (per-team log stores, a
	// compliance-scoped instance, ...). fullnameOverride pins every
	// chart child name to it.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether the gateway deploys (spec default true). The exported
	// endpoints assume it.
	GatewayEnabled bool

	// Name of the gateway Service (`<name>-gateway`) — the one front
	// door for pushes and queries in every mode. Empty when the gateway
	// is disabled.
	GatewayService string

	// In-cluster endpoint of the gateway (port 80).
	GatewayEndpoint string

	// The gateway's OTLP log-ingest endpoint — an otlphttp exporter's
	// `endpoint` value; the exporter itself appends /v1/logs, which the
	// gateway routes (the chart's nginx config serves
	// `location = /otlp/v1/logs`).
	OtlpPushEndpoint string

	// Name of the Loki HTTP Service (`<name>`, port 3100) — the direct
	// internal API behind the gateway.
	LokiService string

	// The r2 arm's module-owned credentials Secret (`<name>-r2-credentials`)
	// and its rendering; nil unless storage.r2 is declared.
	R2CredentialsSecretName string
	R2                      *r2Storage

	// kubectl one-liner for reaching the gateway from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteslokiv1alpha1.KubernetesLokiIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesLoki.String(),
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

	gatewayEnabled := true
	if gw := spec.GetGateway(); gw != nil && gw.Enabled != nil {
		gatewayEnabled = gw.GetEnabled()
	}

	gatewayService := ""
	gatewayEndpoint := ""
	otlpPushEndpoint := ""
	portForward := ""
	if gatewayEnabled {
		// loki.gateway names its Service `<fullname>-gateway`; the
		// fullname is pinned to the resource name.
		gatewayService = fmt.Sprintf("%s-gateway", releaseName)
		gatewayEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local", gatewayService, namespace)
		otlpPushEndpoint = fmt.Sprintf("%s/otlp", gatewayEndpoint)
		portForward = fmt.Sprintf("kubectl port-forward svc/%s -n %s 3100:%d",
			gatewayService, namespace, vars.GatewayServicePort)
	}

	return &Locals{
		Spec:               spec,
		Labels:             labels,
		Namespace:          namespace,
		ReleaseName:        releaseName,
		ChartVersion:       chartVersion,
		GatewayEnabled:     gatewayEnabled,
		GatewayService:     gatewayService,
		GatewayEndpoint:    gatewayEndpoint,
		OtlpPushEndpoint:   otlpPushEndpoint,
		LokiService:        releaseName,
		PortForwardCommand: portForward,

		R2CredentialsSecretName: releaseName + vars.R2CredentialsSecretSuffix,
		R2:                      buildR2Storage(spec.GetStorage()),
	}
}
