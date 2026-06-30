package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// routerNat creates a CloudRouter and CloudNAT according to the
// GcpRouterNatSpec. It exports the outputs defined in outputs.go.
func routerNat(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*compute.RouterNat, error) {

	// ---------------------------------------------------------------------
	// Router
	// ---------------------------------------------------------------------

	createdRouter, err := compute.NewRouter(ctx,
		"router",
		&compute.RouterArgs{
			Name:   pulumi.String(locals.GcpRouterNat.Spec.RouterName),
			Region: pulumi.String(locals.GcpRouterNat.Spec.Region),
			// VPC is passed as self‑link or short name
			Network: pulumi.String(locals.GcpRouterNat.Spec.VpcSelfLink.GetValue()),
			Project: pulumi.String(locals.GcpRouterNat.Spec.ProjectId.GetValue()),
		},
		pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	// export router self‑link
	ctx.Export(OpRouterSelfLink, createdRouter.SelfLink)

	// ---------------------------------------------------------------------
	// Decide NATIP allocation strategy
	// ---------------------------------------------------------------------

	natIpAllocateOption := pulumi.String("AUTO_ONLY")
	var natIps pulumi.StringArray

	if len(locals.GcpRouterNat.Spec.NatIpNames) > 0 {
		natIpAllocateOption = pulumi.String("MANUAL_ONLY")

		for idx, natIpNameOrRef := range locals.GcpRouterNat.Spec.NatIpNames {
			// Create (or claim) a regional static address for each requested name.
			// Using the same name requested in the spec keeps it transparent.
			resourceName := fmt.Sprintf("nat-ip-%d", idx)

			createdNatIp, err := compute.NewAddress(ctx,
				resourceName,
				&compute.AddressArgs{
					Name:        pulumi.String(natIpNameOrRef.GetValue()),
					Region:      pulumi.String(locals.GcpRouterNat.Spec.Region),
					AddressType: pulumi.String("EXTERNAL"),
					Labels:      pulumi.ToStringMap(locals.GcpLabels),
					Project:     pulumi.String(locals.GcpRouterNat.Spec.ProjectId.GetValue()),
				},
				pulumi.Provider(gcpProvider),
				pulumi.Parent(createdRouter))
			if err != nil {
				return nil, errors.Wrap(err, "failed to create static nat ip")
			}

			// collect self‑links for RouterNat
			natIps = append(natIps, createdNatIp.SelfLink)
		}
	}

	// ---------------------------------------------------------------------
	// Handle subnet scoping
	// ---------------------------------------------------------------------

	subnetworks := compute.RouterNatSubnetworkArray{}
	sourceRangeSetting := pulumi.String("ALL_SUBNETWORKS_ALL_IP_RANGES")

	if len(locals.GcpRouterNat.Spec.SubnetworkSelfLinks) > 0 {
		sourceRangeSetting = pulumi.String("LIST_OF_SUBNETWORKS")

		for _, subnetOrRef := range locals.GcpRouterNat.Spec.SubnetworkSelfLinks {
			subnetworks = append(subnetworks, &compute.RouterNatSubnetworkArgs{
				Name:                  pulumi.String(subnetOrRef.GetValue()),
				SourceIpRangesToNats:  pulumi.StringArray{pulumi.String("ALL_IP_RANGES")},
				SecondaryIpRangeNames: pulumi.StringArray{},
			})
		}
	}

	// ---------------------------------------------------------------------
	// Logging configuration - get directly from enum (values match GCP API strings)
	// ---------------------------------------------------------------------

	var logConfig *compute.RouterNatLogConfigArgs
	if locals.GcpRouterNat.Spec.LogFilter != nil {
		logFilter := locals.GcpRouterNat.Spec.LogFilter.String()
		if logFilter == "DISABLED" {
			logConfig = &compute.RouterNatLogConfigArgs{
				Enable: pulumi.Bool(false),
			}
		} else {
			logConfig = &compute.RouterNatLogConfigArgs{
				Enable: pulumi.Bool(true),
				Filter: pulumi.String(logFilter),
			}
		}
	} else {
		// Default to ERRORS_ONLY if not specified
		logConfig = &compute.RouterNatLogConfigArgs{
			Enable: pulumi.Bool(true),
			Filter: pulumi.String("ERRORS_ONLY"),
		}
	}

	// ---------------------------------------------------------------------
	// CloudNAT
	// ---------------------------------------------------------------------

	createdRouterNat, err := compute.NewRouterNat(ctx,
		"router-nat",
		&compute.RouterNatArgs{
			Name:                          pulumi.String(locals.GcpRouterNat.Spec.NatName),
			Router:                        createdRouter.Name,
			Region:                        pulumi.String(locals.GcpRouterNat.Spec.Region),
			NatIpAllocateOption:           natIpAllocateOption,
			NatIps:                        natIps,
			SourceSubnetworkIpRangesToNat: sourceRangeSetting,
			Subnetworks:                   subnetworks,
			LogConfig:                     logConfig,
			Project:                       pulumi.String(locals.GcpRouterNat.Spec.ProjectId.GetValue()),
		},
		pulumi.Provider(gcpProvider),
		pulumi.Parent(createdRouter))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router nat")
	}

	// ---------------------------------------------------------------------
	// Outputs
	// ---------------------------------------------------------------------

	ctx.Export(OpName, createdRouterNat.Name)
	ctx.Export(OpNatIpAddresses, createdRouterNat.NatIps)

	return createdRouterNat, nil
}
