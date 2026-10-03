package module

import (
	"strconv"

	kubernetespriorityclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespriorityclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetespriorityclassv1alpha1.KubernetesPriorityClassSpec
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The API string for preemption_policy, resolved with the server default
	// (PreemptLowerPriority) so both engines submit identical objects.
	PreemptionPolicy string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetespriorityclassv1alpha1.KubernetesPriorityClassIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesPriorityClass.String(),
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

	preemptionPolicy := "PreemptLowerPriority"
	if spec.GetPreemptionPolicy() == kubernetespriorityclassv1alpha1.KubernetesPriorityClassSpec_never {
		preemptionPolicy = "Never"
	}

	return &Locals{
		Context:          ctx,
		Spec:             spec,
		Name:             spec.GetName(),
		Labels:           labels,
		Annotations:      annotations,
		PreemptionPolicy: preemptionPolicy,
	}
}
