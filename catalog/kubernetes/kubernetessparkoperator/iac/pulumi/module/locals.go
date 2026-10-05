package module

import (
	"strconv"

	kubernetessparkoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessparkoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetessparkoperatorv1alpha1.KubernetesSparkOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. The module pins the chart's
	// fullnameOverride AND every RBAC name to it, so every chart-derived
	// name hangs off this value.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// WorkloadNamespaces is the explicit list of namespaces Spark
	// workloads run in. The watch scope and the workload RBAC are ONE
	// chart surface (workloadResources), decided together from this
	// list. Empty = cluster-wide.
	WorkloadNamespaces []string

	// WorkloadFenced is the fenced posture flag: the chart CREATES each
	// listed namespace, plants the service account and a namespace-scoped
	// Role/RoleBinding in each, drops the workload ClusterRole, and
	// overrideWatchedNamespaces (chart default true) wires the operator's
	// spark.kubernetes.operator.watched.namespaces property from the same
	// list — one value, one truth.
	WorkloadFenced bool

	// WorkloadServiceAccount is the service account name Spark
	// driver/executor pods run as in every workload namespace —
	// deliberately the upstream contract ("spark" unless overridden):
	// SparkApplications reference it by that conventional name.
	WorkloadServiceAccount string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetessparkoperatorv1alpha1.KubernetesSparkOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSparkOperator.String(),
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

	workloadNamespaces := spec.GetWorkload().GetNamespaces()

	workloadServiceAccount := spec.GetWorkload().GetServiceAccount()
	if workloadServiceAccount == "" {
		workloadServiceAccount = "spark"
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              spec.Namespace.GetValue(),
		ReleaseName:            target.Metadata.Name,
		ChartVersion:           chartVersion,
		WorkloadNamespaces:     workloadNamespaces,
		WorkloadFenced:         len(workloadNamespaces) > 0,
		WorkloadServiceAccount: workloadServiceAccount,
	}
}
