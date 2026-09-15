package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustgatewaypolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaypolicy/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// gatewayPolicy creates the Gateway policy. Two provider truths this module
// encodes: `enabled` DEFAULTS TO FALSE at Cloudflare (a policy without
// enabled: true deploys inert), and `rule_settings` is ALWAYS sent -- an empty
// object when the spec configures nothing -- because the provider's own test
// fixtures do exactly that to prevent API drift.
//
// KNOWN UPSTREAM DEFECT at v5.23.0 (unfixed through v5.24.0 and provider
// main; upstream issue #7106): the resource's Computed attributes ship no
// UseStateForUnknown plan modifiers, so refresh-inclusive plans NEVER
// converge -- measured live 2026-08-26 on every configuration shape
// (rule_settings absent, empty object, and populated all re-plan an in-place
// update forever; the visible driver is
// rule_settings.ignore_cname_category_matches, which Cloudflare's API never
// echoes back). Module-side remedies were measured non-fixing: an explicit
// send is accepted-and-dropped by the API, and IgnoreChanges cannot suppress
// a provider-planned unknown. Pulumi previews do not refresh, so the drift
// surfaces only under `pulumi refresh`; applies succeed and the repeated
// update is a no-op write. The earlier-known add_headers/override_ips
// first-apply drift is the same family.
func gatewayPolicy(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustGatewayPolicy.Spec

	args := &cloudflare.ZeroTrustGatewayPolicyArgs{
		AccountId:    pulumi.String(spec.AccountId),
		Name:         pulumi.String(spec.Name),
		Action:       pulumi.String(spec.Action),
		RuleSettings: buildRuleSettingsArgs(spec.RuleSettings),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Cloudflare models the filter as a list that can only contain a single
	// value (API-enforced); the spec's singular string keeps that closed.
	if spec.Filter != "" {
		args.Filters = pulumi.StringArray{pulumi.String(spec.Filter)}
	}

	// Explicit pass-through: unset inherits Cloudflare's default of FALSE.
	if spec.Enabled != nil {
		args.Enabled = pulumi.BoolPtr(*spec.Enabled)
	}

	// Unset lets Cloudflare assign a precedence; lower runs earlier.
	if spec.Precedence != nil {
		args.Precedence = pulumi.IntPtr(int(*spec.Precedence))
	}

	// Wirefilter expressions. The API reformats them before storing; if the
	// plan shows drift, adopt the API-returned form.
	if spec.Traffic != "" {
		args.Traffic = pulumi.StringPtr(spec.Traffic)
	}
	if spec.Identity != "" {
		args.Identity = pulumi.StringPtr(spec.Identity)
	}
	if spec.DevicePosture != "" {
		args.DevicePosture = pulumi.StringPtr(spec.DevicePosture)
	}

	if spec.Expiration != nil {
		expirationArgs := &cloudflare.ZeroTrustGatewayPolicyExpirationArgs{
			ExpiresAt: pulumi.String(spec.Expiration.ExpiresAt),
		}
		if spec.Expiration.Duration != nil {
			expirationArgs.Duration = pulumi.IntPtr(int(*spec.Expiration.Duration))
		}
		args.Expiration = expirationArgs
	}

	if spec.Schedule != nil {
		scheduleArgs := &cloudflare.ZeroTrustGatewayPolicyScheduleArgs{}
		setIfNonEmpty := func(target *pulumi.StringPtrInput, value string) {
			if value != "" {
				*target = pulumi.StringPtr(value)
			}
		}
		setIfNonEmpty(&scheduleArgs.Mon, spec.Schedule.Mon)
		setIfNonEmpty(&scheduleArgs.Tue, spec.Schedule.Tue)
		setIfNonEmpty(&scheduleArgs.Wed, spec.Schedule.Wed)
		setIfNonEmpty(&scheduleArgs.Thu, spec.Schedule.Thu)
		setIfNonEmpty(&scheduleArgs.Fri, spec.Schedule.Fri)
		setIfNonEmpty(&scheduleArgs.Sat, spec.Schedule.Sat)
		setIfNonEmpty(&scheduleArgs.Sun, spec.Schedule.Sun)
		setIfNonEmpty(&scheduleArgs.TimeZone, spec.Schedule.TimeZone)
		args.Schedule = scheduleArgs
	}

	createdPolicy, err := cloudflare.NewZeroTrustGatewayPolicy(
		ctx,
		"gateway_policy",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create gateway policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())
	ctx.Export(OpPrecedence, createdPolicy.Precedence)

	return nil
}

// buildRuleSettingsArgs maps the spec's rule settings onto the SDK's args,
// sending only what the manifest set. It always returns a non-nil args object:
// the empty settings object is the provider's own drift workaround.
func buildRuleSettingsArgs(
	ruleSettings *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicyRuleSettings,
) *cloudflare.ZeroTrustGatewayPolicyRuleSettingsArgs {
	settingsArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsArgs{}
	if ruleSettings == nil {
		return settingsArgs
	}

	// The spec's wrapper map values ({values: [...]}) unwrap to the provider's
	// plain map-of-lists shape.
	if len(ruleSettings.AddHeaders) > 0 {
		addHeaders := pulumi.StringArrayMap{}
		for header, wrappedValues := range ruleSettings.AddHeaders {
			addHeaders[header] = pulumi.ToStringArray(wrappedValues.Values)
		}
		settingsArgs.AddHeaders = addHeaders
	}
	if ruleSettings.AllowChildBypass != nil {
		settingsArgs.AllowChildBypass = pulumi.BoolPtr(*ruleSettings.AllowChildBypass)
	}
	if ruleSettings.AuditSsh != nil {
		settingsArgs.AuditSsh = &cloudflare.ZeroTrustGatewayPolicyRuleSettingsAuditSshArgs{
			CommandLogging: pulumi.BoolPtr(ruleSettings.AuditSsh.GetCommandLogging()),
		}
	}
	if ruleSettings.BisoAdminControls != nil {
		settingsArgs.BisoAdminControls = buildBisoAdminControlsArgs(ruleSettings.BisoAdminControls)
	}
	if ruleSettings.BlockPage != nil {
		blockPageArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsBlockPageArgs{
			TargetUri:      pulumi.String(ruleSettings.BlockPage.TargetUri),
			IncludeContext: pulumi.BoolPtr(ruleSettings.BlockPage.IncludeContext),
		}
		settingsArgs.BlockPage = blockPageArgs
	}
	if ruleSettings.BlockPageEnabled != nil {
		settingsArgs.BlockPageEnabled = pulumi.BoolPtr(*ruleSettings.BlockPageEnabled)
	}
	if ruleSettings.BlockReason != "" {
		settingsArgs.BlockReason = pulumi.StringPtr(ruleSettings.BlockReason)
	}
	if ruleSettings.BypassParentRule != nil {
		settingsArgs.BypassParentRule = pulumi.BoolPtr(*ruleSettings.BypassParentRule)
	}
	if ruleSettings.CheckSession != nil {
		checkSessionArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsCheckSessionArgs{
			Enforce: pulumi.BoolPtr(ruleSettings.CheckSession.Enforce),
		}
		if ruleSettings.CheckSession.Duration != "" {
			checkSessionArgs.Duration = pulumi.StringPtr(ruleSettings.CheckSession.Duration)
		}
		settingsArgs.CheckSession = checkSessionArgs
	}
	if len(ruleSettings.DeleteHeaders) > 0 {
		settingsArgs.DeleteHeaders = pulumi.ToStringArray(ruleSettings.DeleteHeaders)
	}
	if ruleSettings.DnsResolvers != nil {
		settingsArgs.DnsResolvers = buildDnsResolversArgs(ruleSettings.DnsResolvers)
	}
	if ruleSettings.Egress != nil {
		egressArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsEgressArgs{}
		if ruleSettings.Egress.Ipv4 != "" {
			egressArgs.Ipv4 = pulumi.StringPtr(ruleSettings.Egress.Ipv4)
		}
		if ruleSettings.Egress.Ipv4Fallback != "" {
			egressArgs.Ipv4Fallback = pulumi.StringPtr(ruleSettings.Egress.Ipv4Fallback)
		}
		if ruleSettings.Egress.Ipv6 != "" {
			egressArgs.Ipv6 = pulumi.StringPtr(ruleSettings.Egress.Ipv6)
		}
		settingsArgs.Egress = egressArgs
	}
	if ruleSettings.ForensicCopy != nil {
		settingsArgs.ForensicCopy = &cloudflare.ZeroTrustGatewayPolicyRuleSettingsForensicCopyArgs{
			Enabled: pulumi.BoolPtr(ruleSettings.ForensicCopy.GetEnabled()),
		}
	}
	if ruleSettings.IgnoreCnameCategoryMatches != nil {
		settingsArgs.IgnoreCnameCategoryMatches = pulumi.BoolPtr(*ruleSettings.IgnoreCnameCategoryMatches)
	}
	if ruleSettings.InsecureDisableDnssecValidation != nil {
		settingsArgs.InsecureDisableDnssecValidation = pulumi.BoolPtr(*ruleSettings.InsecureDisableDnssecValidation)
	}
	if ruleSettings.IpCategories != nil {
		settingsArgs.IpCategories = pulumi.BoolPtr(*ruleSettings.IpCategories)
	}
	if ruleSettings.IpIndicatorFeeds != nil {
		settingsArgs.IpIndicatorFeeds = pulumi.BoolPtr(*ruleSettings.IpIndicatorFeeds)
	}
	if ruleSettings.L4Override != nil {
		l4overrideArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsL4overrideArgs{}
		if ruleSettings.L4Override.Ip != "" {
			l4overrideArgs.Ip = pulumi.StringPtr(ruleSettings.L4Override.Ip)
		}
		if ruleSettings.L4Override.Port != nil {
			l4overrideArgs.Port = pulumi.IntPtr(int(*ruleSettings.L4Override.Port))
		}
		settingsArgs.L4override = l4overrideArgs
	}
	if ruleSettings.NotificationSettings != nil {
		notificationArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsNotificationSettingsArgs{
			Enabled:        pulumi.BoolPtr(ruleSettings.NotificationSettings.Enabled),
			IncludeContext: pulumi.BoolPtr(ruleSettings.NotificationSettings.IncludeContext),
		}
		if ruleSettings.NotificationSettings.Msg != "" {
			notificationArgs.Msg = pulumi.StringPtr(ruleSettings.NotificationSettings.Msg)
		}
		if ruleSettings.NotificationSettings.SupportUrl != "" {
			notificationArgs.SupportUrl = pulumi.StringPtr(ruleSettings.NotificationSettings.SupportUrl)
		}
		settingsArgs.NotificationSettings = notificationArgs
	}
	if ruleSettings.OverrideHost != "" {
		settingsArgs.OverrideHost = pulumi.StringPtr(ruleSettings.OverrideHost)
	}
	if len(ruleSettings.OverrideIps) > 0 {
		settingsArgs.OverrideIps = pulumi.ToStringArray(ruleSettings.OverrideIps)
	}
	if ruleSettings.PayloadLog != nil {
		settingsArgs.PayloadLog = &cloudflare.ZeroTrustGatewayPolicyRuleSettingsPayloadLogArgs{
			Enabled: pulumi.BoolPtr(ruleSettings.PayloadLog.GetEnabled()),
		}
	}
	if ruleSettings.Quarantine != nil {
		quarantineArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsQuarantineArgs{}
		if len(ruleSettings.Quarantine.FileTypes) > 0 {
			quarantineArgs.FileTypes = pulumi.ToStringArray(ruleSettings.Quarantine.FileTypes)
		}
		settingsArgs.Quarantine = quarantineArgs
	}
	if ruleSettings.Redirect != nil {
		settingsArgs.Redirect = &cloudflare.ZeroTrustGatewayPolicyRuleSettingsRedirectArgs{
			TargetUri:            pulumi.String(ruleSettings.Redirect.TargetUri),
			IncludeContext:       pulumi.BoolPtr(ruleSettings.Redirect.IncludeContext),
			PreservePathAndQuery: pulumi.BoolPtr(ruleSettings.Redirect.PreservePathAndQuery),
		}
	}
	if ruleSettings.ResolveDnsInternally != nil {
		resolveInternallyArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsResolveDnsInternallyArgs{}
		if ruleSettings.ResolveDnsInternally.Fallback != "" {
			resolveInternallyArgs.Fallback = pulumi.StringPtr(ruleSettings.ResolveDnsInternally.Fallback)
		}
		if ruleSettings.ResolveDnsInternally.ViewId != "" {
			resolveInternallyArgs.ViewId = pulumi.StringPtr(ruleSettings.ResolveDnsInternally.ViewId)
		}
		settingsArgs.ResolveDnsInternally = resolveInternallyArgs
	}
	if ruleSettings.ResolveDnsThroughCloudflare != nil {
		settingsArgs.ResolveDnsThroughCloudflare = pulumi.BoolPtr(*ruleSettings.ResolveDnsThroughCloudflare)
	}
	if len(ruleSettings.SetHeaders) > 0 {
		setHeaders := pulumi.StringArrayMap{}
		for header, wrappedValues := range ruleSettings.SetHeaders {
			setHeaders[header] = pulumi.ToStringArray(wrappedValues.Values)
		}
		settingsArgs.SetHeaders = setHeaders
	}
	if ruleSettings.UntrustedCert != nil {
		untrustedCertArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsUntrustedCertArgs{}
		if ruleSettings.UntrustedCert.Action != "" {
			untrustedCertArgs.Action = pulumi.StringPtr(ruleSettings.UntrustedCert.Action)
		}
		settingsArgs.UntrustedCert = untrustedCertArgs
	}

	return settingsArgs
}

// buildBisoAdminControlsArgs maps the browser-isolation controls. The v2
// string controls and v1 booleans coexist; only set fields travel.
func buildBisoAdminControlsArgs(
	bisoControls *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicyBisoAdminControls,
) *cloudflare.ZeroTrustGatewayPolicyRuleSettingsBisoAdminControlsArgs {
	bisoArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsBisoAdminControlsArgs{}

	if bisoControls.Version != "" {
		bisoArgs.Version = pulumi.StringPtr(bisoControls.Version)
	}
	if bisoControls.Copy != "" {
		bisoArgs.Copy = pulumi.StringPtr(bisoControls.Copy)
	}
	if bisoControls.Download != "" {
		bisoArgs.Download = pulumi.StringPtr(bisoControls.Download)
	}
	if bisoControls.Paste != "" {
		bisoArgs.Paste = pulumi.StringPtr(bisoControls.Paste)
	}
	if bisoControls.Keyboard != "" {
		bisoArgs.Keyboard = pulumi.StringPtr(bisoControls.Keyboard)
	}
	if bisoControls.Printing != "" {
		bisoArgs.Printing = pulumi.StringPtr(bisoControls.Printing)
	}
	if bisoControls.Upload != "" {
		bisoArgs.Upload = pulumi.StringPtr(bisoControls.Upload)
	}
	if bisoControls.Dcp != nil {
		bisoArgs.Dcp = pulumi.BoolPtr(*bisoControls.Dcp)
	}
	if bisoControls.Dd != nil {
		bisoArgs.Dd = pulumi.BoolPtr(*bisoControls.Dd)
	}
	if bisoControls.Dk != nil {
		bisoArgs.Dk = pulumi.BoolPtr(*bisoControls.Dk)
	}
	if bisoControls.Dp != nil {
		bisoArgs.Dp = pulumi.BoolPtr(*bisoControls.Dp)
	}
	if bisoControls.Du != nil {
		bisoArgs.Du = pulumi.BoolPtr(*bisoControls.Du)
	}
	if bisoControls.WmId != "" {
		bisoArgs.WmId = pulumi.StringPtr(bisoControls.WmId)
	}

	return bisoArgs
}

// buildDnsResolversArgs maps the custom upstream resolvers. vnet_id arrives as
// a resolved string (the spec's reference is resolved before the module runs).
func buildDnsResolversArgs(
	dnsResolvers *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicyDnsResolvers,
) *cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversArgs {
	resolversArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversArgs{}

	if len(dnsResolvers.Ipv4) > 0 {
		ipv4Resolvers := make(cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversIpv4Array, 0, len(dnsResolvers.Ipv4))
		for _, resolver := range dnsResolvers.Ipv4 {
			resolverArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversIpv4Args{
				Ip: pulumi.String(resolver.Ip),
			}
			if resolver.Port != nil {
				resolverArgs.Port = pulumi.IntPtr(int(*resolver.Port))
			}
			if resolver.RouteThroughPrivateNetwork {
				resolverArgs.RouteThroughPrivateNetwork = pulumi.BoolPtr(true)
			}
			if resolver.VnetId.GetValue() != "" {
				resolverArgs.VnetId = pulumi.StringPtr(resolver.VnetId.GetValue())
			}
			ipv4Resolvers = append(ipv4Resolvers, resolverArgs)
		}
		resolversArgs.Ipv4s = ipv4Resolvers
	}

	if len(dnsResolvers.Ipv6) > 0 {
		ipv6Resolvers := make(cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversIpv6Array, 0, len(dnsResolvers.Ipv6))
		for _, resolver := range dnsResolvers.Ipv6 {
			resolverArgs := &cloudflare.ZeroTrustGatewayPolicyRuleSettingsDnsResolversIpv6Args{
				Ip: pulumi.String(resolver.Ip),
			}
			if resolver.Port != nil {
				resolverArgs.Port = pulumi.IntPtr(int(*resolver.Port))
			}
			if resolver.RouteThroughPrivateNetwork {
				resolverArgs.RouteThroughPrivateNetwork = pulumi.BoolPtr(true)
			}
			if resolver.VnetId.GetValue() != "" {
				resolverArgs.VnetId = pulumi.StringPtr(resolver.VnetId.GetValue())
			}
			ipv6Resolvers = append(ipv6Resolvers, resolverArgs)
		}
		resolversArgs.Ipv6s = ipv6Resolvers
	}

	return resolversArgs
}
