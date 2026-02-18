package module

import (
	"github.com/pkg/errors"
	awscognitouserpoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscognitouserpool/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of a Cognito User Pool with app clients and
// an optional domain, then exports outputs for downstream references.
func Resources(ctx *pulumi.Context, stackInput *awscognitouserpoolv1.AwsCognitoUserPoolStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.Target.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.Target.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// User pool (always created)
	createdPool, err := userPool(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cognito user pool")
	}

	// App clients (always created — at least one required by spec)
	if err := clients(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool clients")
	}

	// Domain (optional)
	if err := domain(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool domain")
	}

	return nil
}
