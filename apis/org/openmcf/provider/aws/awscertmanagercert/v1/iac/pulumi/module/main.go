package module

import (
	"github.com/pkg/errors"
	awscertv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscertmanagercert/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the aws_cert_manager_cert Pulumi module.
// It prepares context, configures the AWS provider, and calls certManagerCert().
func Resources(ctx *pulumi.Context, stackInput *awscertv1.AwsCertManagerCertStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsCertManagerCert.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsCertManagerCert.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Call the core logic for ACM certificate + DNS validation setup.
	if err := certManagerCert(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws cert manager cert resource")
	}

	return nil
}
