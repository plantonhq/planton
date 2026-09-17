package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firebase"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// webApp registers the web app in its Firebase project and composes the
// app's App Check surface: the reCAPTCHA v3 and reCAPTCHA Enterprise
// attestation configurations and the debug tokens.
//
// The registration is the resource of substance. A web app has no identity
// beyond its display name, so nothing on it forces replacement; deletion_
// policy DELETE removes it IMMEDIATELY and PERMANENTLY -- the provider
// posts :remove with immediate=true, skipping Firebase's 30-day recoverable
// window. The spec's deletion_policy governs the app and every debug token;
// the attestation configurations have no delete on Google's side (per-app
// singletons the provider only forgets), so they carry none.
//
// Both reCAPTCHA configurations may be set at once: Google keeps them
// independent on the backend, and that is how a site migrates from v3 to
// Enterprise without a gap (a client initialises with one provider).
//
// API enablement is module plumbing: firebaseappcheck.googleapis.com is
// enabled exactly when the spec composes an App Check resource, and never
// disabled on destroy -- the project's other apps may depend on it. The
// Firebase Management API itself is enabled by the GcpFirebaseProject the
// app lives in.
func webApp(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirebaseWebApp.Spec
	projectId := spec.ProjectId.GetValue()

	appArgs := &firebase.WebAppArgs{
		DisplayName: pulumi.String(spec.DisplayName),
	}
	if projectId != "" {
		appArgs.Project = pulumi.StringPtr(projectId)
	}
	// api_key_id is Optional+Computed on the provider: when the spec names
	// no key, Firebase associates or provisions one and the computed value
	// is read back, so it is sent only when the spec sets it.
	if spec.ApiKeyId.GetValue() != "" {
		appArgs.ApiKeyId = pulumi.StringPtr(spec.ApiKeyId.GetValue())
	}
	if spec.DeletionPolicy != "" {
		appArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	createdApp, err := firebase.NewWebApp(ctx, "app", appArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to register the web app")
	}

	// App Check. Both reCAPTCHA blocks carry required content, so their
	// presence is their switch.
	wantsRecaptchaV3 := spec.AppCheck != nil && spec.AppCheck.RecaptchaV3 != nil
	wantsRecaptchaEnterprise := spec.AppCheck != nil && spec.AppCheck.RecaptchaEnterprise != nil
	wantsDebugTokens := spec.AppCheck != nil && len(spec.AppCheck.DebugTokens) > 0

	if wantsRecaptchaV3 || wantsRecaptchaEnterprise || wantsDebugTokens {
		appCheckApiArgs := &projects.ServiceArgs{
			Service:                  pulumi.String("firebaseappcheck.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
		}
		if projectId != "" {
			appCheckApiArgs.Project = pulumi.String(projectId)
		}
		appCheckApi, err := projects.NewService(ctx, "firebaseappcheck-api", appCheckApiArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to enable the app check api")
		}
		// Every App Check resource waits on the registration AND the API:
		// the configurations are addressed by the app id the registration
		// returns, and the App Check service must be reachable.
		deps := pulumi.DependsOn([]pulumi.Resource{createdApp, appCheckApi})

		if wantsRecaptchaV3 {
			v3 := spec.AppCheck.RecaptchaV3
			args := &firebase.AppCheckRecaptchaV3ConfigArgs{
				AppId: createdApp.AppId,
				// The site secret is a secret Google never returns: ToSecret
				// keeps it encrypted in state, and the SDK marks the resource's
				// siteSecret output secret on its own.
				SiteSecret: pulumi.ToSecret(pulumi.String(v3.SiteSecret)).(pulumi.StringOutput),
			}
			// token_ttl is Optional+Computed: unset lets Google assume its
			// 1-hour default and read it back, so it is sent only when set.
			if v3.TokenTtl != "" {
				args.TokenTtl = pulumi.StringPtr(v3.TokenTtl)
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if _, err := firebase.NewAppCheckRecaptchaV3Config(ctx, "app-check-recaptcha-v3", args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrap(err, "failed to configure recaptcha v3 attestation")
			}
		}

		if wantsRecaptchaEnterprise {
			ent := spec.AppCheck.RecaptchaEnterprise
			args := &firebase.AppCheckRecaptchaEnterpriseConfigArgs{
				AppId: createdApp.AppId,
				// The site key is the PUBLIC half of the reCAPTCHA key -- the
				// value the page embeds -- so it is sent plain.
				SiteKey: pulumi.String(ent.SiteKey),
			}
			if ent.TokenTtl != "" {
				args.TokenTtl = pulumi.StringPtr(ent.TokenTtl)
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if _, err := firebase.NewAppCheckRecaptchaEnterpriseConfig(ctx, "app-check-recaptcha-enterprise", args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrap(err, "failed to configure recaptcha enterprise attestation")
			}
		}

		// Debug tokens are keyed by their display name (unique within the
		// app by validation) so plans stay stable as list order changes. The
		// token value is a secret: ToSecret keeps it encrypted in state, and
		// the SDK marks the resource's token output secret on its own.
		for _, dt := range spec.AppCheck.DebugTokens {
			args := &firebase.AppCheckDebugTokenArgs{
				AppId:       createdApp.AppId,
				DisplayName: pulumi.String(dt.DisplayName),
				Token:       pulumi.ToSecret(pulumi.String(dt.Token)).(pulumi.StringOutput),
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if spec.DeletionPolicy != "" {
				args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
			}
			if _, err := firebase.NewAppCheckDebugToken(ctx, "debug-token-"+dt.DisplayName, args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrapf(err, "failed to create debug token %s", dt.DisplayName)
			}
		}
	}

	// The app's firebaseConfig, read AFTER the registration: the lookup
	// takes the created app's id (the web config lookup names its input
	// web_app_id -- the one naming divergence in the family), so the invoke
	// waits on it. Every value is a client identifier that ships in the
	// page, exported plain; the conditionally present ones (RTDB, default
	// bucket, finalized location, linked Analytics) degrade to "".
	configArgs := firebase.GetWebAppConfigOutputArgs{
		WebAppId: createdApp.AppId,
	}
	if projectId != "" {
		configArgs.Project = pulumi.StringPtr(projectId)
	}
	config := firebase.GetWebAppConfigOutput(ctx, configArgs, pulumi.Provider(gcpProvider))

	ctx.Export(OpAppId, createdApp.AppId)
	ctx.Export(OpName, createdApp.Name)
	ctx.Export(OpApiKeyId, createdApp.ApiKeyId)
	ctx.Export(OpAppUrls, createdApp.AppUrls)
	ctx.Export(OpApiKey, config.ApiKey())
	ctx.Export(OpAuthDomain, config.AuthDomain())
	ctx.Export(OpDatabaseUrl, config.DatabaseUrl())
	ctx.Export(OpStorageBucket, config.StorageBucket())
	ctx.Export(OpLocationId, config.LocationId())
	ctx.Export(OpMessagingSenderId, config.MessagingSenderId())
	ctx.Export(OpMeasurementId, config.MeasurementId())

	return nil
}
