package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createClient creates an Auth0 client (application) based on the configuration
func createClient(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Client, error) {
	// Build client arguments
	clientArgs := &auth0.ClientArgs{
		Name:        pulumi.String(locals.ClientName),
		AppType:     pulumi.String(locals.ApplicationType),
		Description: pulumi.String(locals.Description),
	}

	// Add logo URI if specified
	if locals.LogoUri != "" {
		clientArgs.LogoUri = pulumi.String(locals.LogoUri)
	}

	// Add callbacks if specified
	if len(locals.Callbacks) > 0 {
		callbacksArray := pulumi.StringArray{}
		for _, callback := range locals.Callbacks {
			callbacksArray = append(callbacksArray, pulumi.String(callback))
		}
		clientArgs.Callbacks = callbacksArray
	}

	// Add allowed logout URLs if specified
	if len(locals.AllowedLogoutUrls) > 0 {
		logoutUrlsArray := pulumi.StringArray{}
		for _, url := range locals.AllowedLogoutUrls {
			logoutUrlsArray = append(logoutUrlsArray, pulumi.String(url))
		}
		clientArgs.AllowedLogoutUrls = logoutUrlsArray
	}

	// Add web origins if specified
	if len(locals.WebOrigins) > 0 {
		webOriginsArray := pulumi.StringArray{}
		for _, origin := range locals.WebOrigins {
			webOriginsArray = append(webOriginsArray, pulumi.String(origin))
		}
		clientArgs.WebOrigins = webOriginsArray
	}

	// Add allowed origins if specified
	if len(locals.AllowedOrigins) > 0 {
		allowedOriginsArray := pulumi.StringArray{}
		for _, origin := range locals.AllowedOrigins {
			allowedOriginsArray = append(allowedOriginsArray, pulumi.String(origin))
		}
		clientArgs.AllowedOrigins = allowedOriginsArray
	}

	// Add grant types if specified
	if len(locals.GrantTypes) > 0 {
		grantTypesArray := pulumi.StringArray{}
		for _, grantType := range locals.GrantTypes {
			grantTypesArray = append(grantTypesArray, pulumi.String(grantType))
		}
		clientArgs.GrantTypes = grantTypesArray
	}

	// OAuth settings
	clientArgs.OidcConformant = pulumi.Bool(locals.OidcConformant)
	clientArgs.IsFirstParty = pulumi.Bool(locals.IsFirstParty)

	// Cross-origin settings
	clientArgs.CrossOriginAuth = pulumi.Bool(locals.CrossOriginAuthentication)
	if locals.CrossOriginLoc != "" {
		clientArgs.CrossOriginLoc = pulumi.String(locals.CrossOriginLoc)
	}

	// SSO settings
	clientArgs.Sso = pulumi.Bool(locals.Sso)
	clientArgs.SsoDisabled = pulumi.Bool(locals.SsoDisabled)

	// Custom login page
	if locals.CustomLoginPage != "" {
		clientArgs.CustomLoginPage = pulumi.String(locals.CustomLoginPage)
	}
	clientArgs.CustomLoginPageOn = pulumi.Bool(locals.CustomLoginPageOn)

	if locals.InitiateLoginUri != "" {
		clientArgs.InitiateLoginUri = pulumi.String(locals.InitiateLoginUri)
	}

	// Organization settings
	if locals.OrganizationUsage != "" {
		clientArgs.OrganizationUsage = pulumi.String(locals.OrganizationUsage)
	}
	if locals.OrganizationRequireBehavior != "" {
		clientArgs.OrganizationRequireBehavior = pulumi.String(locals.OrganizationRequireBehavior)
	}

	// JWT configuration
	if locals.JwtConfiguration != nil {
		jwtConfig := &auth0.ClientJwtConfigurationArgs{}
		hasJwtConfig := false

		if locals.JwtConfiguration.LifetimeInSeconds > 0 {
			jwtConfig.LifetimeInSeconds = pulumi.Int(int(locals.JwtConfiguration.LifetimeInSeconds))
			hasJwtConfig = true
		}

		// Default the id_token signing alg to RS256 (the proto's documented default and
		// Auth0's recommendation): JWKS-verifying clients (e.g. NextAuth) reject an HS256
		// id_token. Kept in parity with the tofu module's variables.tf
		// `alg = optional(string, "RS256")`.
		alg := locals.JwtConfiguration.Alg
		if alg == "" {
			alg = "RS256"
		}
		jwtConfig.Alg = pulumi.String(alg)
		hasJwtConfig = true

		jwtConfig.SecretEncoded = pulumi.Bool(locals.JwtConfiguration.SecretEncoded)

		if len(locals.JwtConfiguration.Scopes) > 0 {
			scopesMap := pulumi.StringMap{}
			for k, v := range locals.JwtConfiguration.Scopes {
				scopesMap[k] = pulumi.String(v)
			}
			jwtConfig.Scopes = scopesMap
			hasJwtConfig = true
		}

		if hasJwtConfig {
			clientArgs.JwtConfiguration = jwtConfig
		}
	}

	// Refresh token configuration
	if locals.RefreshToken != nil {
		refreshTokenConfig := &auth0.ClientRefreshTokenArgs{}
		hasRefreshConfig := false

		if locals.RefreshToken.RotationType != "" {
			refreshTokenConfig.RotationType = pulumi.String(locals.RefreshToken.RotationType)
			hasRefreshConfig = true
		}

		if locals.RefreshToken.ExpirationType != "" {
			refreshTokenConfig.ExpirationType = pulumi.String(locals.RefreshToken.ExpirationType)
			hasRefreshConfig = true
		}

		if locals.RefreshToken.TokenLifetime > 0 {
			refreshTokenConfig.TokenLifetime = pulumi.Int(int(locals.RefreshToken.TokenLifetime))
			hasRefreshConfig = true
		}

		if locals.RefreshToken.IdleTokenLifetime > 0 {
			refreshTokenConfig.IdleTokenLifetime = pulumi.Int(int(locals.RefreshToken.IdleTokenLifetime))
			hasRefreshConfig = true
		}

		refreshTokenConfig.InfiniteTokenLifetime = pulumi.Bool(locals.RefreshToken.InfiniteTokenLifetime)
		refreshTokenConfig.InfiniteIdleTokenLifetime = pulumi.Bool(locals.RefreshToken.InfiniteIdleTokenLifetime)

		if locals.RefreshToken.Leeway > 0 {
			refreshTokenConfig.Leeway = pulumi.Int(int(locals.RefreshToken.Leeway))
			hasRefreshConfig = true
		}

		if hasRefreshConfig {
			clientArgs.RefreshToken = refreshTokenConfig
		}
	}

	// Native social login configuration
	if locals.NativeSocialLogin != nil {
		nativeSocialLogin := &auth0.ClientNativeSocialLoginArgs{}
		hasNativeSocialLogin := false

		if locals.NativeSocialLogin.Apple != nil {
			nativeSocialLogin.Apple = &auth0.ClientNativeSocialLoginAppleArgs{
				Enabled: pulumi.Bool(locals.NativeSocialLogin.Apple.Enabled),
			}
			hasNativeSocialLogin = true
		}

		if locals.NativeSocialLogin.Facebook != nil {
			nativeSocialLogin.Facebook = &auth0.ClientNativeSocialLoginFacebookArgs{
				Enabled: pulumi.Bool(locals.NativeSocialLogin.Facebook.Enabled),
			}
			hasNativeSocialLogin = true
		}

		if hasNativeSocialLogin {
			clientArgs.NativeSocialLogin = nativeSocialLogin
		}
	}

	// Mobile configuration
	if locals.Mobile != nil {
		mobile := &auth0.ClientMobileArgs{}
		hasMobile := false

		if locals.Mobile.Ios != nil {
			iosArgs := &auth0.ClientMobileIosArgs{}
			if locals.Mobile.Ios.TeamId != "" {
				iosArgs.TeamId = pulumi.String(locals.Mobile.Ios.TeamId)
				hasMobile = true
			}
			if locals.Mobile.Ios.AppBundleIdentifier != "" {
				iosArgs.AppBundleIdentifier = pulumi.String(locals.Mobile.Ios.AppBundleIdentifier)
				hasMobile = true
			}
			mobile.Ios = iosArgs
		}

		if locals.Mobile.Android != nil {
			androidArgs := &auth0.ClientMobileAndroidArgs{}
			if locals.Mobile.Android.AppPackageName != "" {
				androidArgs.AppPackageName = pulumi.String(locals.Mobile.Android.AppPackageName)
				hasMobile = true
			}
			if len(locals.Mobile.Android.Sha256CertFingerprints) > 0 {
				fingerprintsArray := pulumi.StringArray{}
				for _, fp := range locals.Mobile.Android.Sha256CertFingerprints {
					fingerprintsArray = append(fingerprintsArray, pulumi.String(fp))
				}
				androidArgs.Sha256CertFingerprints = fingerprintsArray
				hasMobile = true
			}
			mobile.Android = androidArgs
		}

		if hasMobile {
			clientArgs.Mobile = mobile
		}
	}

	// Client metadata
	if len(locals.ClientMetadata) > 0 {
		metadataMap := pulumi.StringMap{}
		for k, v := range locals.ClientMetadata {
			metadataMap[k] = pulumi.String(v)
		}
		clientArgs.ClientMetadata = metadataMap
	}

	// Client aliases
	if len(locals.ClientAliases) > 0 {
		aliasesArray := pulumi.StringArray{}
		for _, alias := range locals.ClientAliases {
			aliasesArray = append(aliasesArray, pulumi.String(alias))
		}
		clientArgs.ClientAliases = aliasesArray
	}

	// Token endpoint IP header trust
	clientArgs.IsTokenEndpointIpHeaderTrusted = pulumi.Bool(locals.IsTokenEndpointIpHeaderTrusted)

	// Note: OIDC backchannel logout may not be supported in all versions of the SDK
	// If the SDK version supports it, uncomment and use the following:
	// if locals.OidcBackchannelLogout != nil && len(locals.OidcBackchannelLogout.BackchannelLogoutUrls) > 0 {
	//     // Configure backchannel logout URLs
	// }

	// Create the client resource
	client, err := auth0.NewClient(ctx, locals.ClientName, clientArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 client %s", locals.ClientName)
	}

	return client, nil
}

// createClientGrants creates Auth0 client grants to authorize API access
func createClientGrants(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider, client *auth0.Client) error {
	if len(locals.ApiGrants) == 0 {
		return nil
	}

	for i, grant := range locals.ApiGrants {
		if grant == nil || grant.Audience == "" {
			continue
		}

		// Build scopes array
		scopesArray := pulumi.StringArray{}
		for _, scope := range grant.Scopes {
			scopesArray = append(scopesArray, pulumi.String(scope))
		}

		// Build client grant arguments (audience is already resolved from StringValueOrRef)
		grantArgs := &auth0.ClientGrantArgs{
			ClientId: client.ClientId,
			Audience: pulumi.String(grant.Audience),
			Scopes:   scopesArray,
		}

		// Add organization_usage if specified
		if grant.OrganizationUsage != "" {
			grantArgs.OrganizationUsage = pulumi.String(grant.OrganizationUsage)
		}

		// Add allow_any_organization if set
		if grant.AllowAnyOrganization {
			grantArgs.AllowAnyOrganization = pulumi.Bool(grant.AllowAnyOrganization)
		}

		// Create the client grant resource
		grantName := fmt.Sprintf("%s-grant-%d", locals.ClientName, i)
		_, err := auth0.NewClientGrant(ctx, grantName, grantArgs,
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{client}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create Auth0 client grant %s for audience %s", grantName, grant.Audience)
		}
	}

	return nil
}
