package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpvpc/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/servicenetworking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpc creates the VPC Network and enables the Compute API when necessary.
// It closely mirrors a Terraform-style module: enable provider‑level services first,
// then declare the core resource, and finally export stack outputs.
func vpc(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*compute.Network, error) {
	// 1. Enable the Compute Engine API for the target project.
	createdComputeService, err := projects.NewService(ctx,
		"compute-api",
		&projects.ServiceArgs{
			Project:                  pulumi.String(locals.GcpVpc.Spec.ProjectId.GetValue()),
			Service:                  pulumi.String("compute.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable compute api")
	}

	// 2. Build the args for the network resource.
	networkArgs := &compute.NetworkArgs{
		AutoCreateSubnetworks: pulumi.BoolPtr(locals.GcpVpc.Spec.AutoCreateSubnetworks),
		Name:                  pulumi.String(locals.GcpVpc.Spec.NetworkName),
		Project:               pulumi.String(locals.GcpVpc.Spec.ProjectId.GetValue()),
	}

	// Map the routing mode enum to the expected GCP value if explicitly set.
	if locals.GcpVpc.Spec.GetRoutingMode() != gcpvpcv1.GcpVpcRoutingMode_REGIONAL {
		// GLOBAL is the only alternative at present.
		networkArgs.RoutingMode = pulumi.StringPtr("GLOBAL")
	}

	// 3. Create the VPC network.
	createdNetwork, err := compute.NewNetwork(ctx,
		"vpc",
		networkArgs,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdComputeService}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vpc network")
	}

	// 4. Export stack outputs.
	ctx.Export(OpNetworkSelfLink, createdNetwork.SelfLink)
	ctx.Export(OpNetworkName, createdNetwork.Name)
	ctx.Export(OpNetworkId, createdNetwork.ID())

	// 5. Configure Private Services Access if enabled.
	// This creates VPC peering with Google's service network for managed services (Cloud SQL, Memorystore, etc.)
	// PREREQUISITE: servicenetworking.googleapis.com must be enabled on the project via GcpProject.
	if locals.GcpVpc.Spec.PrivateServicesAccess != nil && locals.GcpVpc.Spec.PrivateServicesAccess.Enabled {
		if err := privateServicesAccess(ctx, locals, gcpProvider, createdNetwork); err != nil {
			return nil, errors.Wrap(err, "failed to configure private services access")
		}
	}

	return createdNetwork, nil
}

// privateServicesAccess creates the IP allocation and VPC peering with Google's service network.
// This enables Google managed services (Cloud SQL, Memorystore, etc.) to use private IP addresses.
func privateServicesAccess(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, network *compute.Network) error {
	spec := locals.GcpVpc.Spec
	projectId := spec.ProjectId.GetValue()

	// Determine prefix length (default to /16 if not specified)
	prefixLength := 16
	if spec.PrivateServicesAccess.IpRangePrefixLength > 0 {
		prefixLength = int(spec.PrivateServicesAccess.IpRangePrefixLength)
	}

	// Allocate IP range for private services
	privateIpAlloc, err := compute.NewGlobalAddress(ctx, "private-services-range",
		&compute.GlobalAddressArgs{
			Name:         pulumi.String(fmt.Sprintf("%s-private-svc", spec.NetworkName)),
			Project:      pulumi.String(projectId),
			Purpose:      pulumi.String("VPC_PEERING"),
			AddressType:  pulumi.String("INTERNAL"),
			PrefixLength: pulumi.Int(prefixLength),
			Network:      network.ID(),
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to allocate private services IP range")
	}

	// Create private service connection (VPC peering with Google's service network)
	_, err = servicenetworking.NewConnection(ctx, "private-services-connection",
		&servicenetworking.ConnectionArgs{
			Network:               network.ID(),
			Service:               pulumi.String("servicenetworking.googleapis.com"),
			ReservedPeeringRanges: pulumi.StringArray{privateIpAlloc.Name},
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create private services connection")
	}

	// Export Private Services Access outputs
	ctx.Export(OpPrivateServicesIpRangeName, privateIpAlloc.Name)
	ctx.Export(OpPrivateServicesIpRangeCidr,
		pulumi.Sprintf("%s/%d", privateIpAlloc.Address, prefixLength))

	return nil
}
