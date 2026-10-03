package module

import (
	"fmt"
	"strconv"

	kubernetesaltinityoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesaltinityoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesaltinityoperatorv1alpha1.KubernetesAltinityOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// DeploymentName is the operator Deployment name. The chart names it
	// with its fullname helper and the module pins fullnameOverride to
	// the resource name (see buildHelmValues), so the Deployment name IS
	// the release name (templates/generated/Deployment-clickhouse-operator.yaml).
	DeploymentName string

	// CredentialsSecretName is the chart-managed Secret holding the
	// operator's ClickHouse credentials (keys username/password). Also
	// named with the fullname helper — the release name
	// (templates/generated/Secret-clickhouse-operator.yaml).
	CredentialsSecretName string

	// MetricsEndpoint is the in-cluster URL of the metrics-exporter's
	// per-cluster Prometheus metrics: the chart's "<fullname>-metrics"
	// Service exposes port 8888 only while metrics are enabled
	// (templates/generated/Service-clickhouse-operator-metrics.yaml).
	// Empty when the exporter is disabled.
	MetricsEndpoint string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesaltinityoperatorv1alpha1.KubernetesAltinityOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesAltinityOperator.String(),
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

	releaseName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	// metrics.enabled is true-defaulted upstream AND in the spec — the
	// endpoint clears only on an explicit false.
	metricsEnabled := true
	if m := spec.GetMetrics(); m != nil && m.Enabled != nil {
		metricsEnabled = m.GetEnabled()
	}
	metricsEndpoint := ""
	if metricsEnabled {
		metricsEndpoint = fmt.Sprintf("http://%s-metrics.%s.svc.cluster.local:%d/metrics",
			releaseName, namespace, vars.MetricsPort)
	}

	return &Locals{
		Spec:                  spec,
		Labels:                labels,
		Namespace:             namespace,
		ReleaseName:           releaseName,
		ChartVersion:          chartVersion,
		DeploymentName:        releaseName,
		CredentialsSecretName: releaseName,
		MetricsEndpoint:       metricsEndpoint,
	}
}
