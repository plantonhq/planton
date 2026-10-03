package module

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	cloudflarezerotrustgatewaysettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaysettings/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nonSlugCharacters matches everything a PAC-file slug may not carry; the
// tofu module applies the identical substitution (cross-engine parity).
var nonSlugCharacters = regexp.MustCompile(`[^A-Za-z0-9-]`)

// slugFromName derives the deterministic PAC-file slug the modules send
// when a row leaves slug unset (see the fan-out comment below for why an
// omitted slug cannot be left to the server).
func slugFromName(name string) string {
	return strings.ToLower(nonSlugCharacters.ReplaceAllString(name, "-"))
}

// gatewaySettings applies the three folded Gateway surfaces:
//
//   - The account configuration SINGLETON (settings): create and update are
//     the same PUT; DESTROY IS A NO-OP that abandons the live configuration.
//     An unset spec sub-object is never sent, leaving the live value
//     (dashboard-set or default) untouched.
//   - The logging SINGLETON: same lifecycle. Unlike settings, the module
//     sends the COMPLETE logging tree when the spec declares it -- omitted
//     logging fields drift at Cloudflare (the provider's own tests accept
//     non-empty plans on partial sends), so explicit emission is what keeps
//     apply idempotent.
//   - The PAC-file COLLECTION: one provider resource per spec row, each with
//     a real create/update/delete lifecycle.
func gatewaySettings(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustGatewaySettings.Spec

	if spec.Settings != nil {
		args := &cloudflare.ZeroTrustGatewaySettingsArgs{
			AccountId: pulumi.String(spec.AccountId),
			Settings:  buildSettings(spec.Settings),
		}
		if _, err := cloudflare.NewZeroTrustGatewaySettings(
			ctx,
			"gateway_settings",
			args,
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to apply gateway configuration")
		}
	}

	if spec.Logging != nil {
		if _, err := cloudflare.NewZeroTrustGatewayLogging(
			ctx,
			"gateway_logging",
			buildLogging(spec.AccountId, spec.Logging),
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to apply gateway logging")
		}
	}

	// PAC files fan out one provider resource per row, keyed by name so a
	// row edit replaces only its own file. The server-assigned file ids are
	// collected keyed by that same name for the pacfile_ids output.
	//
	// The slug is ALWAYS sent, derived deterministically from the row's
	// name when unset: an omitted slug gets a RANDOM server-generated one
	// that the refresh echoes into state, and since slug is
	// replace-forcing, every subsequent plan proposes destroying and
	// recreating the file -- breaking its public URL (live-measured at
	// provider v5.23.0). A deterministic name-derived slug keeps config and
	// state equal forever.
	pacfileIds := pulumi.StringMap{}
	for _, pacFile := range spec.PacFiles {
		pacArgs := &cloudflare.ZeroTrustGatewayPacfileArgs{
			AccountId: pulumi.String(spec.AccountId),
			Name:      pulumi.String(pacFile.Name),
			Contents:  pulumi.String(pacFile.Contents),
			Slug:      pulumi.String(slugFromName(pacFile.Name)),
		}
		if pacFile.Slug != "" {
			pacArgs.Slug = pulumi.String(pacFile.Slug)
		}
		if pacFile.Description != "" {
			pacArgs.Description = pulumi.String(pacFile.Description)
		}
		createdPacFile, err := cloudflare.NewZeroTrustGatewayPacfile(
			ctx,
			fmt.Sprintf("pac_file_%s", pacFile.Name),
			pacArgs,
			pulumi.Provider(cloudflareProvider),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create pac file %s", pacFile.Name)
		}
		pacfileIds[pacFile.Name] = createdPacFile.ID()
	}

	ctx.Export(OpAccountId, pulumi.String(spec.AccountId))
	ctx.Export(OpPacfileIds, pacfileIds)

	return nil
}

// buildSettings maps only the sub-objects the spec declares -- an unset
// sub-object is NOT MANAGED and never sent.
func buildSettings(settings *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewayConfig) cloudflare.ZeroTrustGatewaySettingsSettingsPtrInput {
	args := cloudflare.ZeroTrustGatewaySettingsSettingsArgs{}

	if settings.ActivityLog != nil && settings.ActivityLog.Enabled != nil {
		args.ActivityLog = cloudflare.ZeroTrustGatewaySettingsSettingsActivityLogArgs{
			Enabled: pulumi.BoolPtr(settings.ActivityLog.GetEnabled()),
		}
	}

	if settings.Antivirus != nil {
		antivirus := cloudflare.ZeroTrustGatewaySettingsSettingsAntivirusArgs{}
		if settings.Antivirus.EnabledDownloadPhase != nil {
			antivirus.EnabledDownloadPhase = pulumi.BoolPtr(settings.Antivirus.GetEnabledDownloadPhase())
		}
		if settings.Antivirus.EnabledUploadPhase != nil {
			antivirus.EnabledUploadPhase = pulumi.BoolPtr(settings.Antivirus.GetEnabledUploadPhase())
		}
		if settings.Antivirus.FailClosed != nil {
			antivirus.FailClosed = pulumi.BoolPtr(settings.Antivirus.GetFailClosed())
		}
		if settings.Antivirus.NotificationSettings != nil {
			notification := cloudflare.ZeroTrustGatewaySettingsSettingsAntivirusNotificationSettingsArgs{}
			if settings.Antivirus.NotificationSettings.Enabled != nil {
				notification.Enabled = pulumi.BoolPtr(settings.Antivirus.NotificationSettings.GetEnabled())
			}
			if settings.Antivirus.NotificationSettings.Msg != "" {
				notification.Msg = pulumi.String(settings.Antivirus.NotificationSettings.Msg)
			}
			if settings.Antivirus.NotificationSettings.SupportUrl != "" {
				notification.SupportUrl = pulumi.String(settings.Antivirus.NotificationSettings.SupportUrl)
			}
			if settings.Antivirus.NotificationSettings.IncludeContext != nil {
				notification.IncludeContext = pulumi.BoolPtr(settings.Antivirus.NotificationSettings.GetIncludeContext())
			}
			antivirus.NotificationSettings = notification
		}
		args.Antivirus = antivirus
	}

	if settings.BlockPage != nil {
		blockPage := cloudflare.ZeroTrustGatewaySettingsSettingsBlockPageArgs{}
		if settings.BlockPage.Enabled != nil {
			blockPage.Enabled = pulumi.BoolPtr(settings.BlockPage.GetEnabled())
		}
		if settings.BlockPage.Mode != "" {
			blockPage.Mode = pulumi.String(settings.BlockPage.Mode)
		}
		if settings.BlockPage.TargetUri != "" {
			blockPage.TargetUri = pulumi.String(settings.BlockPage.TargetUri)
		}
		if settings.BlockPage.IncludeContext != nil {
			blockPage.IncludeContext = pulumi.BoolPtr(settings.BlockPage.GetIncludeContext())
		}
		if settings.BlockPage.Name != "" {
			blockPage.Name = pulumi.String(settings.BlockPage.Name)
		}
		if settings.BlockPage.HeaderText != "" {
			blockPage.HeaderText = pulumi.String(settings.BlockPage.HeaderText)
		}
		if settings.BlockPage.FooterText != "" {
			blockPage.FooterText = pulumi.String(settings.BlockPage.FooterText)
		}
		if settings.BlockPage.SuppressFooter != nil {
			blockPage.SuppressFooter = pulumi.BoolPtr(settings.BlockPage.GetSuppressFooter())
		}
		if settings.BlockPage.BackgroundColor != "" {
			blockPage.BackgroundColor = pulumi.String(settings.BlockPage.BackgroundColor)
		}
		if settings.BlockPage.LogoPath != "" {
			blockPage.LogoPath = pulumi.String(settings.BlockPage.LogoPath)
		}
		if settings.BlockPage.MailtoAddress != "" {
			blockPage.MailtoAddress = pulumi.String(settings.BlockPage.MailtoAddress)
		}
		if settings.BlockPage.MailtoSubject != "" {
			blockPage.MailtoSubject = pulumi.String(settings.BlockPage.MailtoSubject)
		}
		args.BlockPage = blockPage
	}

	if settings.BodyScanning != nil && settings.BodyScanning.InspectionMode != "" {
		args.BodyScanning = cloudflare.ZeroTrustGatewaySettingsSettingsBodyScanningArgs{
			InspectionMode: pulumi.String(settings.BodyScanning.InspectionMode),
		}
	}

	if settings.BrowserIsolation != nil {
		isolation := cloudflare.ZeroTrustGatewaySettingsSettingsBrowserIsolationArgs{}
		if settings.BrowserIsolation.NonIdentityEnabled != nil {
			isolation.NonIdentityEnabled = pulumi.BoolPtr(settings.BrowserIsolation.GetNonIdentityEnabled())
		}
		if settings.BrowserIsolation.UrlBrowserIsolationEnabled != nil {
			isolation.UrlBrowserIsolationEnabled = pulumi.BoolPtr(settings.BrowserIsolation.GetUrlBrowserIsolationEnabled())
		}
		args.BrowserIsolation = isolation
	}

	if settings.Certificate != nil {
		args.Certificate = cloudflare.ZeroTrustGatewaySettingsSettingsCertificateArgs{
			Id: pulumi.String(settings.Certificate.Id.GetValue()),
		}
	}

	if settings.ExtendedEmailMatching != nil && settings.ExtendedEmailMatching.Enabled != nil {
		args.ExtendedEmailMatching = cloudflare.ZeroTrustGatewaySettingsSettingsExtendedEmailMatchingArgs{
			Enabled: pulumi.BoolPtr(settings.ExtendedEmailMatching.GetEnabled()),
		}
	}

	if settings.Fips != nil && settings.Fips.Tls != nil {
		args.Fips = cloudflare.ZeroTrustGatewaySettingsSettingsFipsArgs{
			Tls: pulumi.BoolPtr(settings.Fips.GetTls()),
		}
	}

	if settings.HostSelector != nil && settings.HostSelector.Enabled != nil {
		args.HostSelector = cloudflare.ZeroTrustGatewaySettingsSettingsHostSelectorArgs{
			Enabled: pulumi.BoolPtr(settings.HostSelector.GetEnabled()),
		}
	}

	if settings.Inspection != nil && settings.Inspection.Mode != "" {
		args.Inspection = cloudflare.ZeroTrustGatewaySettingsSettingsInspectionArgs{
			Mode: pulumi.String(settings.Inspection.Mode),
		}
	}

	if settings.MaxTtlSecs != nil {
		args.MaxTtlSecs = pulumi.IntPtr(int(settings.GetMaxTtlSecs()))
	}

	if settings.ProtocolDetection != nil && settings.ProtocolDetection.Enabled != nil {
		args.ProtocolDetection = cloudflare.ZeroTrustGatewaySettingsSettingsProtocolDetectionArgs{
			Enabled: pulumi.BoolPtr(settings.ProtocolDetection.GetEnabled()),
		}
	}

	if settings.Sandbox != nil {
		sandbox := cloudflare.ZeroTrustGatewaySettingsSettingsSandboxArgs{}
		if settings.Sandbox.Enabled != nil {
			sandbox.Enabled = pulumi.BoolPtr(settings.Sandbox.GetEnabled())
		}
		if settings.Sandbox.FallbackAction != "" {
			sandbox.FallbackAction = pulumi.String(settings.Sandbox.FallbackAction)
		}
		args.Sandbox = sandbox
	}

	if settings.TlsDecrypt != nil && settings.TlsDecrypt.Enabled != nil {
		args.TlsDecrypt = cloudflare.ZeroTrustGatewaySettingsSettingsTlsDecryptArgs{
			Enabled: pulumi.BoolPtr(settings.TlsDecrypt.GetEnabled()),
		}
	}

	return args
}

// buildLogging emits the COMPLETE logging tree: every switch is sent
// explicitly (unset spec fields become false -- Cloudflare's own default),
// because partial sends drift forever on this surface.
func buildLogging(accountId string, logging *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewayLogging) *cloudflare.ZeroTrustGatewayLoggingArgs {
	// switches reads one rule type's pair, treating an absent row as
	// all-false (Cloudflare's default).
	switches := func(rule *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewayLoggingRule) (pulumi.BoolPtrInput, pulumi.BoolPtrInput) {
		logAll, logBlocks := false, false
		if rule != nil {
			logAll = rule.GetLogAll()
			logBlocks = rule.GetLogBlocks()
		}
		return pulumi.BoolPtr(logAll), pulumi.BoolPtr(logBlocks)
	}

	byRuleType := logging.SettingsByRuleType
	if byRuleType == nil {
		byRuleType = &cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewayLoggingByRuleType{}
	}

	dnsAll, dnsBlocks := switches(byRuleType.Dns)
	httpAll, httpBlocks := switches(byRuleType.Http)
	l4All, l4Blocks := switches(byRuleType.L4)

	return &cloudflare.ZeroTrustGatewayLoggingArgs{
		AccountId: pulumi.String(accountId),
		RedactPii: pulumi.BoolPtr(logging.GetRedactPii()),
		SettingsByRuleType: cloudflare.ZeroTrustGatewayLoggingSettingsByRuleTypeArgs{
			Dns:  cloudflare.ZeroTrustGatewayLoggingSettingsByRuleTypeDnsArgs{LogAll: dnsAll, LogBlocks: dnsBlocks},
			Http: cloudflare.ZeroTrustGatewayLoggingSettingsByRuleTypeHttpArgs{LogAll: httpAll, LogBlocks: httpBlocks},
			L4:   cloudflare.ZeroTrustGatewayLoggingSettingsByRuleTypeL4Args{LogAll: l4All, LogBlocks: l4Blocks},
		},
	}
}
