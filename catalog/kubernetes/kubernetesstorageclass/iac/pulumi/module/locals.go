package module

import (
	"strconv"

	kubernetesstorageclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstorageclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The upstream mechanism for marking a cluster's default StorageClass — the
// spec's is_default_class bool renders to this annotation on both engines.
const isDefaultClassAnnotation = "storageclass.kubernetes.io/is-default-class"

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetesstorageclassv1alpha1.KubernetesStorageClassSpec
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The Kubernetes API strings for the enum-typed policies, resolved with
	// the API server's own defaults so both engines submit identical objects
	// whether or not the spec set the optional fields.
	ReclaimPolicy     string
	VolumeBindingMode string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesstorageclassv1alpha1.KubernetesStorageClassIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesStorageClass.String(),
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
	// The default-class marker is a first-class spec field; the annotation is
	// only the wire form. Written only when true (identical to the Terraform
	// module): demotion works by the annotation leaving the desired state,
	// and setting it after the user map keeps it from being overridden.
	if spec.GetIsDefaultClass() {
		annotations[isDefaultClassAnnotation] = "true"
	}

	return &Locals{
		Context:           ctx,
		Spec:              spec,
		Name:              spec.GetName(),
		Labels:            labels,
		Annotations:       annotations,
		ReclaimPolicy:     resolveReclaimPolicy(spec),
		VolumeBindingMode: resolveVolumeBindingMode(spec),
	}
}

// resolveReclaimPolicy returns the API string, applying the Kubernetes
// default (Delete) when the spec omits the field.
func resolveReclaimPolicy(spec *kubernetesstorageclassv1alpha1.KubernetesStorageClassSpec) string {
	if spec.GetReclaimPolicy() == kubernetesstorageclassv1alpha1.KubernetesStorageClassSpec_retain {
		return "Retain"
	}
	return "Delete"
}

// resolveVolumeBindingMode returns the API string, applying the Kubernetes
// default (Immediate) when the spec omits the field.
func resolveVolumeBindingMode(spec *kubernetesstorageclassv1alpha1.KubernetesStorageClassSpec) string {
	if spec.GetVolumeBindingMode() == kubernetesstorageclassv1alpha1.KubernetesStorageClassSpec_wait_for_first_consumer {
		return "WaitForFirstConsumer"
	}
	return "Immediate"
}
