package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func clients(ctx *pulumi.Context, locals *Locals, createdPool *cognito.UserPool, provider *aws.Provider) error {
	clientIdMap := pulumi.StringMap{}
	clientSecretMap := pulumi.StringMap{}

	for _, client := range locals.Spec.Clients {
		resourceName := fmt.Sprintf("%s-%s", locals.Target.Metadata.Name, client.Name)

		args := &cognito.UserPoolClientArgs{
			Name:       pulumi.String(client.Name),
			UserPoolId: createdPool.ID(),
		}

		// -----------------------------------------------------------
		// Secret generation
		// -----------------------------------------------------------

		if client.GenerateSecret {
			args.GenerateSecret = pulumi.BoolPtr(true)
		}

		// -----------------------------------------------------------
		// OAuth configuration
		// -----------------------------------------------------------

		if client.AllowedOauthFlowsUserPoolClient {
			args.AllowedOauthFlowsUserPoolClient = pulumi.BoolPtr(true)
		}

		if len(client.AllowedOauthFlows) > 0 {
			args.AllowedOauthFlows = pulumi.ToStringArray(client.AllowedOauthFlows)
		}

		if len(client.AllowedOauthScopes) > 0 {
			args.AllowedOauthScopes = pulumi.ToStringArray(client.AllowedOauthScopes)
		}

		if len(client.CallbackUrls) > 0 {
			args.CallbackUrls = pulumi.ToStringArray(client.CallbackUrls)
		}

		if len(client.LogoutUrls) > 0 {
			args.LogoutUrls = pulumi.ToStringArray(client.LogoutUrls)
		}

		if client.DefaultRedirectUri != "" {
			args.DefaultRedirectUri = pulumi.StringPtr(client.DefaultRedirectUri)
		}

		if len(client.SupportedIdentityProviders) > 0 {
			args.SupportedIdentityProviders = pulumi.ToStringArray(client.SupportedIdentityProviders)
		}

		// -----------------------------------------------------------
		// Authentication flows
		// -----------------------------------------------------------

		if len(client.ExplicitAuthFlows) > 0 {
			args.ExplicitAuthFlows = pulumi.ToStringArray(client.ExplicitAuthFlows)
		}

		// -----------------------------------------------------------
		// Token validity (units hardcoded to match spec field semantics)
		// -----------------------------------------------------------

		if client.AccessTokenValidityMinutes > 0 {
			args.AccessTokenValidity = pulumi.IntPtr(int(client.AccessTokenValidityMinutes))
			args.TokenValidityUnits = &cognito.UserPoolClientTokenValidityUnitsArgs{
				AccessToken:  pulumi.StringPtr("minutes"),
				IdToken:      pulumi.StringPtr("minutes"),
				RefreshToken: pulumi.StringPtr("days"),
			}
		}

		if client.IdTokenValidityMinutes > 0 {
			args.IdTokenValidity = pulumi.IntPtr(int(client.IdTokenValidityMinutes))
			// Ensure token_validity_units is set (may already be set by access token above)
			if args.TokenValidityUnits == nil {
				args.TokenValidityUnits = &cognito.UserPoolClientTokenValidityUnitsArgs{
					AccessToken:  pulumi.StringPtr("minutes"),
					IdToken:      pulumi.StringPtr("minutes"),
					RefreshToken: pulumi.StringPtr("days"),
				}
			}
		}

		if client.RefreshTokenValidityDays > 0 {
			args.RefreshTokenValidity = pulumi.IntPtr(int(client.RefreshTokenValidityDays))
			if args.TokenValidityUnits == nil {
				args.TokenValidityUnits = &cognito.UserPoolClientTokenValidityUnitsArgs{
					AccessToken:  pulumi.StringPtr("minutes"),
					IdToken:      pulumi.StringPtr("minutes"),
					RefreshToken: pulumi.StringPtr("days"),
				}
			}
		}

		// -----------------------------------------------------------
		// Security settings
		// -----------------------------------------------------------

		if client.EnableTokenRevocation {
			args.EnableTokenRevocation = pulumi.BoolPtr(true)
		}

		if client.PreventUserExistenceErrors != "" {
			args.PreventUserExistenceErrors = pulumi.StringPtr(client.PreventUserExistenceErrors)
		}

		// -----------------------------------------------------------
		// Create client
		// -----------------------------------------------------------

		created, err := cognito.NewUserPoolClient(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create Cognito user pool client %q", client.Name)
		}

		clientIdMap[client.Name] = created.ID()

		if client.GenerateSecret {
			clientSecretMap[client.Name] = created.ClientSecret
		}
	}

	// Export client ID and secret maps.
	ctx.Export(OpClientIds, clientIdMap)
	ctx.Export(OpClientSecrets, clientSecretMap)

	return nil
}
