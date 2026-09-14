package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/identityplatform"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityPlatformConfig provisions the project's Identity Platform
// configuration plus its composed identity-provider configs.
//
// The config resource is a ONE-WAY project singleton: the provider's
// create is a bare initializeAuth call (permanently enabling Identity
// Platform on the project, billing required) followed by an update that
// applies every setting, and its delete abandons the configuration in
// place — GCP has no de-initialize. That is why the resource carries no
// deletion_policy: the spec's deletion_policy governs only the composed
// IdP configs below.
//
// Every sign-in arm's `enabled` is sent EXPLICITLY whenever its message is
// present: the fields drive live authentication surfaces, and a spec
// transition true -> false must reach the API rather than being omitted
// (the send-true-or-omit class).
func identityPlatformConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpIdentityPlatformConfig.Spec

	// Enable the Identity Toolkit API so a fresh project can be
	// initialized. disable_on_destroy stays false: tearing down this
	// resource must never disable authentication for everything else in
	// the project. Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("identitytoolkit.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"config-identitytoolkit.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable identitytoolkit.googleapis.com api")
	}

	args := &identityplatform.ConfigArgs{}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if len(spec.AuthorizedDomains) > 0 {
		args.AuthorizedDomains = pulumi.ToStringArray(spec.AuthorizedDomains)
	}
	if spec.AutodeleteAnonymousUsers {
		args.AutodeleteAnonymousUsers = pulumi.BoolPtr(true)
	}

	if spec.SignIn != nil {
		signInArgs := &identityplatform.ConfigSignInArgs{
			AllowDuplicateEmails: pulumi.BoolPtr(spec.SignIn.AllowDuplicateEmails),
		}
		// The API always materializes the email and phone_number policies in
		// its read-back (enabled=false when never configured), so BOTH args
		// are set explicitly whenever sign_in is present — omitting one
		// leaves a perpetual diff on every re-preview (idempotency-gate
		// caught, live-verified). The anonymous arm is NOT echoed when unset
		// and is set only when the spec sets it.
		emailArgs := &identityplatform.ConfigSignInEmailArgs{
			Enabled: pulumi.Bool(spec.SignIn.Email != nil && spec.SignIn.Email.GetEnabled()),
		}
		if spec.SignIn.Email != nil {
			emailArgs.PasswordRequired = pulumi.BoolPtr(spec.SignIn.Email.PasswordRequired)
		}
		signInArgs.Email = emailArgs
		phoneArgs := &identityplatform.ConfigSignInPhoneNumberArgs{
			Enabled: pulumi.Bool(spec.SignIn.PhoneNumber != nil && spec.SignIn.PhoneNumber.GetEnabled()),
		}
		if spec.SignIn.PhoneNumber != nil && len(spec.SignIn.PhoneNumber.TestPhoneNumbers) > 0 {
			phoneArgs.TestPhoneNumbers = pulumi.ToStringMap(spec.SignIn.PhoneNumber.TestPhoneNumbers)
		}
		signInArgs.PhoneNumber = phoneArgs
		if spec.SignIn.Anonymous != nil {
			signInArgs.Anonymous = &identityplatform.ConfigSignInAnonymousArgs{
				// Explicit send — see the function comment.
				Enabled: pulumi.Bool(spec.SignIn.Anonymous.GetEnabled()),
			}
		}
		args.SignIn = signInArgs
	}

	if spec.Mfa != nil {
		mfaArgs := &identityplatform.ConfigMfaArgs{}
		if spec.Mfa.State != "" {
			mfaArgs.State = pulumi.StringPtr(spec.Mfa.State)
		}
		if len(spec.Mfa.EnabledProviders) > 0 {
			mfaArgs.EnabledProviders = pulumi.ToStringArray(spec.Mfa.EnabledProviders)
		}
		if len(spec.Mfa.ProviderConfigs) > 0 {
			providerConfigs := identityplatform.ConfigMfaProviderConfigArray{}
			for _, providerConfig := range spec.Mfa.ProviderConfigs {
				providerConfigArgs := &identityplatform.ConfigMfaProviderConfigArgs{}
				if providerConfig.State != "" {
					providerConfigArgs.State = pulumi.StringPtr(providerConfig.State)
				}
				if providerConfig.TotpProviderConfig != nil {
					providerConfigArgs.TotpProviderConfig = &identityplatform.ConfigMfaProviderConfigTotpProviderConfigArgs{
						AdjacentIntervals: pulumi.IntPtr(int(providerConfig.TotpProviderConfig.AdjacentIntervals)),
					}
				}
				providerConfigs = append(providerConfigs, providerConfigArgs)
			}
			mfaArgs.ProviderConfigs = providerConfigs
		}
		args.Mfa = mfaArgs
	}

	if spec.BlockingFunctions != nil {
		triggers := identityplatform.ConfigBlockingFunctionsTriggerArray{}
		for _, trigger := range spec.BlockingFunctions.Triggers {
			triggers = append(triggers, &identityplatform.ConfigBlockingFunctionsTriggerArgs{
				EventType:   pulumi.String(trigger.EventType),
				FunctionUri: pulumi.String(trigger.FunctionUri.GetValue()),
			})
		}
		blockingArgs := &identityplatform.ConfigBlockingFunctionsArgs{
			Triggers: triggers,
		}
		if spec.BlockingFunctions.ForwardInboundCredentials != nil {
			blockingArgs.ForwardInboundCredentials = &identityplatform.ConfigBlockingFunctionsForwardInboundCredentialsArgs{
				AccessToken:  pulumi.BoolPtr(spec.BlockingFunctions.ForwardInboundCredentials.AccessToken),
				IdToken:      pulumi.BoolPtr(spec.BlockingFunctions.ForwardInboundCredentials.IdToken),
				RefreshToken: pulumi.BoolPtr(spec.BlockingFunctions.ForwardInboundCredentials.RefreshToken),
			}
		}
		args.BlockingFunctions = blockingArgs
	}

	if spec.SignUpQuota != nil {
		args.Quota = &identityplatform.ConfigQuotaArgs{
			SignUpQuotaConfig: &identityplatform.ConfigQuotaSignUpQuotaConfigArgs{
				Quota:         pulumi.IntPtr(int(spec.SignUpQuota.Quota)),
				QuotaDuration: pulumi.StringPtr(spec.SignUpQuota.QuotaDuration),
				StartTime:     pulumi.StringPtr(spec.SignUpQuota.StartTime),
			},
		}
	}

	if spec.SmsRegionConfig != nil {
		smsArgs := &identityplatform.ConfigSmsRegionConfigArgs{}
		if spec.SmsRegionConfig.AllowByDefault != nil {
			allowByDefaultArgs := &identityplatform.ConfigSmsRegionConfigAllowByDefaultArgs{}
			if len(spec.SmsRegionConfig.AllowByDefault.DisallowedRegions) > 0 {
				allowByDefaultArgs.DisallowedRegions = pulumi.ToStringArray(spec.SmsRegionConfig.AllowByDefault.DisallowedRegions)
			}
			smsArgs.AllowByDefault = allowByDefaultArgs
		}
		if spec.SmsRegionConfig.AllowlistOnly != nil {
			allowlistOnlyArgs := &identityplatform.ConfigSmsRegionConfigAllowlistOnlyArgs{}
			if len(spec.SmsRegionConfig.AllowlistOnly.AllowedRegions) > 0 {
				allowlistOnlyArgs.AllowedRegions = pulumi.ToStringArray(spec.SmsRegionConfig.AllowlistOnly.AllowedRegions)
			}
			smsArgs.AllowlistOnly = allowlistOnlyArgs
		}
		args.SmsRegionConfig = smsArgs
	}

	if spec.ClientPermissions != nil {
		args.Client = &identityplatform.ConfigClientArgs{
			Permissions: &identityplatform.ConfigClientPermissionsArgs{
				DisabledUserSignup:   pulumi.BoolPtr(spec.ClientPermissions.DisabledUserSignup),
				DisabledUserDeletion: pulumi.BoolPtr(spec.ClientPermissions.DisabledUserDeletion),
			},
		}
	}

	if spec.RequestLoggingEnabled != nil {
		args.Monitoring = &identityplatform.ConfigMonitoringArgs{
			RequestLogging: &identityplatform.ConfigMonitoringRequestLoggingArgs{
				// Explicit send whenever set — true AND false must reach
				// the API (the send-true-or-omit class).
				Enabled: pulumi.Bool(spec.GetRequestLoggingEnabled()),
			},
		}
	}

	if spec.MultiTenant != nil {
		multiTenantArgs := &identityplatform.ConfigMultiTenantArgs{
			AllowTenants: pulumi.BoolPtr(spec.MultiTenant.AllowTenants),
		}
		if spec.MultiTenant.DefaultTenantLocation != "" {
			multiTenantArgs.DefaultTenantLocation = pulumi.StringPtr(spec.MultiTenant.DefaultTenantLocation)
		}
		args.MultiTenant = multiTenantArgs
	}

	configOpts := []pulumi.ResourceOption{
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}),
	}
	// Adoption: initialization is one-way and ONCE-ONLY — GCP rejects a
	// second initializeAuth with 400 "Identity Platform has already been
	// enabled for this project" (live-verified). When the spec arms
	// adopt_existing, the deterministic singleton (projects/{project}/config)
	// is imported instead of created; the option is consulted only while the
	// resource is not yet in state, so re-applies stay clean.
	if spec.AdoptExisting {
		adoptionProject := spec.ProjectId.GetValue()
		if adoptionProject == "" {
			// Ambient-project fallback for the import ID only — mirrors the
			// Terraform module's count-gated client-config read.
			clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
			if err != nil {
				return errors.Wrap(err, "failed to resolve ambient project for identity platform adoption")
			}
			adoptionProject = clientConfig.Project
		}
		configOpts = append(configOpts, pulumi.Import(pulumi.ID(fmt.Sprintf("projects/%s/config", adoptionProject))))
	}

	createdConfig, err := identityplatform.NewConfig(ctx, "config", args, configOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create identity platform config")
	}

	ctx.Export(OpConfigName, createdConfig.Name)
	// The client block is computed by GCP whether or not permissions were
	// configured; normalize absent values to "" for a stable output shape.
	ctx.Export(OpApiKey, createdConfig.Client.ApiKey().ApplyT(func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}))
	ctx.Export(OpFirebaseSubdomain, createdConfig.Client.FirebaseSubdomain().ApplyT(func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}))

	// The composed project-level IdP configs. Each depends on the config
	// so it lands only after the project is initialized, and each carries
	// the spec's deletion_policy (the config itself cannot be deleted).
	for _, idp := range spec.DefaultSupportedIdps {
		idpArgs := &identityplatform.DefaultSupportedIdpConfigArgs{
			IdpId:        pulumi.String(idp.IdpId),
			ClientId:     pulumi.String(idp.ClientId),
			ClientSecret: pulumi.String(idp.ClientSecret),
			// Explicit send — a disabled IdP must reach the API.
			Enabled: pulumi.Bool(idp.Enabled == nil || idp.GetEnabled()),
		}
		if spec.DeletionPolicy != "" {
			idpArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			idpArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if _, err := identityplatform.NewDefaultSupportedIdpConfig(ctx,
			"default-idp-"+idp.IdpId, idpArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdConfig})); err != nil {
			return errors.Wrapf(err, "failed to create default supported idp config %s", idp.IdpId)
		}
	}

	for _, oidc := range spec.OauthIdpConfigs {
		oidcArgs := &identityplatform.OauthIdpConfigArgs{
			Name:     pulumi.String(oidc.Name),
			Issuer:   pulumi.String(oidc.Issuer),
			ClientId: pulumi.String(oidc.ClientId),
			// Explicit send — a disabled IdP must reach the API.
			Enabled: pulumi.Bool(oidc.Enabled == nil || oidc.GetEnabled()),
		}
		if oidc.ClientSecret != "" {
			oidcArgs.ClientSecret = pulumi.StringPtr(oidc.ClientSecret)
		}
		if oidc.DisplayName != "" {
			oidcArgs.DisplayName = pulumi.StringPtr(oidc.DisplayName)
		}
		if oidc.ResponseType != nil {
			oidcArgs.ResponseType = &identityplatform.OauthIdpConfigResponseTypeArgs{
				Code:    pulumi.BoolPtr(oidc.ResponseType.Code),
				IdToken: pulumi.BoolPtr(oidc.ResponseType.IdToken),
			}
		}
		if spec.DeletionPolicy != "" {
			oidcArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			oidcArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if _, err := identityplatform.NewOauthIdpConfig(ctx,
			"oauth-idp-"+oidc.Name, oidcArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdConfig})); err != nil {
			return errors.Wrapf(err, "failed to create oauth idp config %s", oidc.Name)
		}
	}

	for _, saml := range spec.InboundSamlConfigs {
		idpCertificates := identityplatform.InboundSamlConfigIdpConfigIdpCertificateArray{}
		for _, certificate := range saml.IdpConfig.IdpCertificates {
			idpCertificates = append(idpCertificates, &identityplatform.InboundSamlConfigIdpConfigIdpCertificateArgs{
				X509Certificate: pulumi.StringPtr(certificate.X509Certificate),
			})
		}
		samlArgs := &identityplatform.InboundSamlConfigArgs{
			Name:        pulumi.String(saml.Name),
			DisplayName: pulumi.String(saml.DisplayName),
			// Explicit send — a disabled IdP must reach the API.
			Enabled: pulumi.Bool(saml.Enabled == nil || saml.GetEnabled()),
			IdpConfig: &identityplatform.InboundSamlConfigIdpConfigArgs{
				IdpEntityId:     pulumi.String(saml.IdpConfig.IdpEntityId),
				SsoUrl:          pulumi.String(saml.IdpConfig.SsoUrl),
				SignRequest:     pulumi.BoolPtr(saml.IdpConfig.SignRequest),
				IdpCertificates: idpCertificates,
			},
		}
		if saml.SpConfig != nil {
			spArgs := &identityplatform.InboundSamlConfigSpConfigArgs{}
			if saml.SpConfig.CallbackUri != "" {
				spArgs.CallbackUri = pulumi.StringPtr(saml.SpConfig.CallbackUri)
			}
			if saml.SpConfig.SpEntityId != "" {
				spArgs.SpEntityId = pulumi.StringPtr(saml.SpConfig.SpEntityId)
			}
			samlArgs.SpConfig = spArgs
		}
		if spec.DeletionPolicy != "" {
			samlArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			samlArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if _, err := identityplatform.NewInboundSamlConfig(ctx,
			"saml-idp-"+saml.Name, samlArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdConfig})); err != nil {
			return errors.Wrapf(err, "failed to create inbound saml config %s", saml.Name)
		}
	}

	return nil
}
