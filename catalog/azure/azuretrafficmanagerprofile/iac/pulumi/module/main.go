package module

import (
	"github.com/pkg/errors"
	azuretrafficmanagerprofilev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuretrafficmanagerprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the Traffic Manager profile -- Azure's DNS-based
// traffic director. The profile owns {relative_name}.trafficmanager.net
// and answers each lookup with the address of one of its endpoints
// (AzureTrafficManagerEndpoint resources referencing this profile),
// chosen by routing method and endpoint health.
//
// Traffic Manager is GLOBAL: the provider pins the ARM location to
// "global" itself, which is why the spec carries no region. This SDK
// serves the resource under the network package (the trafficmanager
// package's Profile is the deprecated legacy token for the same ARM
// object).
func Resources(ctx *pulumi.Context, iacInput *azuretrafficmanagerprofilev1alpha1.AzureTrafficManagerProfileIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureTrafficManagerProfile.Spec
	monitor := spec.MonitorConfig

	// The provider's two-value Enabled/Disabled vocabulary, driven by
	// the spec's bool (unset applies true, the provider's own default).
	profileStatus := "Enabled"
	if spec.Enabled != nil && !*spec.Enabled {
		profileStatus = "Disabled"
	}

	// The DNS TTL is always sent explicitly -- 60 (the Azure portal's
	// own default) when the spec leaves it unset -- so both engines send
	// identical wire shapes.
	dnsTtl := 60
	if spec.DnsConfig.TtlSeconds != nil {
		dnsTtl = int(*spec.DnsConfig.TtlSeconds)
	}

	// Probe cadence defaults (30s interval / 10s timeout / 3 tolerated
	// failures) are always sent explicitly. Spec validation already
	// enforces the provider's fast-interval contract (interval 10
	// narrows timeout to an explicit 5-9).
	interval := 30
	if monitor.IntervalInSeconds != nil {
		interval = int(*monitor.IntervalInSeconds)
	}
	timeout := 10
	if monitor.TimeoutInSeconds != nil {
		timeout = int(*monitor.TimeoutInSeconds)
	}
	toleratedFailures := 3
	if monitor.ToleratedNumberOfFailures != nil {
		toleratedFailures = int(*monitor.ToleratedNumberOfFailures)
	}

	monitorArgs := &network.TrafficManagerProfileMonitorConfigArgs{
		Protocol:                  pulumi.String(monitor.Protocol),
		Port:                      pulumi.Int(int(monitor.GetPort())),
		IntervalInSeconds:         pulumi.Int(interval),
		TimeoutInSeconds:          pulumi.Int(timeout),
		ToleratedNumberOfFailures: pulumi.Int(toleratedFailures),
	}
	if monitor.Path != "" {
		monitorArgs.Path = pulumi.String(monitor.Path)
	}
	if len(monitor.ExpectedStatusCodeRanges) > 0 {
		monitorArgs.ExpectedStatusCodeRanges = pulumi.ToStringArray(monitor.ExpectedStatusCodeRanges)
	}
	if len(monitor.CustomHeaders) > 0 {
		headers := network.TrafficManagerProfileMonitorConfigCustomHeaderArray{}
		for _, header := range monitor.CustomHeaders {
			headers = append(headers, &network.TrafficManagerProfileMonitorConfigCustomHeaderArgs{
				Name:  pulumi.String(header.Name),
				Value: pulumi.String(header.Value),
			})
		}
		monitorArgs.CustomHeaders = headers
	}

	profileArgs := &network.TrafficManagerProfileArgs{
		Name:                 pulumi.String(spec.Name),
		ResourceGroupName:    pulumi.String(locals.ResourceGroupName),
		TrafficRoutingMethod: pulumi.String(spec.RoutingMethod),
		ProfileStatus:        pulumi.String(profileStatus),
		TrafficViewEnabled:   pulumi.Bool(spec.TrafficViewEnabled),
		DnsConfig: &network.TrafficManagerProfileDnsConfigArgs{
			// Globally unique across ALL of Azure (the trafficmanager.net
			// namespace is shared) -- Azure rejects a taken name at apply
			// time.
			RelativeName: pulumi.String(spec.DnsConfig.RelativeName),
			Ttl:          pulumi.Int(dnsTtl),
		},
		MonitorConfig: monitorArgs,
		Tags:          pulumi.ToStringMap(locals.AzureTags),
	}
	// Present only for MultiValue routing (spec validation requires it
	// there); sent only when set, mirroring the provider.
	if spec.MaxReturn != nil {
		profileArgs.MaxReturn = pulumi.Int(int(*spec.MaxReturn))
	}

	createdProfile, err := network.NewTrafficManagerProfile(ctx,
		spec.Name,
		profileArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create traffic manager profile %s", spec.Name)
	}

	ctx.Export(OpTrafficManagerProfileId, createdProfile.ID())
	ctx.Export(OpTrafficManagerProfileName, createdProfile.Name)
	ctx.Export(OpFqdn, createdProfile.Fqdn)

	return nil
}
