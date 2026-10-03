package module

import (
	"strconv"

	kubernetesmanifestv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmanifest/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module.
type Locals struct {
	Spec *kubernetesmanifestv1alpha1.KubernetesManifestSpec

	// Resource-identity labels stamped on the namespace this module creates.
	// NEVER injected into the manifest's own documents — the manifest is
	// applied exactly as the user wrote it.
	Labels map[string]string

	// The anchor namespace: where namespaced documents without an explicit
	// metadata.namespace land (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// The applied-resource inventory ("apiVersion/Kind/name" per document,
	// manifest order), parsed from the input YAML so both engines export an
	// identical list regardless of how each engine tracks child resources.
	AppliedResources []string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesmanifestv1alpha1.KubernetesManifestIacInput) (*Locals, error) {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to
	// what the Terraform module stamps for the same manifest.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesManifest.String(),
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

	appliedResources, err := parseAppliedResources(spec.GetManifestYaml())
	if err != nil {
		return nil, err
	}

	return &Locals{
		Spec:             spec,
		Labels:           labels,
		Namespace:        spec.Namespace.GetValue(),
		AppliedResources: appliedResources,
	}, nil
}
