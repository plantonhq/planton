package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func client(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*cognito.UserPoolClient, error) {
	spec := locals.Spec

	// The client's cloud name is metadata.name -- the cross-engine naming
	// basis. Note the client is NOT a taggable AWS resource. The two booleans
	// are always sent (both engines state-pin them) so a true->false edit
	// takes effect instead of silently no-opping.
	args := &cognito.UserPoolClientArgs{
		Name:       pulumi.String(locals.Target.Metadata.Name),
		UserPoolId: pulumi.String(spec.UserPoolId.GetValue()),

		// ForceNew: confidential (secret-holding) vs public (PKCE) is decided
		// at creation.
		GenerateSecret: pulumi.BoolPtr(spec.GenerateSecret),

		AllowedOauthFlowsUserPoolClient: pulumi.BoolPtr(spec.AllowedOauthFlowsUserPoolClient),
	}

	// ---------------------------------------------------------------------------
	// OAuth 2.0 / OIDC contract
	// ---------------------------------------------------------------------------
	if len(spec.AllowedOauthFlows) > 0 {
		args.AllowedOauthFlows = pulumi.ToStringArray(spec.AllowedOauthFlows)
	}
	if len(spec.AllowedOauthScopes) > 0 {
		args.AllowedOauthScopes = pulumi.ToStringArray(spec.AllowedOauthScopes)
	}
	if len(spec.CallbackUrls) > 0 {
		args.CallbackUrls = pulumi.ToStringArray(spec.CallbackUrls)
	}
	if len(spec.LogoutUrls) > 0 {
		args.LogoutUrls = pulumi.ToStringArray(spec.LogoutUrls)
	}
	if spec.DefaultRedirectUri != "" {
		args.DefaultRedirectUri = pulumi.StringPtr(spec.DefaultRedirectUri)
	}

	// Provider names resolve from references (an AwsCognitoIdentityProvider's
	// provider_name output) or literals ("COGNITO", "Google") -- the reference
	// arm also gives the graph the right ordering: the IdP exists before the
	// client lists it.
	if len(spec.SupportedIdentityProviders) > 0 {
		providers := make([]string, 0, len(spec.SupportedIdentityProviders))
		for _, p := range spec.SupportedIdentityProviders {
			providers = append(providers, p.GetValue())
		}
		args.SupportedIdentityProviders = pulumi.ToStringArray(providers)
	}

	// ---------------------------------------------------------------------------
	// Authentication flows
	// ---------------------------------------------------------------------------

	if len(spec.ExplicitAuthFlows) > 0 {
		args.ExplicitAuthFlows = pulumi.ToStringArray(spec.ExplicitAuthFlows)
	}
	if spec.AuthSessionValidity != nil {
		args.AuthSessionValidity = pulumi.IntPtr(int(*spec.AuthSessionValidity))
	}

	// ---------------------------------------------------------------------------
	// Token lifetimes. Values pair with token_validity_units; omitted values
	// keep AWS defaults (1h access/ID, 30d refresh).
	// ---------------------------------------------------------------------------

	if spec.AccessTokenValidity != nil {
		args.AccessTokenValidity = pulumi.IntPtr(int(*spec.AccessTokenValidity))
	}
	if spec.IdTokenValidity != nil {
		args.IdTokenValidity = pulumi.IntPtr(int(*spec.IdTokenValidity))
	}
	if spec.RefreshTokenValidity != nil {
		args.RefreshTokenValidity = pulumi.IntPtr(int(*spec.RefreshTokenValidity))
	}
	if spec.TokenValidityUnits != nil {
		units := &cognito.UserPoolClientTokenValidityUnitsArgs{}
		if spec.TokenValidityUnits.AccessToken != "" {
			units.AccessToken = pulumi.StringPtr(spec.TokenValidityUnits.AccessToken)
		}
		if spec.TokenValidityUnits.IdToken != "" {
			units.IdToken = pulumi.StringPtr(spec.TokenValidityUnits.IdToken)
		}
		if spec.TokenValidityUnits.RefreshToken != "" {
			units.RefreshToken = pulumi.StringPtr(spec.TokenValidityUnits.RefreshToken)
		}
		args.TokenValidityUnits = units
	}

	if spec.RefreshTokenRotation != nil {
		rotation := &cognito.UserPoolClientRefreshTokenRotationArgs{
			Feature: pulumi.String(spec.RefreshTokenRotation.Feature),
		}
		// Tri-state: unset leaves the choice to AWS, explicit 0 pins
		// immediate retirement (grace disabled), 1-60 grants that many
		// seconds of grace.
		if spec.RefreshTokenRotation.RetryGracePeriodSeconds != nil {
			rotation.RetryGracePeriodSeconds = pulumi.IntPtr(int(*spec.RefreshTokenRotation.RetryGracePeriodSeconds))
		}
		args.RefreshTokenRotation = rotation
	}

	// ---------------------------------------------------------------------------
	// Security posture
	// ---------------------------------------------------------------------------

	// Optional bool: absent means AWS's default (true). Only an explicit
	// choice is forwarded so the module never silently flips revocability.
	if spec.EnableTokenRevocation != nil {
		args.EnableTokenRevocation = pulumi.BoolPtr(*spec.EnableTokenRevocation)
	}
	if spec.EnablePropagateAdditionalUserContextData {
		args.EnablePropagateAdditionalUserContextData = pulumi.BoolPtr(true)
	}
	if spec.PreventUserExistenceErrors != "" {
		args.PreventUserExistenceErrors = pulumi.StringPtr(spec.PreventUserExistenceErrors)
	}

	// ---------------------------------------------------------------------------
	// Attribute access
	// ---------------------------------------------------------------------------

	if len(spec.ReadAttributes) > 0 {
		args.ReadAttributes = pulumi.ToStringArray(spec.ReadAttributes)
	}
	if len(spec.WriteAttributes) > 0 {
		args.WriteAttributes = pulumi.ToStringArray(spec.WriteAttributes)
	}

	// ---------------------------------------------------------------------------
	// Pinpoint analytics
	// ---------------------------------------------------------------------------

	if spec.AnalyticsConfiguration != nil {
		ac := spec.AnalyticsConfiguration
		analytics := &cognito.UserPoolClientAnalyticsConfigurationArgs{}
		// Exactly one identity arm (spec CEL): the ARN arm derives the publish
		// role; the ID arm wires it explicitly.
		if ac.ApplicationArn != "" {
			analytics.ApplicationArn = pulumi.StringPtr(ac.ApplicationArn)
		}
		if ac.ApplicationId != "" {
			analytics.ApplicationId = pulumi.StringPtr(ac.ApplicationId)
			analytics.ExternalId = pulumi.StringPtr(ac.ExternalId)
			analytics.RoleArn = pulumi.StringPtr(ac.RoleArn.GetValue())
		}
		// Always sent (both engines state-pin it) so a true->false edit stops
		// sharing user data instead of silently no-opping.
		analytics.UserDataShared = pulumi.BoolPtr(ac.UserDataShared)
		args.AnalyticsConfiguration = analytics
	}

	// ---------------------------------------------------------------------------
	// Create app client
	// ---------------------------------------------------------------------------

	created, err := cognito.NewUserPoolClient(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cognito user pool client")
	}

	// ---------------------------------------------------------------------------
	// Exports. The secret is only minted when generate_secret is true; the
	// export is empty otherwise. Treat it as a credential.
	// ---------------------------------------------------------------------------

	ctx.Export(OpClientId, created.ID())
	ctx.Export(OpClientSecret, pulumi.ToSecret(created.ClientSecret))
	// Echo the resolved pool id: AWS keys clients by (pool id, client id), and
	// application configs typically need the pair together.
	ctx.Export(OpUserPoolId, created.UserPoolId)

	return created, nil
}
