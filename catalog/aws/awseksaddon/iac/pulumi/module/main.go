package module

import (
	"github.com/pkg/errors"
	awseksaddonv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksaddon/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the managed add-on. The cluster attaches by
// reference, and the add-on's IAM identity -- when it needs one beyond
// the node role -- is a referenced AwsIamRole wired through IRSA or EKS
// Pod Identity (this module never modifies a role it merely references).
//
// AWS keys the add-on on (cluster, addon_name): the Pulumi resource name
// uses the manifest's metadata.name for a stable URN, while the add-on's
// real identity comes from the spec's addon_name.
func Resources(ctx *pulumi.Context, iacInput *awseksaddonv1alpha1.AwsEksAddonIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEksAddon.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsEksAddon.Spec

	args := &eks.AddonArgs{
		ClusterName: pulumi.String(spec.ClusterName.GetValue()),
		AddonName:   pulumi.String(spec.AddonName),
		Tags:        pulumi.ToStringMap(locals.AwsTags),
	}

	// Empty means the AWS default version for the cluster's Kubernetes
	// version -- the never-goes-stale choice. AWS reports the resolved
	// version back through the addon_version output either way.
	if spec.AddonVersion != "" {
		args.AddonVersion = pulumi.StringPtr(spec.AddonVersion)
	}

	// AWS's conflict handling is asymmetric (create: NONE/OVERWRITE;
	// update: +PRESERVE) -- the spec's CEL rules enforce the split before
	// anything reaches the API. Unset falls back to AWS's NONE, which
	// fails loudly on conflicts instead of adopting silently.
	if spec.ResolveConflictsOnCreate != "" {
		args.ResolveConflictsOnCreate = pulumi.StringPtr(spec.ResolveConflictsOnCreate)
	}
	if spec.ResolveConflictsOnUpdate != "" {
		args.ResolveConflictsOnUpdate = pulumi.StringPtr(spec.ResolveConflictsOnUpdate)
	}

	if spec.ConfigurationValues != "" {
		args.ConfigurationValues = pulumi.StringPtr(spec.ConfigurationValues)
	}

	// IRSA: the referenced role must already exist and trust the cluster's
	// OIDC provider; empty means the add-on's pods use node-role permissions.
	if spec.ServiceAccountRoleArn != nil && spec.ServiceAccountRoleArn.GetValue() != "" {
		args.ServiceAccountRoleArn = pulumi.StringPtr(spec.ServiceAccountRoleArn.GetValue())
	}

	// EKS Pod Identity: the modern no-OIDC-provider alternative. Each
	// association binds one service account to one referenced role.
	if len(spec.PodIdentityAssociations) > 0 {
		associations := make(eks.AddonPodIdentityAssociationArray, 0, len(spec.PodIdentityAssociations))
		for _, association := range spec.PodIdentityAssociations {
			associations = append(associations, &eks.AddonPodIdentityAssociationArgs{
				RoleArn:        pulumi.String(association.RoleArn.GetValue()),
				ServiceAccount: pulumi.String(association.ServiceAccount),
			})
		}
		args.PodIdentityAssociations = associations
	}

	// Deleting with preserve leaves the add-on's Kubernetes resources
	// running as self-managed software -- the no-outage way to hand an
	// add-on's lifecycle back to cluster operators.
	if spec.Preserve {
		args.Preserve = pulumi.BoolPtr(true)
	}

	// Create-only in AWS: changing the namespace requires replacing the
	// add-on, which the provider handles as delete + create.
	if spec.NamespaceConfig != nil {
		args.NamespaceConfig = &eks.AddonNamespaceConfigArgs{
			Namespace: pulumi.StringPtr(spec.NamespaceConfig.Namespace),
		}
	}

	created, err := eks.NewAddon(ctx, locals.AwsEksAddon.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EKS add-on")
	}

	ctx.Export(OpAddonArn, created.Arn)
	ctx.Export(OpAddonName, created.AddonName)
	// The version actually running -- the resolved AWS default when the
	// spec pinned nothing.
	ctx.Export(OpAddonVersion, created.AddonVersion)

	return nil
}
