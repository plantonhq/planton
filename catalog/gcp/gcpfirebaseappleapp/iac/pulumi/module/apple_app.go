package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firebase"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// appleApp registers the iOS / macOS app in its Firebase project and
// composes the app's App Check surface: the App Attest and DeviceCheck
// attestation configurations and the debug tokens.
//
// The registration is the resource of substance. Its bundle id is
// immutable (Firebase's identity for the app; a change replaces the
// registration), and deletion_policy DELETE removes it IMMEDIATELY and
// PERMANENTLY -- the provider posts :remove with immediate=true, skipping
// Firebase's 30-day recoverable window. The spec's deletion_policy governs
// the app and every debug token; the attestation configurations have no
// delete on Google's side (per-app singletons the provider only forgets),
// so they carry none.
//
// API enablement is module plumbing: firebaseappcheck.googleapis.com is
// enabled exactly when the spec composes an App Check resource, and never
// disabled on destroy -- the project's other apps may depend on it. The
// Firebase Management API itself is enabled by the GcpFirebaseProject the
// app lives in.
//
// What this module does NOT do, by Google's design: upload the APNs
// authentication key push needs on Apple platforms. Firebase exposes no
// API for it; it is a console step, recorded where the estate keeps its
// vendor procedures.
func appleApp(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirebaseAppleApp.Spec
	projectId := spec.ProjectId.GetValue()

	appArgs := &firebase.AppleAppArgs{
		DisplayName: pulumi.String(spec.DisplayName),
		BundleId:    pulumi.String(spec.BundleId),
	}
	if projectId != "" {
		appArgs.Project = pulumi.StringPtr(projectId)
	}
	// app_store_id and team_id are plain Optional on the provider (not
	// Computed): sent only when set, so an unset field never diffs.
	if spec.AppStoreId != "" {
		appArgs.AppStoreId = pulumi.StringPtr(spec.AppStoreId)
	}
	if spec.TeamId != "" {
		appArgs.TeamId = pulumi.StringPtr(spec.TeamId)
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
	createdApp, err := firebase.NewAppleApp(ctx, "app", appArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to register the apple app")
	}

	// App Check. App Attest counts as configured when its block is present
	// and not declared off (enabled defaults to true -- the wire has no
	// switch, the configuration's existence is the switch). DeviceCheck's
	// block carries required content, so its presence is its switch.
	wantsAppAttest := spec.AppCheck != nil && spec.AppCheck.AppAttest != nil &&
		(spec.AppCheck.AppAttest.Enabled == nil || spec.AppCheck.AppAttest.GetEnabled())
	wantsDeviceCheck := spec.AppCheck != nil && spec.AppCheck.DeviceCheck != nil
	wantsDebugTokens := spec.AppCheck != nil && len(spec.AppCheck.DebugTokens) > 0

	if wantsAppAttest || wantsDeviceCheck || wantsDebugTokens {
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

		if wantsAppAttest {
			args := &firebase.AppCheckAppAttestConfigArgs{
				AppId: createdApp.AppId,
			}
			// token_ttl is Optional+Computed: unset lets Google assume its
			// 1-hour default and read it back, so it is sent only when set.
			if spec.AppCheck.AppAttest.TokenTtl != "" {
				args.TokenTtl = pulumi.StringPtr(spec.AppCheck.AppAttest.TokenTtl)
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if _, err := firebase.NewAppCheckAppAttestConfig(ctx, "app-check-app-attest", args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrap(err, "failed to configure app attest attestation")
			}
		}

		if wantsDeviceCheck {
			dc := spec.AppCheck.DeviceCheck
			args := &firebase.AppCheckDeviceCheckConfigArgs{
				AppId: createdApp.AppId,
				KeyId: pulumi.String(dc.KeyId),
				// The DeviceCheck private key is a secret Google never returns:
				// ToSecret keeps it encrypted in state, and the SDK marks the
				// resource's privateKey output secret on its own.
				PrivateKey: pulumi.ToSecret(pulumi.String(dc.PrivateKey)).(pulumi.StringOutput),
			}
			if dc.TokenTtl != "" {
				args.TokenTtl = pulumi.StringPtr(dc.TokenTtl)
			}
			if projectId != "" {
				args.Project = pulumi.StringPtr(projectId)
			}
			if _, err := firebase.NewAppCheckDeviceCheckConfig(ctx, "app-check-device-check", args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrap(err, "failed to configure devicecheck attestation")
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

	// The app's configuration file (GoogleService-Info.plist), read AFTER
	// the registration: the lookup takes the created app's id, so the
	// invoke waits on it. A build input that ships in the bundle, exported
	// plain.
	configArgs := firebase.GetAppleAppConfigOutputArgs{
		AppId: createdApp.AppId,
	}
	if projectId != "" {
		configArgs.Project = pulumi.StringPtr(projectId)
	}
	config := firebase.GetAppleAppConfigOutput(ctx, configArgs, pulumi.Provider(gcpProvider))

	ctx.Export(OpAppId, createdApp.AppId)
	ctx.Export(OpName, createdApp.Name)
	ctx.Export(OpApiKeyId, createdApp.ApiKeyId)
	ctx.Export(OpConfigFilename, config.ConfigFilename())
	ctx.Export(OpConfigFileContents, config.ConfigFileContents())

	return nil
}
