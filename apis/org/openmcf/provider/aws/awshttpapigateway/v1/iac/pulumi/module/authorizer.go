package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// authorizers creates named authorizers and returns a map of authorizer name
// to the created Authorizer resource. Routes reference authorizers by name.
func authorizers(
	ctx *pulumi.Context,
	locals *Locals,
	createdApi *apigatewayv2.Api,
	provider *aws.Provider,
) (map[string]*apigatewayv2.Authorizer, error) {
	result := make(map[string]*apigatewayv2.Authorizer)

	if len(locals.Spec.Authorizers) == 0 {
		return result, nil
	}

	for i, auth := range locals.Spec.Authorizers {
		resourceName := fmt.Sprintf("%s-authorizer-%s", locals.ApiName, auth.Name)

		args := &apigatewayv2.AuthorizerArgs{
			ApiId:          createdApi.ID(),
			Name:           pulumi.StringPtr(auth.Name),
			AuthorizerType: pulumi.String(auth.AuthorizerType),
		}

		// Identity sources
		if len(auth.IdentitySources) > 0 {
			args.IdentitySources = pulumi.ToStringArray(auth.IdentitySources)
		}

		// JWT configuration
		if auth.AuthorizerType == "JWT" && auth.JwtConfiguration != nil {
			jwtConfig := &apigatewayv2.AuthorizerJwtConfigurationArgs{
				Issuer: pulumi.StringPtr(auth.JwtConfiguration.Issuer),
			}
			if len(auth.JwtConfiguration.Audiences) > 0 {
				jwtConfig.Audiences = pulumi.ToStringArray(auth.JwtConfiguration.Audiences)
			}
			args.JwtConfiguration = jwtConfig
		}

		// Lambda authorizer (REQUEST type)
		if auth.AuthorizerType == "REQUEST" {
			if auth.AuthorizerUri.GetValue() != "" {
				args.AuthorizerUri = pulumi.StringPtr(auth.AuthorizerUri.GetValue())
			}
			if auth.AuthorizerCredentialsArn.GetValue() != "" {
				args.AuthorizerCredentialsArn = pulumi.StringPtr(auth.AuthorizerCredentialsArn.GetValue())
			}
			if auth.EnableSimpleResponses {
				args.EnableSimpleResponses = pulumi.BoolPtr(true)
			}
			if auth.AuthorizerPayloadFormatVersion != "" {
				args.AuthorizerPayloadFormatVersion = pulumi.StringPtr(auth.AuthorizerPayloadFormatVersion)
			}
		}

		// Cache TTL
		if auth.ResultTtlSeconds > 0 {
			args.AuthorizerResultTtlInSeconds = pulumi.IntPtr(int(auth.ResultTtlSeconds))
		}

		created, err := apigatewayv2.NewAuthorizer(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create authorizer %q (index %d)", auth.Name, i)
		}

		result[auth.Name] = created
	}

	return result, nil
}
