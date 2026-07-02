package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// instanceProfile provisions the profile that delivers an IAM role to EC2:
// instances cannot assume a role directly, they can only be launched with a
// profile that carries one. The profile holds exactly one role (an AWS limit,
// not a provider choice). Name and path are create-only; the role can be
// swapped in place -- AWS removes the old role and adds the new one without
// replacing the profile, so running instances pick up the new role's
// credentials on their next metadata refresh.
func instanceProfile(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	profileName := locals.AwsIamInstanceProfile.Metadata.Name
	spec := locals.AwsIamInstanceProfile.Spec

	profileArgs := &iam.InstanceProfileArgs{
		Name: pulumi.String(profileName),
		// The role is attached by NAME (not ARN) -- that is what the underlying
		// AddRoleToInstanceProfile API takes. A valueFrom reference is resolved to
		// the referenced AwsIamRole's role_name before the module runs. IAM is
		// eventually consistent; the provider retries the attach internally until
		// a freshly-created role is visible.
		Role: pulumi.String(spec.Role.GetValue()),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", profileName)),
	}

	if spec.Path != "" {
		profileArgs.Path = pulumi.StringPtr(spec.Path)
	}

	createdProfile, err := iam.NewInstanceProfile(ctx, profileName, profileArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create iam instance profile")
	}

	ctx.Export(OpInstanceProfileArn, createdProfile.Arn)
	ctx.Export(OpInstanceProfileName, createdProfile.Name)
	ctx.Export(OpInstanceProfileId, createdProfile.UniqueId)
	ctx.Export(OpRoleName, createdProfile.Role)

	return nil
}
