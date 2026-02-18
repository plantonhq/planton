package module

import (
	"github.com/pkg/errors"
	awswafwebaclv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awswafwebacl/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsWafWebAcl Pulumi module.
// It creates the Web ACL with rules and optional logging configuration.
func Resources(ctx *pulumi.Context, stackInput *awswafwebaclv1.AwsWafWebAclStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.WebAcl.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.WebAcl.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	createdWebAcl, err := webAcl(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create WAF Web ACL")
	}

	if locals.WebAcl.Spec.Logging != nil {
		if err := logging(ctx, locals, provider, createdWebAcl); err != nil {
			return errors.Wrap(err, "failed to configure WAF logging")
		}
	}

	return nil
}
