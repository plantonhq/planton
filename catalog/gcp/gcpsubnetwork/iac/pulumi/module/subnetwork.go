package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetwork provisions a subnetwork in a custom-mode VPC — the regional
// address space workloads live in: primary IPv4 range for VM interfaces,
// secondary ranges for alias IPs (GKE pods/services), optional IPv6, and
// VPC Flow Logs.
//
// name, project, region, network, and description are immutable (ForceNew in
// the provider): changing any of them destroys and recreates the subnet — an
// outage for everything addressed in it. The primary range is the one
// asymmetric knob: EXPANDING ip_cidr_range updates in place, shrinking
// recreates.
//
// purpose creates the special-role subnets other features depend on:
// REGIONAL_MANAGED_PROXY reserves Envoy address space for the region's
// regional load balancers; PRIVATE_SERVICE_CONNECT backs published PSC
// services.
func subnetwork(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*compute.Subnetwork, error) {
	spec := locals.GcpSubnetwork.Spec

	// Enable the Compute Engine API so a fresh project can host the subnet.
	// disable_on_destroy stays false (the provider default): tearing down one
	// subnet must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		DisableDependentServices: pulumi.BoolPtr(true),
		Service:                  pulumi.String("compute.googleapis.com"),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project. Leaving Project unset lets the gcp provider
	// resolve its own project (configuration or the GOOGLE_PROJECT /
	// GOOGLE_CLOUD_PROJECT environment chain); an empty string would be sent
	// verbatim and rejected.
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"subnetwork-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	// Drop entries that arrive as empty objects so the API never sees a blank
	// secondary range — identical to the Terraform module's filter. A range
	// is real when it has a name and either CIDR source (literal or reserved
	// internal range).
	var secondaryRanges compute.SubnetworkSecondaryIpRangeArray
	for _, secondaryRange := range spec.SecondaryIpRanges {
		if secondaryRange.RangeName == "" ||
			(secondaryRange.IpCidrRange == "" && secondaryRange.ReservedInternalRange == "") {
			continue
		}
		rangeArgs := &compute.SubnetworkSecondaryIpRangeArgs{
			RangeName: pulumi.String(secondaryRange.RangeName),
		}
		if secondaryRange.IpCidrRange != "" {
			rangeArgs.IpCidrRange = pulumi.String(secondaryRange.IpCidrRange)
		}
		if secondaryRange.ReservedInternalRange != "" {
			rangeArgs.ReservedInternalRange = pulumi.String(secondaryRange.ReservedInternalRange)
		}
		secondaryRanges = append(secondaryRanges, rangeArgs)
	}

	args := &compute.SubnetworkArgs{
		Name:    pulumi.String(spec.SubnetworkName),
		Region:  pulumi.String(spec.Region),
		Network: pulumi.String(spec.VpcSelfLink.GetValue()),

		// Planton middleware applies the spec's proto defaults (PRIVATE /
		// IPV4_ONLY) before the module runs; GetPurpose/GetStackType echo them
		// here so direct invocations behave identically.
		Purpose:   pulumi.String(spec.GetPurpose()),
		StackType: pulumi.String(spec.GetStackType()),

		PrivateIpGoogleAccess: pulumi.BoolPtr(spec.PrivateIpGoogleAccess),

		// Safety latch: by default an empty secondary-range list is NOT sent
		// on update, so a partial manifest cannot silently wipe GKE pod
		// ranges.
		SendSecondaryIpRangeIfEmpty: pulumi.BoolPtr(spec.SendSecondaryIpRangeIfEmpty),

		// Deliberate address-space reclaims only: subnet routes still win
		// over the overlapping peer/on-prem routes this permits.
		AllowSubnetCidrRoutesOverlap: pulumi.BoolPtr(spec.AllowSubnetCidrRoutesOverlap),

		SecondaryIpRanges: secondaryRanges,
	}

	// Omitted optionals stay unset (matching the Terraform module's null)
	// rather than being sent as empty strings the API would reject.
	if spec.IpCidrRange != "" {
		args.IpCidrRange = pulumi.String(spec.IpCidrRange)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.Role != "" {
		args.Role = pulumi.String(spec.Role)
	}
	if spec.PrivateIpv6GoogleAccess != "" {
		args.PrivateIpv6GoogleAccess = pulumi.String(spec.PrivateIpv6GoogleAccess)
	}
	// Dual-stack wiring: ipv6_access_type decides whether the assigned prefix
	// is internet-routable (EXTERNAL GUAs) or VPC-internal (INTERNAL ULAs —
	// requires the VPC's ULA range enabled).
	if spec.Ipv6AccessType != "" {
		args.Ipv6AccessType = pulumi.String(spec.Ipv6AccessType)
	}
	if spec.ExternalIpv6Prefix != "" {
		args.ExternalIpv6Prefix = pulumi.String(spec.ExternalIpv6Prefix)
	}
	// INTERNAL counterpart of external_ipv6_prefix: pin a ULA prefix instead
	// of letting Google allocate one from the VPC's internal IPv6 range.
	if spec.InternalIpv6Prefix != "" {
		args.InternalIpv6Prefix = pulumi.String(spec.InternalIpv6Prefix)
	}
	// Alternative primary-CIDR source: a centrally allocated Network
	// Connectivity internal range (exactly one of the two is set,
	// spec-enforced).
	if spec.ReservedInternalRange != "" {
		args.ReservedInternalRange = pulumi.String(spec.ReservedInternalRange)
	}
	// BYOIP: draw the subnet's IPv6 space from a PublicDelegatedPrefix.
	if spec.IpCollection != "" {
		args.IpCollection = pulumi.String(spec.IpCollection)
	}
	// ARP subnet-mask resolution for appliance/NFV subnets.
	if spec.ResolveSubnetMask != "" {
		args.ResolveSubnetMask = pulumi.String(spec.ResolveSubnetMask)
	}
	// Create-time Resource Manager tag bindings; the params block is
	// create-only, so tag changes replace the subnetwork.
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.SubnetworkParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// VPC Flow Logs: presence of the message enables logging. Defaults mirror
	// the API's own (5s aggregation, 50% sampling, all metadata) so an empty
	// spec object behaves sanely — identical to the Terraform module.
	// The block's switch (on by default once declared, filled by the platform
	// before this runs) decides whether flow logs render.
	if spec.LogConfig != nil && spec.LogConfig.GetEnabled() {
		logConfig := &compute.SubnetworkLogConfigArgs{
			AggregationInterval: pulumi.String(spec.LogConfig.GetAggregationInterval()),
			FlowSampling:        pulumi.Float64(spec.LogConfig.GetFlowSampling()),
			Metadata:            pulumi.String(spec.LogConfig.GetMetadata()),
		}
		// metadata_fields only accompanies CUSTOM_METADATA (spec-enforced).
		if spec.LogConfig.GetMetadata() == "CUSTOM_METADATA" && len(spec.LogConfig.MetadataFields) > 0 {
			metadataFields := pulumi.StringArray{}
			for _, metadataField := range spec.LogConfig.MetadataFields {
				metadataFields = append(metadataFields, pulumi.String(metadataField))
			}
			logConfig.MetadataFields = metadataFields
		}
		if spec.LogConfig.FilterExpr != "" {
			logConfig.FilterExpr = pulumi.String(spec.LogConfig.FilterExpr)
		}
		args.LogConfig = logConfig
	}

	createdSubnetwork, err := compute.NewSubnetwork(ctx, "subnetwork", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create subnetwork")
	}

	ctx.Export(OpSubnetworkSelfLink, createdSubnetwork.SelfLink)
	ctx.Export(OpSubnetworkName, createdSubnetwork.Name)
	ctx.Export(OpRegion, createdSubnetwork.Region)
	ctx.Export(OpIpCidrRange, createdSubnetwork.IpCidrRange)
	ctx.Export(OpGatewayAddress, createdSubnetwork.GatewayAddress)
	// Exported as a string for a stable cross-engine output shape (the
	// Terraform module does the same with tostring()).
	ctx.Export(OpSubnetworkId, createdSubnetwork.SubnetworkId.ApplyT(func(id int) string {
		return fmt.Sprintf("%d", id)
	}).(pulumi.StringOutput))
	ctx.Export(OpInternalIpv6Prefix, createdSubnetwork.InternalIpv6Prefix)
	ctx.Export(OpExternalIpv6Prefix, createdSubnetwork.ExternalIpv6Prefix)

	// Export secondary ranges with individual per-index keys so the outputs
	// transformer can map them onto the repeated proto message. The exports
	// are registered HERE, synchronously, with values DERIVED from the
	// resolved array — never inside an ApplyT callback. Apply callbacks run
	// on output-resolution goroutines, and a ctx.Export there writes the
	// SDK's shared exports map while the engine's end-of-program
	// stack-output marshaling reads it: a data race that crashes the whole
	// program with `fatal error: concurrent map read and map write`,
	// timing-dependent and therefore flaky. The index space is known up
	// front from the spec (the same filtered list sent to the API), so
	// every key can be registered before the program returns.
	for index := range secondaryRanges {
		ctx.Export(fmt.Sprintf("%s.%d.%s", OpSecondaryRanges, index, "range_name"),
			createdSubnetwork.SecondaryIpRanges.ApplyT(func(ranges []compute.SubnetworkSecondaryIpRange) string {
				if index < len(ranges) {
					return ranges[index].RangeName
				}
				return ""
			}).(pulumi.StringOutput))
		ctx.Export(fmt.Sprintf("%s.%d.%s", OpSecondaryRanges, index, "ip_cidr_range"),
			createdSubnetwork.SecondaryIpRanges.ApplyT(func(ranges []compute.SubnetworkSecondaryIpRange) string {
				if index < len(ranges) && ranges[index].IpCidrRange != nil {
					return *ranges[index].IpCidrRange
				}
				return ""
			}).(pulumi.StringOutput))
	}

	return createdSubnetwork, nil
}
