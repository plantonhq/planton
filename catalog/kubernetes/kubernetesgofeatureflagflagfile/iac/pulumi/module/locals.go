package module

import (
	"strconv"

	kubernetesgofeatureflagflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflagflagfile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// DefaultKey is the spec default data key, used when the platform's
// defaulting middleware did not run.
const DefaultKey = "flags.goff.yaml"

// Locals holds computed values derived from the IaC input. Every resolution
// has an exact twin in the Terraform module's locals.tf.
type Locals struct {
	Spec      *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileSpec
	Name      string
	Namespace string
	Key       string
	Labels    map[string]string
}

func initializeLocals(iacInput *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileIacInput) *Locals {
	target := iacInput.Target
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesGoFeatureFlagFlagFile.String(),
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
	key := target.Spec.GetKey()
	if key == "" {
		key = DefaultKey
	}
	return &Locals{
		Spec:      target.Spec,
		Name:      target.Metadata.Name,
		Namespace: target.Spec.Namespace.GetValue(),
		Key:       key,
		Labels:    labels,
	}
}
