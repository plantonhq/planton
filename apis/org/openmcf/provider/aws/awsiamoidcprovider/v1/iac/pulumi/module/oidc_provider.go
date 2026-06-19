package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func oidcProvider(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	name := locals.AwsIamOidcProvider.Metadata.Name
	spec := locals.AwsIamOidcProvider.Spec

	args := &iam.OpenIdConnectProviderArgs{
		// spec.Url is a StringValueOrRef; GetValue() returns the resolved literal
		// (either the inline value or the value pulled from the referenced resource).
		Url:           pulumi.String(spec.Url.GetValue()),
		ClientIdLists: pulumi.ToStringArray(spec.ClientIdList),
		Tags:          pulumi.ToStringMap(locals.AwsTags),
	}

	// Only set thumbprints when explicitly provided. For issuers backed by a
	// well-known CA, omitting them lets AWS derive the thumbprint from its trusted
	// CA store -- passing an empty list would force a needless (and possibly wrong)
	// value, so we leave the field unset instead. This mirrors the Terraform
	// module, which leaves thumbprint_list unset for the same reason.
	if len(spec.ThumbprintList) > 0 {
		args.ThumbprintLists = pulumi.ToStringArray(spec.ThumbprintList)
	}

	createdProvider, err := iam.NewOpenIdConnectProvider(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IAM OIDC provider")
	}

	// Export final outputs. createdProvider.Url is the issuer URL AWS stores with
	// the scheme stripped, matching the provider-url segment of the ARN.
	ctx.Export(OpProviderArn, createdProvider.Arn)
	ctx.Export(OpProviderUrl, createdProvider.Url)

	return nil
}
