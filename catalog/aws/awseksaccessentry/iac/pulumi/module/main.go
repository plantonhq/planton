package module

import (
	"strings"

	"github.com/pkg/errors"
	awseksaccessentryv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksaccessentry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the access entry and its folded policy
// associations. The cluster and principal attach by reference; the
// associations are AWS sub-resources of exactly this (cluster,
// principal) pair, so each spec entry materializes as its own
// AccessPolicyAssociation resource keyed by the policy name -- adding,
// re-scoping, or removing one association diffs in place and never
// touches the entry or its siblings.
func Resources(ctx *pulumi.Context, iacInput *awseksaccessentryv1alpha1.AwsEksAccessEntryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEksAccessEntry.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsEksAccessEntry.Spec

	args := &eks.AccessEntryArgs{
		ClusterName:  pulumi.String(spec.ClusterName.GetValue()),
		PrincipalArn: pulumi.String(spec.PrincipalArn.GetValue()),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}

	// Empty means STANDARD (the AWS default). The node types exist for
	// self-managed/hybrid node registration; the spec's CEL rules keep
	// groups/username/associations off them, mirroring AWS's runtime rules.
	if spec.Type != "" {
		args.Type = pulumi.StringPtr(spec.Type)
	}
	if len(spec.KubernetesGroups) > 0 {
		args.KubernetesGroups = pulumi.ToStringArray(spec.KubernetesGroups)
	}
	// Empty lets AWS default the username (principal ARN for users; a
	// session-templated name for roles, which preserves the session name
	// in audit logs).
	if spec.UserName != "" {
		args.UserName = pulumi.StringPtr(spec.UserName)
	}

	created, err := eks.NewAccessEntry(ctx, locals.AwsEksAccessEntry.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EKS access entry")
	}

	// One AccessPolicyAssociation resource per spec entry, named by the
	// policy so a change to one association never replaces another. The
	// entry is the parent: AWS requires it to exist first, and deleting
	// the entry cascades the associations.
	for _, association := range spec.PolicyAssociations {
		accessScope := &eks.AccessPolicyAssociationAccessScopeArgs{
			Type: pulumi.String(association.AccessScope.Type),
		}
		if len(association.AccessScope.Namespaces) > 0 {
			accessScope.Namespaces = pulumi.ToStringArray(association.AccessScope.Namespaces)
		}

		associationName := locals.AwsEksAccessEntry.Metadata.Name + "-" + policyShortName(association.PolicyArn)
		_, err := eks.NewAccessPolicyAssociation(ctx, associationName, &eks.AccessPolicyAssociationArgs{
			ClusterName:  pulumi.String(spec.ClusterName.GetValue()),
			PrincipalArn: pulumi.String(spec.PrincipalArn.GetValue()),
			PolicyArn:    pulumi.String(association.PolicyArn),
			AccessScope:  accessScope,
		}, pulumi.Provider(provider), pulumi.Parent(created))
		if err != nil {
			return errors.Wrapf(err, "failed to associate access policy %s", association.PolicyArn)
		}
	}

	ctx.Export(OpAccessEntryArn, created.AccessEntryArn)
	ctx.Export(OpPrincipalArn, created.PrincipalArn)

	return nil
}

// policyShortName keys an association resource by its policy's name (the
// last ARN segment, e.g. "AmazonEKSViewPolicy") -- stable across scope
// edits, unique per entry because AWS allows one association per policy
// per principal.
func policyShortName(policyArn string) string {
	if idx := strings.LastIndex(policyArn, "/"); idx >= 0 && idx < len(policyArn)-1 {
		return policyArn[idx+1:]
	}
	return policyArn
}
