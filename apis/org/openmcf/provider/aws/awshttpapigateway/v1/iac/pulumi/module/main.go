package module

import (
	"github.com/pkg/errors"
	awshttpapigatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awshttpapigateway/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates HTTP API Gateway creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awshttpapigatewayv1.AwsHttpApiGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// -----------------------------------------------------------------------
	// AWS provider
	// -----------------------------------------------------------------------

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

	// -----------------------------------------------------------------------
	// 1. Create the HTTP API
	// -----------------------------------------------------------------------

	createdApi, err := httpApi(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "http api")
	}

	// -----------------------------------------------------------------------
	// 2. Create the stage
	// -----------------------------------------------------------------------

	if err := stage(ctx, locals, createdApi, provider); err != nil {
		return errors.Wrap(err, "api stage")
	}

	// -----------------------------------------------------------------------
	// 3. Create integrations (deduplicated)
	// -----------------------------------------------------------------------

	integrationMap, err := integrations(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api integrations")
	}

	// -----------------------------------------------------------------------
	// 4. Create authorizers (if any)
	// -----------------------------------------------------------------------

	authorizerMap, err := authorizers(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api authorizers")
	}

	// -----------------------------------------------------------------------
	// 5. Create routes
	// -----------------------------------------------------------------------

	if err := routes(ctx, locals, createdApi, integrationMap, authorizerMap, provider); err != nil {
		return errors.Wrap(err, "api routes")
	}

	return nil
}
