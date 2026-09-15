package module

import (
	"github.com/pkg/errors"
	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources declares one PlantonPlatform custom resource that the Planton
// operator (KubernetesPlantonOperator) reconciles into a running
// self-hosted platform, plus the optional owning namespace.
//
// The CR is rendered UNTYPED (apiextensions.NewCustomResource with a plain
// nested map): the PlantonPlatform schema is deliberately consumed as data
// — keys render only when the manifest declared them, so the operator's
// own defaulting stays authoritative for everything unset (see
// platform_cr.go). The exact twin of the Terraform module's
// kubectl_manifest + yamlencode locals.
//
// DESTROY: platform teardown is Kubernetes garbage collection — every
// operator-created object is owner-referenced to this CR, so deletion
// completes even when the operator itself is already gone. The delete
// timeout is headroom, not an expected wait (see vars.go).
func Resources(ctx *pulumi.Context, stackInput *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// ------------------------------ materialized Secrets -------------------
	// The credentials the declaration speaks as values — the database's
	// backup and recovery stores, the vault's seal — materialized as the
	// Secrets the CR names, before the CR, so the database is born archiving
	// and the vault's seal finds its credential at its first start.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}
	materializedSecrets, err := createMaterializedSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create the platform's credential Secrets")
	}

	// ------------------------------ the platform CR -----------------------
	resourceOptions := []pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		// Headroom for the delete — see the DESTROY note above.
		pulumi.Timeouts(&pulumi.CustomTimeouts{Delete: vars.DeleteTimeout}),
	}
	crDeps := append([]pulumi.Resource{}, materializedSecrets...)
	if createdNamespace != nil {
		crDeps = append(crDeps, createdNamespace)
	}
	if len(crDeps) > 0 {
		resourceOptions = append(resourceOptions, pulumi.DependsOn(crDeps))
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.PlatformName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(vars.ApiVersion),
			Kind:       pulumi.String(vars.Kind),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				// Namespaced, named from THIS resource's metadata.name —
				// the prefix of every object the operator creates for the
				// platform.
				Name:      pulumi.String(locals.PlatformName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: map[string]interface{}{
				"spec": platformSpecBody(locals),
			},
		},
		resourceOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to declare the PlantonPlatform resource")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpPlatformName, pulumi.String(locals.PlatformName))
	ctx.Export(OpGatewayService, pulumi.String(locals.PlatformName+vars.GatewayServiceSuffix))
	ctx.Export(OpSetupCodeSecret, pulumi.String(locals.PlatformName+vars.SetupCodeSecretSuffix))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.portForwardCommand()))
	ctx.Export(OpSetupCodeCommand, pulumi.String(locals.setupCodeCommand()))

	return nil
}
