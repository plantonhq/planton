package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func stage(ctx *pulumi.Context, locals *Locals, createdApi *apigatewayv2.Api, provider *aws.Provider) error {
	spec := locals.Spec

	// Determine stage name and auto-deploy defaults.
	stageName := "$default"
	autoDeploy := true
	if spec.Stage != nil && spec.Stage.Name != "" {
		stageName = spec.Stage.Name
	}
	if spec.Stage != nil {
		autoDeploy = spec.Stage.AutoDeploy
	}

	resourceName := locals.ApiName + "-stage"

	args := &apigatewayv2.StageArgs{
		ApiId:      createdApi.ID(),
		Name:       pulumi.String(stageName),
		AutoDeploy: pulumi.BoolPtr(autoDeploy),
		Tags:       pulumi.ToStringMap(locals.AwsTags),
	}

	// Access logging
	if spec.Stage != nil && spec.Stage.AccessLog != nil {
		accessLog := spec.Stage.AccessLog
		args.AccessLogSettings = &apigatewayv2.StageAccessLogSettingsArgs{
			DestinationArn: pulumi.String(accessLog.DestinationArn.GetValue()),
			Format:         pulumi.String(accessLog.Format),
		}
	}

	// Default throttling
	if spec.Stage != nil && spec.Stage.DefaultThrottle != nil {
		throttle := spec.Stage.DefaultThrottle
		args.DefaultRouteSettings = &apigatewayv2.StageDefaultRouteSettingsArgs{
			ThrottlingBurstLimit: pulumi.IntPtr(int(throttle.BurstLimit)),
			ThrottlingRateLimit:  pulumi.Float64Ptr(throttle.RateLimit),
		}
	}

	// Stage variables
	if spec.Stage != nil && len(spec.Stage.StageVariables) > 0 {
		args.StageVariables = pulumi.ToStringMap(spec.Stage.StageVariables)
	}

	createdStage, err := apigatewayv2.NewStage(ctx, resourceName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create API stage")
	}

	// Export stage outputs
	ctx.Export(OpStageInvokeUrl, createdStage.InvokeUrl)
	ctx.Export(OpStageName, pulumi.String(stageName))

	return nil
}
