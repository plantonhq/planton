package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustdevicedefaultprofilev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdevicedefaultprofile/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// defaultProfile applies the account's default WARP device profile and its
// folded companions.
//
// The profile always exists on the account -- create and update are the same
// PATCH against the singleton, and DESTROY IS A NO-OP that leaves the
// last-applied values standing. Unset spec fields are never sent, leaving
// the live value (or Cloudflare's default) untouched.
func defaultProfile(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustDeviceDefaultProfile.Spec

	args := &cloudflare.ZeroTrustDeviceDefaultProfileArgs{
		AccountId: pulumi.String(spec.AccountId),
	}

	if spec.AllowModeSwitch != nil {
		args.AllowModeSwitch = pulumi.BoolPtr(spec.GetAllowModeSwitch())
	}
	if spec.AllowUpdates != nil {
		args.AllowUpdates = pulumi.BoolPtr(spec.GetAllowUpdates())
	}
	if spec.AllowedToLeave != nil {
		args.AllowedToLeave = pulumi.BoolPtr(spec.GetAllowedToLeave())
	}
	if spec.AutoConnect != nil {
		args.AutoConnect = pulumi.Float64Ptr(float64(spec.GetAutoConnect()))
	}
	if spec.CaptivePortal != nil {
		args.CaptivePortal = pulumi.Float64Ptr(float64(spec.GetCaptivePortal()))
	}
	if spec.DisableAutoFallback != nil {
		args.DisableAutoFallback = pulumi.BoolPtr(spec.GetDisableAutoFallback())
	}
	if spec.ExcludeOfficeIps != nil {
		args.ExcludeOfficeIps = pulumi.BoolPtr(spec.GetExcludeOfficeIps())
	}
	if spec.RegisterInterfaceIpWithDns != nil {
		args.RegisterInterfaceIpWithDns = pulumi.BoolPtr(spec.GetRegisterInterfaceIpWithDns())
	}
	if spec.SccmVpnBoundarySupport != nil {
		args.SccmVpnBoundarySupport = pulumi.BoolPtr(spec.GetSccmVpnBoundarySupport())
	}
	if spec.SwitchLocked != nil {
		args.SwitchLocked = pulumi.BoolPtr(spec.GetSwitchLocked())
	}
	if spec.SupportUrl != "" {
		args.SupportUrl = pulumi.String(spec.SupportUrl)
	}
	if spec.TunnelProtocol != "" {
		args.TunnelProtocol = pulumi.String(spec.TunnelProtocol)
	}
	if spec.LanAllowMinutes != nil {
		args.LanAllowMinutes = pulumi.Float64Ptr(float64(spec.GetLanAllowMinutes()))
	}
	if spec.LanAllowSubnetSize != nil {
		args.LanAllowSubnetSize = pulumi.Float64Ptr(float64(spec.GetLanAllowSubnetSize()))
	}

	// Split tunnel: exclude and include are mutually exclusive (spec
	// validation enforces it). Each entry carries exactly one of address or
	// host; the unset one is never sent.
	if len(spec.Exclude) > 0 {
		excludes := cloudflare.ZeroTrustDeviceDefaultProfileExcludeArray{}
		for _, entry := range spec.Exclude {
			row := cloudflare.ZeroTrustDeviceDefaultProfileExcludeArgs{}
			if entry.Address != "" {
				row.Address = pulumi.String(entry.Address)
			}
			if entry.Host != "" {
				row.Host = pulumi.String(entry.Host)
			}
			if entry.Description != "" {
				row.Description = pulumi.String(entry.Description)
			}
			excludes = append(excludes, row)
		}
		args.Excludes = excludes
	}

	if len(spec.Include) > 0 {
		includes := cloudflare.ZeroTrustDeviceDefaultProfileIncludeArray{}
		for _, entry := range spec.Include {
			row := cloudflare.ZeroTrustDeviceDefaultProfileIncludeArgs{}
			if entry.Address != "" {
				row.Address = pulumi.String(entry.Address)
			}
			if entry.Host != "" {
				row.Host = pulumi.String(entry.Host)
			}
			if entry.Description != "" {
				row.Description = pulumi.String(entry.Description)
			}
			includes = append(includes, row)
		}
		args.Includes = includes
	}

	if spec.ServiceModeV2 != nil {
		serviceMode := cloudflare.ZeroTrustDeviceDefaultProfileServiceModeV2Args{
			Mode: pulumi.StringPtr(spec.ServiceModeV2.Mode),
		}
		if spec.ServiceModeV2.Port != nil {
			serviceMode.Port = pulumi.Float64Ptr(float64(spec.ServiceModeV2.GetPort()))
		}
		args.ServiceModeV2 = serviceMode
	}

	if spec.VirtualNetworks != nil {
		allowed := make([]string, 0, len(spec.VirtualNetworks.Allowed))
		for _, vnet := range spec.VirtualNetworks.Allowed {
			allowed = append(allowed, vnet.GetValue())
		}
		args.VirtualNetworks = cloudflare.ZeroTrustDeviceDefaultProfileVirtualNetworksArgs{
			Alloweds: pulumi.ToStringArray(allowed),
			Default:  pulumi.String(spec.VirtualNetworks.DefaultVirtualNetworkId.GetValue()),
		}
	}

	// Always send the list, empty included. The API echoes [] for an empty
	// suffix list, and the provider's Computed+Optional attribute carries no
	// state-preserving plan modifier -- a null config re-plans an in-place
	// update forever on refresh-inclusive plans (measured live 2026-08-27 at
	// v5.23.0 on the Terraform engine; previews hide the same latent class
	// here). Sending [] matches the spec's documented contract (an empty
	// list clears the account's list; proto3 cannot distinguish unset from
	// empty), and converges.
	suffixes := cloudflare.ZeroTrustDeviceDefaultProfileDnsSearchSuffixArray{}
	for _, row := range spec.DnsSearchSuffixes {
		suffix := cloudflare.ZeroTrustDeviceDefaultProfileDnsSearchSuffixArgs{
			Suffix: pulumi.String(row.Suffix),
		}
		if row.Description != "" {
			suffix.Description = pulumi.String(row.Description)
		}
		suffixes = append(suffixes, suffix)
	}
	args.DnsSearchSuffixes = suffixes

	// policy_id is a pure-computed server echo (the singleton's stable id)
	// with no state-preserving plan modifier at v5.23.0: the bridged
	// provider proposes a phantom no-op update on every preview against
	// stored state (measured live 2026-08-27; the value never actually
	// changes). Ignoring it is safe -- the attribute is never sent, and the
	// output still reads the real value after apply.
	createdProfile, err := cloudflare.NewZeroTrustDeviceDefaultProfile(
		ctx,
		"default_profile",
		args,
		pulumi.Provider(cloudflareProvider),
		pulumi.IgnoreChanges([]string{"policyId"}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to apply default device profile")
	}

	// The profile's local-DNS fallback list. The profile resource reports
	// fallback_domains READ-ONLY; this dedicated companion is the only write
	// path, and it replaces the whole list on every apply (declarative: what
	// the spec lists is exactly what exists). Its destroy is also a no-op.
	// Deployed only when the spec declares rows, so an unmanaged list is
	// never touched.
	if len(spec.FallbackDomains) > 0 {
		if err := fallbackDomains(ctx, spec, createdProfile, cloudflareProvider); err != nil {
			return err
		}
	}

	// Per-zone WARP client certificate provisioning -- the one ZONE-scoped
	// surface on this account-scoped kind. Cloudflare offers no delete for
	// it (removing this resource leaves the toggle as last applied) and no
	// import. Deployed only when the spec declares the fold.
	if spec.ZoneCertificates != nil {
		_, err := cloudflare.NewZeroTrustDeviceDefaultProfileCertificates(
			ctx,
			"zone_certificates",
			&cloudflare.ZeroTrustDeviceDefaultProfileCertificatesArgs{
				ZoneId:  pulumi.String(spec.ZoneCertificates.ZoneId.GetValue()),
				Enabled: pulumi.Bool(spec.ZoneCertificates.GetEnabled()),
			},
			pulumi.Provider(cloudflareProvider),
		)
		if err != nil {
			return errors.Wrap(err, "failed to apply zone certificate provisioning")
		}
	}

	ctx.Export(OpAccountId, pulumi.String(spec.AccountId))
	ctx.Export(OpGatewayUniqueId, createdProfile.GatewayUniqueId)
	ctx.Export(OpPolicyId, createdProfile.PolicyId)

	return nil
}

// fallbackDomains applies the folded full-replacement fallback-domain list,
// ordered after the profile.
func fallbackDomains(
	ctx *pulumi.Context,
	spec *cloudflarezerotrustdevicedefaultprofilev1alpha1.CloudflareZeroTrustDeviceDefaultProfileSpec,
	createdProfile *cloudflare.ZeroTrustDeviceDefaultProfile,
	cloudflareProvider *cloudflare.Provider,
) error {
	domains := cloudflare.ZeroTrustDeviceDefaultProfileLocalDomainFallbackDomainArray{}
	for _, row := range spec.FallbackDomains {
		domain := cloudflare.ZeroTrustDeviceDefaultProfileLocalDomainFallbackDomainArgs{
			Suffix: pulumi.String(row.Suffix),
		}
		if row.Description != "" {
			domain.Description = pulumi.String(row.Description)
		}
		if len(row.DnsServer) > 0 {
			domain.DnsServers = pulumi.ToStringArray(row.DnsServer)
		}
		domains = append(domains, domain)
	}

	_, err := cloudflare.NewZeroTrustDeviceDefaultProfileLocalDomainFallback(
		ctx,
		"fallback_domains",
		&cloudflare.ZeroTrustDeviceDefaultProfileLocalDomainFallbackArgs{
			AccountId: pulumi.String(spec.AccountId),
			Domains:   domains,
		},
		pulumi.Provider(cloudflareProvider),
		pulumi.DependsOn([]pulumi.Resource{createdProfile}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to apply fallback domains")
	}
	return nil
}
