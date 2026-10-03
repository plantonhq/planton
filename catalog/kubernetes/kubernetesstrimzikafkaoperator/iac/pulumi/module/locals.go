package module

import (
	"strconv"

	kubernetesstrimzikafkaoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstrimzikafkaoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesstrimzikafkaoperatorv1alpha1.KubernetesStrimziKafkaOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref). With the default watch scope this is also
	// the only namespace the operator reconciles Kafka clusters in.
	Namespace string

	// Helm release name — metadata.name. The chart derives its resource
	// names from fixed Strimzi names (strimzi-cluster-operator), so a
	// SECOND install in one cluster additionally needs
	// create_global_resources false — see the spec comment.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesstrimzikafkaoperatorv1alpha1.KubernetesStrimziKafkaOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesStrimziKafkaOperator.String(),
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

	return &Locals{
		Spec:         spec,
		Labels:       labels,
		Namespace:    spec.Namespace.GetValue(),
		ReleaseName:  target.Metadata.Name,
		ChartVersion: chartVersion,
	}
}
