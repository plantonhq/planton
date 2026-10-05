package module

import (
	"strconv"

	kuberneteshelmreleasev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteshelmrelease/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module.
type Locals struct {
	Spec *kuberneteshelmreleasev1alpha1.KubernetesHelmReleaseSpec

	// Resource-identity labels stamped on the namespace this module creates
	// (never injected into the chart's own resources — Helm owns those).
	Labels map[string]string

	// The namespace the release installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// The Helm release name: spec.release_name when set, otherwise the
	// resource's metadata.name — identical resolution in the Terraform
	// module.
	ReleaseName string

	// Helm's own defaults resolved for the optional knobs (like the A4
	// kinds resolve API-server defaults) so both engines send identical
	// values whether or not the spec set the fields: timeout 300s,
	// max_history 10.
	TimeoutSeconds int
	MaxHistory     int
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteshelmreleasev1alpha1.KubernetesHelmReleaseIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to
	// what the Terraform module stamps for the same manifest. User-supplied
	// values never override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesHelmRelease.String(),
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

	releaseName := spec.GetReleaseName()
	if releaseName == "" {
		releaseName = target.Metadata.Name
	}

	timeoutSeconds := 300
	if spec.TimeoutSeconds != nil {
		timeoutSeconds = int(spec.GetTimeoutSeconds())
	}
	maxHistory := 10
	if spec.MaxHistory != nil {
		maxHistory = int(spec.GetMaxHistory())
	}

	return &Locals{
		Spec:           spec,
		Labels:         labels,
		Namespace:      spec.Namespace.GetValue(),
		ReleaseName:    releaseName,
		TimeoutSeconds: timeoutSeconds,
		MaxHistory:     maxHistory,
	}
}
