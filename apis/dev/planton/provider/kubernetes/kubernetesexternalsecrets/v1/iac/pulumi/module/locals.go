package module

import (
	"strconv"

	kubernetesexternalsecretsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesexternalsecrets/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all computed values for the KubernetesExternalSecrets module
type Locals struct {
	// KubernetesExternalSecrets is the target resource
	KubernetesExternalSecrets *kubernetesexternalsecretsv1.KubernetesExternalSecrets

	// Namespace where resources will be deployed
	Namespace string

	// Labels to apply to all resources
	Labels map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	ServiceAccountName string
	HelmReleaseName    string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesexternalsecretsv1.KubernetesExternalSecretsStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	locals := &Locals{
		KubernetesExternalSecrets: target,
	}

	// Labels
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesExternalSecrets.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Namespace with default
	locals.Namespace = spec.Namespace.GetValue()
	if locals.Namespace == "" {
		locals.Namespace = vars.Namespace
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "eso-my-instance")
	locals.ServiceAccountName = target.Metadata.Name
	locals.HelmReleaseName = target.Metadata.Name

	// Export namespace
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return locals
}
