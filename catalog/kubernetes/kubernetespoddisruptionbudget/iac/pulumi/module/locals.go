package module

import (
	"strconv"

	kubernetespoddisruptionbudgetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespoddisruptionbudget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetespoddisruptionbudgetv1alpha1.KubernetesPodDisruptionBudgetSpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The API string for unhealthy_pod_eviction_policy, resolved with the
	// server default (IfHealthyBudget) so both engines submit identical
	// objects.
	UnhealthyPodEvictionPolicy string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetespoddisruptionbudgetv1alpha1.KubernetesPodDisruptionBudgetIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesPodDisruptionBudget.String(),
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
	for k, v := range spec.GetLabels() {
		if _, isIdentityKey := labels[k]; !isIdentityKey {
			labels[k] = v
		}
	}

	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When omitted entirely, fall back to the cluster's
	// "default" namespace — the same behavior as kubectl without a namespace flag.
	namespace := spec.GetNamespace().GetValue()
	if namespace == "" {
		namespace = "default"
	}

	unhealthyPolicy := "IfHealthyBudget"
	if spec.GetUnhealthyPodEvictionPolicy() == kubernetespoddisruptionbudgetv1alpha1.KubernetesPodDisruptionBudgetSpec_always_allow {
		unhealthyPolicy = "AlwaysAllow"
	}

	return &Locals{
		Context:                    ctx,
		Spec:                       spec,
		Namespace:                  namespace,
		Name:                       spec.GetName(),
		Labels:                     labels,
		Annotations:                annotations,
		UnhealthyPodEvictionPolicy: unhealthyPolicy,
	}
}
