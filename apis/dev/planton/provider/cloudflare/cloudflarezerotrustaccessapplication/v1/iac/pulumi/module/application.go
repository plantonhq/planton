package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccessapplicationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrustaccessapplication/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// application provisions the Cloudflare Zero Trust Access application, wiring it to
// standalone Access policies by reference, and exports its outputs.
func application(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ZeroTrustAccessApplication, error) {
	spec := locals.CloudflareZeroTrustAccessApplication.Spec

	appType := "self_hosted"
	if spec.Type != cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessApplicationType_application_type_unspecified {
		appType = spec.Type.String()
	}

	args := &cloudflare.ZeroTrustAccessApplicationArgs{
		Name: pulumi.String(spec.Name),
		Type: pulumi.String(appType),
	}
	if spec.AccountId != "" {
		args.AccountId = pulumi.StringPtr(spec.AccountId)
	}
	if spec.ZoneId != nil && spec.ZoneId.GetValue() != "" {
		args.ZoneId = pulumi.StringPtr(spec.ZoneId.GetValue())
	}
	if spec.Domain != "" {
		args.Domain = pulumi.StringPtr(spec.Domain)
	}

	// Policies, referenced by ID with an evaluation precedence.
	if len(spec.Policies) > 0 {
		var policies cloudflare.ZeroTrustAccessApplicationPolicyArray
		for _, p := range spec.Policies {
			pa := &cloudflare.ZeroTrustAccessApplicationPolicyArgs{Id: pulumi.String(p.Policy.GetValue())}
			if p.Precedence > 0 {
				pa.Precedence = pulumi.IntPtr(int(p.Precedence))
			}
			policies = append(policies, pa)
		}
		args.Policies = policies
	}

	// Allowed identity providers.
	if len(spec.AllowedIdps) > 0 {
		idps := make([]string, 0, len(spec.AllowedIdps))
		for _, idp := range spec.AllowedIdps {
			if idp.GetValue() != "" {
				idps = append(idps, idp.GetValue())
			}
		}
		if len(idps) > 0 {
			args.AllowedIdps = pulumi.ToStringArray(idps)
		}
	}

	// Destinations.
	if len(spec.Destinations) > 0 {
		var dests cloudflare.ZeroTrustAccessApplicationDestinationArray
		for _, d := range spec.Destinations {
			da := &cloudflare.ZeroTrustAccessApplicationDestinationArgs{}
			if d.Type != "" {
				da.Type = pulumi.StringPtr(d.Type)
			}
			if d.Uri != "" {
				da.Uri = pulumi.StringPtr(d.Uri)
			}
			if d.Cidr != "" {
				da.Cidr = pulumi.StringPtr(d.Cidr)
			}
			if d.Hostname != "" {
				da.Hostname = pulumi.StringPtr(d.Hostname)
			}
			if d.L4Protocol != "" {
				da.L4Protocol = pulumi.StringPtr(d.L4Protocol)
			}
			if d.PortRange != "" {
				da.PortRange = pulumi.StringPtr(d.PortRange)
			}
			if d.VnetId != "" {
				da.VnetId = pulumi.StringPtr(d.VnetId)
			}
			if d.McpServerId != "" {
				da.McpServerId = pulumi.StringPtr(d.McpServerId)
			}
			dests = append(dests, da)
		}
		args.Destinations = dests
	}

	// Scalar toggles and strings.
	// Type-restricted toggles: the provider rejects these (even when false) on
	// incompatible application types, so send them only when enabled (false == the
	// provider default == omitted).
	if spec.AutoRedirectToIdentity {
		args.AutoRedirectToIdentity = pulumi.BoolPtr(true)
	}
	if spec.SkipAppLauncherLoginPage {
		args.SkipAppLauncherLoginPage = pulumi.BoolPtr(true)
	}
	if spec.AllowAuthenticateViaWarp {
		args.AllowAuthenticateViaWarp = pulumi.BoolPtr(true)
	}
	if spec.AllowIframe {
		args.AllowIframe = pulumi.BoolPtr(true)
	}
	if spec.OptionsPreflightBypass {
		args.OptionsPreflightBypass = pulumi.BoolPtr(true)
	}
	if spec.ServiceAuth_401Redirect {
		args.ServiceAuth401Redirect = pulumi.BoolPtr(true)
	}
	if spec.SkipInterstitial {
		args.SkipInterstitial = pulumi.BoolPtr(true)
	}
	if spec.EnableBindingCookie {
		args.EnableBindingCookie = pulumi.BoolPtr(true)
	}
	if spec.PathCookieAttribute {
		args.PathCookieAttribute = pulumi.BoolPtr(true)
	}
	// Optional+computed bools: only set when explicitly provided so the provider's
	// own default applies otherwise.
	if spec.AppLauncherVisible != nil {
		args.AppLauncherVisible = pulumi.BoolPtr(*spec.AppLauncherVisible)
	}
	if spec.HttpOnlyCookieAttribute != nil {
		args.HttpOnlyCookieAttribute = pulumi.BoolPtr(*spec.HttpOnlyCookieAttribute)
	}
	if spec.SessionDuration != "" {
		args.SessionDuration = pulumi.StringPtr(spec.SessionDuration)
	}
	if len(spec.Tags) > 0 {
		args.Tags = pulumi.ToStringArray(spec.Tags)
	}
	if len(spec.CustomPages) > 0 {
		args.CustomPages = pulumi.ToStringArray(spec.CustomPages)
	}
	if spec.AppLauncherLogoUrl != "" {
		args.AppLauncherLogoUrl = pulumi.StringPtr(spec.AppLauncherLogoUrl)
	}
	if spec.BgColor != "" {
		args.BgColor = pulumi.StringPtr(spec.BgColor)
	}
	if spec.HeaderBgColor != "" {
		args.HeaderBgColor = pulumi.StringPtr(spec.HeaderBgColor)
	}
	if spec.LogoUrl != "" {
		args.LogoUrl = pulumi.StringPtr(spec.LogoUrl)
	}
	if spec.ReadServiceTokensFromHeader != "" {
		args.ReadServiceTokensFromHeader = pulumi.StringPtr(spec.ReadServiceTokensFromHeader)
	}
	if spec.SameSiteCookieAttribute != "" {
		args.SameSiteCookieAttribute = pulumi.StringPtr(spec.SameSiteCookieAttribute)
	}
	if spec.CustomDenyMessage != "" {
		args.CustomDenyMessage = pulumi.StringPtr(spec.CustomDenyMessage)
	}
	if spec.CustomDenyUrl != "" {
		args.CustomDenyUrl = pulumi.StringPtr(spec.CustomDenyUrl)
	}
	if spec.CustomNonIdentityDenyUrl != "" {
		args.CustomNonIdentityDenyUrl = pulumi.StringPtr(spec.CustomNonIdentityDenyUrl)
	}

	// Landing page design.
	if lp := spec.LandingPageDesign; lp != nil {
		args.LandingPageDesign = &cloudflare.ZeroTrustAccessApplicationLandingPageDesignArgs{
			Title:           strPtrOrNil(lp.Title),
			Message:         strPtrOrNil(lp.Message),
			ImageUrl:        strPtrOrNil(lp.ImageUrl),
			ButtonColor:     strPtrOrNil(lp.ButtonColor),
			ButtonTextColor: strPtrOrNil(lp.ButtonTextColor),
		}
	}

	// Footer links.
	if len(spec.FooterLinks) > 0 {
		var links cloudflare.ZeroTrustAccessApplicationFooterLinkArray
		for _, l := range spec.FooterLinks {
			links = append(links, &cloudflare.ZeroTrustAccessApplicationFooterLinkArgs{
				Name: pulumi.String(l.Name),
				Url:  pulumi.String(l.Url),
			})
		}
		args.FooterLinks = links
	}

	// CORS headers.
	if c := spec.CorsHeaders; c != nil {
		ch := &cloudflare.ZeroTrustAccessApplicationCorsHeadersArgs{
			AllowAllHeaders:  pulumi.BoolPtr(c.AllowAllHeaders),
			AllowAllMethods:  pulumi.BoolPtr(c.AllowAllMethods),
			AllowAllOrigins:  pulumi.BoolPtr(c.AllowAllOrigins),
			AllowCredentials: pulumi.BoolPtr(c.AllowCredentials),
		}
		if len(c.AllowedHeaders) > 0 {
			ch.AllowedHeaders = pulumi.ToStringArray(c.AllowedHeaders)
		}
		if len(c.AllowedMethods) > 0 {
			methods := make([]string, 0, len(c.AllowedMethods))
			for _, m := range c.AllowedMethods {
				methods = append(methods, m.String())
			}
			ch.AllowedMethods = pulumi.ToStringArray(methods)
		}
		if len(c.AllowedOrigins) > 0 {
			ch.AllowedOrigins = pulumi.ToStringArray(c.AllowedOrigins)
		}
		if c.MaxAge != 0 {
			ch.MaxAge = pulumi.Float64Ptr(float64(c.MaxAge))
		}
		args.CorsHeaders = ch
	}

	// Application-level MFA.
	if m := spec.MfaConfig; m != nil {
		mfa := &cloudflare.ZeroTrustAccessApplicationMfaConfigArgs{MfaDisabled: pulumi.BoolPtr(m.MfaDisabled)}
		if len(m.AllowedAuthenticators) > 0 {
			auths := make([]string, 0, len(m.AllowedAuthenticators))
			for _, a := range m.AllowedAuthenticators {
				auths = append(auths, a.String())
			}
			mfa.AllowedAuthenticators = pulumi.ToStringArray(auths)
		}
		if m.SessionDuration != "" {
			mfa.SessionDuration = pulumi.StringPtr(m.SessionDuration)
		}
		args.MfaConfig = mfa
	}

	// OAuth configuration (MCP authorization server).
	if o := spec.OauthConfiguration; o != nil {
		oc := &cloudflare.ZeroTrustAccessApplicationOauthConfigurationArgs{Enabled: pulumi.BoolPtr(o.Enabled)}
		if d := o.DynamicClientRegistration; d != nil {
			dcr := &cloudflare.ZeroTrustAccessApplicationOauthConfigurationDynamicClientRegistrationArgs{
				Enabled:             pulumi.BoolPtr(d.Enabled),
				AllowAnyOnLocalhost: pulumi.BoolPtr(d.AllowAnyOnLocalhost),
				AllowAnyOnLoopback:  pulumi.BoolPtr(d.AllowAnyOnLoopback),
			}
			if len(d.AllowedUris) > 0 {
				dcr.AllowedUris = pulumi.ToStringArray(d.AllowedUris)
			}
			oc.DynamicClientRegistration = dcr
		}
		if g := o.Grant; g != nil {
			oc.Grant = &cloudflare.ZeroTrustAccessApplicationOauthConfigurationGrantArgs{
				AccessTokenLifetime: strPtrOrNil(g.AccessTokenLifetime),
				SessionDuration:     strPtrOrNil(g.SessionDuration),
			}
		}
		args.OauthConfiguration = oc
	}

	// Target criteria for rdp/infrastructure apps.
	if len(spec.TargetCriteria) > 0 {
		var criteria cloudflare.ZeroTrustAccessApplicationTargetCriteriaArray
		for _, tc := range spec.TargetCriteria {
			attrs := pulumi.StringArrayMap{}
			for _, a := range tc.TargetAttributes {
				attrs[a.Name] = pulumi.ToStringArray(a.Values)
			}
			criteria = append(criteria, &cloudflare.ZeroTrustAccessApplicationTargetCriteriaArgs{
				Port:             pulumi.Int(int(tc.Port)),
				Protocol:         pulumi.String(tc.Protocol.String()),
				TargetAttributes: attrs,
			})
		}
		args.TargetCriterias = criteria
	}

	// SaaS application (SAML / OIDC).
	if s := spec.SaasApp; s != nil {
		args.SaasApp = buildSaasApp(s)
	}

	// SCIM provisioning.
	if sc := spec.ScimConfig; sc != nil {
		args.ScimConfig = buildScimConfig(sc)
	}

	created, err := cloudflare.NewZeroTrustAccessApplication(
		ctx,
		"access_application",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access application")
	}

	ctx.Export(OpApplicationId, created.ID())
	ctx.Export(OpAud, created.Aud)
	ctx.Export(OpDomain, created.Domain)
	ctx.Export(OpSaasClientId, created.SaasApp.ClientId())
	ctx.Export(OpSaasClientSecret, created.SaasApp.ClientSecret())
	ctx.Export(OpSaasPublicKey, created.SaasApp.PublicKey())
	ctx.Export(OpSaasSsoEndpoint, created.SaasApp.SsoEndpoint())
	ctx.Export(OpSaasIdpEntityId, created.SaasApp.IdpEntityId())

	return created, nil
}

func strPtrOrNil(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.StringPtr(s)
}

func buildSaasApp(s *cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessSaasApp) *cloudflare.ZeroTrustAccessApplicationSaasAppArgs {
	sa := &cloudflare.ZeroTrustAccessApplicationSaasAppArgs{}
	if s.AuthType != cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessSaasApp_auth_type_unspecified {
		sa.AuthType = pulumi.StringPtr(s.AuthType.String())
	}
	// SAML.
	sa.ConsumerServiceUrl = strPtrOrNil(s.ConsumerServiceUrl)
	sa.SpEntityId = strPtrOrNil(s.SpEntityId)
	sa.NameIdTransformJsonata = strPtrOrNil(s.NameIdTransformJsonata)
	sa.SamlAttributeTransformJsonata = strPtrOrNil(s.SamlAttributeTransformJsonata)
	sa.DefaultRelayState = strPtrOrNil(s.DefaultRelayState)
	if s.NameIdFormat != cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessSaasApp_name_id_format_unspecified {
		sa.NameIdFormat = pulumi.StringPtr(s.NameIdFormat.String())
	}
	if len(s.CustomAttributes) > 0 {
		var attrs cloudflare.ZeroTrustAccessApplicationSaasAppCustomAttributeArray
		for _, a := range s.CustomAttributes {
			ca := &cloudflare.ZeroTrustAccessApplicationSaasAppCustomAttributeArgs{
				Name:         strPtrOrNil(a.Name),
				FriendlyName: strPtrOrNil(a.FriendlyName),
				NameFormat:   strPtrOrNil(a.NameFormat),
				Required:     pulumi.BoolPtr(a.Required),
			}
			if a.Source != nil {
				src := &cloudflare.ZeroTrustAccessApplicationSaasAppCustomAttributeSourceArgs{Name: strPtrOrNil(a.Source.Name)}
				if len(a.Source.NameByIdp) > 0 {
					var nbi cloudflare.ZeroTrustAccessApplicationSaasAppCustomAttributeSourceNameByIdpArray
					for _, n := range a.Source.NameByIdp {
						nbi = append(nbi, &cloudflare.ZeroTrustAccessApplicationSaasAppCustomAttributeSourceNameByIdpArgs{
							IdpId:      strPtrOrNil(n.IdpId.GetValue()),
							SourceName: strPtrOrNil(n.SourceName),
						})
					}
					src.NameByIdps = nbi
				}
				ca.Source = src
			}
			attrs = append(attrs, ca)
		}
		sa.CustomAttributes = attrs
	}
	// OIDC.
	if len(s.RedirectUris) > 0 {
		sa.RedirectUris = pulumi.ToStringArray(s.RedirectUris)
	}
	if len(s.GrantTypes) > 0 {
		gt := make([]string, 0, len(s.GrantTypes))
		for _, g := range s.GrantTypes {
			gt = append(gt, g.String())
		}
		sa.GrantTypes = pulumi.ToStringArray(gt)
	}
	if len(s.Scopes) > 0 {
		sc := make([]string, 0, len(s.Scopes))
		for _, x := range s.Scopes {
			sc = append(sc, x.String())
		}
		sa.Scopes = pulumi.ToStringArray(sc)
	}
	sa.GroupFilterRegex = strPtrOrNil(s.GroupFilterRegex)
	sa.AppLauncherUrl = strPtrOrNil(s.AppLauncherUrl)
	sa.AccessTokenLifetime = strPtrOrNil(s.AccessTokenLifetime)
	sa.AllowPkceWithoutClientSecret = pulumi.BoolPtr(s.AllowPkceWithoutClientSecret)
	if len(s.CustomClaims) > 0 {
		var claims cloudflare.ZeroTrustAccessApplicationSaasAppCustomClaimArray
		for _, c := range s.CustomClaims {
			cc := &cloudflare.ZeroTrustAccessApplicationSaasAppCustomClaimArgs{
				Name:     strPtrOrNil(c.Name),
				Required: pulumi.BoolPtr(c.Required),
			}
			if c.Scope != cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessSaasScope_scope_unspecified {
				cc.Scope = pulumi.StringPtr(c.Scope.String())
			}
			if c.Source != nil {
				csrc := &cloudflare.ZeroTrustAccessApplicationSaasAppCustomClaimSourceArgs{Name: strPtrOrNil(c.Source.Name)}
				if len(c.Source.NameByIdp) > 0 {
					m := pulumi.StringMap{}
					for k, val := range c.Source.NameByIdp {
						m[k] = pulumi.String(val)
					}
					csrc.NameByIdp = m
				}
				cc.Source = csrc
			}
			claims = append(claims, cc)
		}
		sa.CustomClaims = claims
	}
	if h := s.HybridAndImplicitOptions; h != nil {
		sa.HybridAndImplicitOptions = &cloudflare.ZeroTrustAccessApplicationSaasAppHybridAndImplicitOptionsArgs{
			ReturnAccessTokenFromAuthorizationEndpoint: pulumi.BoolPtr(h.ReturnAccessTokenFromAuthorizationEndpoint),
			ReturnIdTokenFromAuthorizationEndpoint:     pulumi.BoolPtr(h.ReturnIdTokenFromAuthorizationEndpoint),
		}
	}
	if rt := s.RefreshTokenOptions; rt != nil {
		sa.RefreshTokenOptions = &cloudflare.ZeroTrustAccessApplicationSaasAppRefreshTokenOptionsArgs{Lifetime: strPtrOrNil(rt.Lifetime)}
	}
	return sa
}

func buildScimConfig(sc *cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessScimConfig) *cloudflare.ZeroTrustAccessApplicationScimConfigArgs {
	cfg := &cloudflare.ZeroTrustAccessApplicationScimConfigArgs{
		IdpUid:             pulumi.String(sc.IdpUid.GetValue()),
		RemoteUri:          pulumi.String(sc.RemoteUri),
		Enabled:            pulumi.BoolPtr(sc.Enabled),
		DeactivateOnDelete: pulumi.BoolPtr(sc.DeactivateOnDelete),
	}
	if a := sc.Authentication; a != nil {
		auth := &cloudflare.ZeroTrustAccessApplicationScimConfigAuthenticationArgs{
			Scheme:           pulumi.String(a.Scheme.String()),
			User:             strPtrOrNil(a.User),
			Password:         strPtrOrNil(a.Password),
			Token:            strPtrOrNil(a.Token),
			ClientId:         strPtrOrNil(a.ClientId),
			ClientSecret:     strPtrOrNil(a.ClientSecret),
			AuthorizationUrl: strPtrOrNil(a.AuthorizationUrl),
			TokenUrl:         strPtrOrNil(a.TokenUrl),
		}
		if len(a.Scopes) > 0 {
			auth.Scopes = pulumi.ToStringArray(a.Scopes)
		}
		cfg.Authentication = auth
	}
	if len(sc.Mappings) > 0 {
		var mappings cloudflare.ZeroTrustAccessApplicationScimConfigMappingArray
		for _, m := range sc.Mappings {
			ma := &cloudflare.ZeroTrustAccessApplicationScimConfigMappingArgs{
				Schema:           pulumi.String(m.Schema),
				Enabled:          pulumi.BoolPtr(m.Enabled),
				Filter:           strPtrOrNil(m.Filter),
				TransformJsonata: strPtrOrNil(m.TransformJsonata),
			}
			if m.Strictness != cloudflarezerotrustaccessapplicationv1.CloudflareZeroTrustAccessScimMapping_strictness_unspecified {
				ma.Strictness = pulumi.StringPtr(m.Strictness.String())
			}
			if o := m.Operations; o != nil {
				ma.Operations = &cloudflare.ZeroTrustAccessApplicationScimConfigMappingOperationsArgs{
					Create: pulumi.BoolPtr(o.Create),
					Update: pulumi.BoolPtr(o.Update),
					Delete: pulumi.BoolPtr(o.Delete),
				}
			}
			mappings = append(mappings, ma)
		}
		cfg.Mappings = mappings
	}
	return cfg
}
