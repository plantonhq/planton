package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func httpApi(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*apigatewayv2.Api, error) {
	spec := locals.Spec

	args := &apigatewayv2.ApiArgs{
		Name:         pulumi.String(locals.ApiName),
		ProtocolType: pulumi.String("HTTP"),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Disable default execute-api endpoint
	if spec.DisableExecuteApiEndpoint {
		args.DisableExecuteApiEndpoint = pulumi.BoolPtr(true)
	}

	// CORS configuration
	if spec.CorsConfiguration != nil {
		cors := spec.CorsConfiguration
		corsArgs := &apigatewayv2.ApiCorsConfigurationArgs{
			AllowCredentials: pulumi.BoolPtr(cors.AllowCredentials),
		}
		if len(cors.AllowOrigins) > 0 {
			corsArgs.AllowOrigins = pulumi.ToStringArray(cors.AllowOrigins)
		}
		if len(cors.AllowMethods) > 0 {
			corsArgs.AllowMethods = pulumi.ToStringArray(cors.AllowMethods)
		}
		if len(cors.AllowHeaders) > 0 {
			corsArgs.AllowHeaders = pulumi.ToStringArray(cors.AllowHeaders)
		}
		if len(cors.ExposeHeaders) > 0 {
			corsArgs.ExposeHeaders = pulumi.ToStringArray(cors.ExposeHeaders)
		}
		if cors.MaxAgeSeconds > 0 {
			corsArgs.MaxAge = pulumi.IntPtr(int(cors.MaxAgeSeconds))
		}
		args.CorsConfiguration = corsArgs
	}

	// Create the API
	createdApi, err := apigatewayv2.NewApi(ctx, locals.ApiName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP API")
	}

	// Export API-level outputs
	ctx.Export(OpApiId, createdApi.ID())
	ctx.Export(OpApiEndpoint, createdApi.ApiEndpoint)
	ctx.Export(OpApiArn, createdApi.Arn)
	ctx.Export(OpExecutionArn, createdApi.ExecutionArn)

	return createdApi, nil
}
