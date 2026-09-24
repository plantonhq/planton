package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firebase"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// androidApp registers the Android app in its Firebase project and
// composes the app's App Check surface: the Play Integrity attestation
// configuration and the debug tokens.
//
// The registration is the resource of substance. Its package name is
// immutable (Firebase's identity for the app; a change replaces the
// registration), and deletion_policy DELETE removes it IMMEDIATELY and
// PERMANENTLY -- the provider posts :remove with immediate=true, skipping
// Firebase's 30-day recoverable window. The spec's deletion_policy governs
// the app and every debug token; the Play Integrity configuration has no
// delete on Google's side (a per-app singleton the provider only forgets),
// so it carries none.
//
// API enablement is module plumbing: firebaseappcheck.googleapis.com is
// enabled exactly when the spec composes an App Check resource, and never
// disabled on destroy -- the project's other apps may depend on it. The
// Firebase Management API itself is enabled by the GcpFirebaseProject the
// app lives in.
func androidApp(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirebaseAndroidApp.Spec
	projectId := spec.ProjectId.GetValue()

	appArgs := &firebase.AndroidAppArgs{
		DisplayName: pulumi.String(spec.DisplayName),
		PackageName: pulumi.String(spec.PackageName),
	}
	if projectId != "" {
		appArgs.Project = pulumi.StringPtr(projectId)
	}
	// Certificate fingerprints are sent only when declared: the provider's
	// lists are Optional and an absent list means "none registered", which
	// is exactly what an empty spec list means too.
	if len(spec.Sha1Hashes) > 0 {
		appArgs.Sha1Hashes = pulumi.ToStringArray(spec.Sha1Hashes)
	}
	if len(spec.Sha256Hashes) > 0 {
		appArgs.Sha256Hashes = pulumi.ToStringArray(spec.Sha256Hashes)
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
	createdApp, err := firebase.NewAndroidApp(ctx, "app", appArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to register the android app")
	}

	// App Check. Play Integrity counts as configured when its block is
	// present and not declared off (enabled defaults to true -- the wire
	// has no switch, the configuration's existence is the switch).
	wantsPlayIntegrity := spec.AppCheck != nil && spec.AppCheck.PlayIntegrity != nil &&
		(spec.AppCheck.PlayIntegrity.Enabled == nil || spec.AppCheck.PlayIntegrity.GetEnabled())
	wantsDebugTokens := spec.AppCheck != nil && len(spec.AppCheck.DebugTokens) > 0

	if wantsPlayIntegrity || wantsDebugTokens {
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

		if wantsPlayIntegrity {
			args := &firebase.AppCheckPlayIntegrityConfigArgs{
				AppId: createdApp.AppId,
			}
			// token_ttl is Optional+Computed: unset lets Google assume its
			// 1-hour default and read it back, so it is sent only when set.
			if spec.AppCheck.PlayIntegrity.TokenTtl != "" {
				args.TokenTtl = pulumi.StringPtr(spec.AppCheck.PlayIntegrity.TokenTtl)
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if _, err := firebase.NewAppCheckPlayIntegrityConfig(ctx, "app-check-play-integrity", args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrap(err, "failed to configure play integrity attestation")
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

	// The app's configuration file (google-services.json), read AFTER the
	// registration: the lookup takes the created app's id, so the invoke
	// waits on it. A build input that ships in the APK, exported plain.
	configArgs := firebase.GetAndroidAppConfigOutputArgs{
		AppId: createdApp.AppId,
	}
	if projectId != "" {
		configArgs.Project = pulumi.StringPtr(projectId)
	}
	config := firebase.GetAndroidAppConfigOutput(ctx, configArgs, pulumi.Provider(gcpProvider))

	ctx.Export(OpAppId, createdApp.AppId)
	ctx.Export(OpName, createdApp.Name)
	ctx.Export(OpApiKeyId, createdApp.ApiKeyId)
	ctx.Export(OpConfigFilename, config.ConfigFilename())
	ctx.Export(OpConfigFileContents, config.ConfigFileContents())

	return nil
}
