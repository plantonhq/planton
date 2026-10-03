package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureprivateendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivateendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatelink"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivateendpointv1alpha1.AzurePrivateEndpointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateEndpoint.Spec

	// The single private service connection. Azure models this as one block
	// per endpoint, so the spec folds it into one message. The connection
	// name is derived from the endpoint name -- it is an internal handle
	// Azure requires but nothing references. References resolve to literal
	// ARM IDs before the module runs, so GetValue() returns the resolved id.
	conn := spec.PrivateServiceConnection
	connectionArgs := &privatelink.EndpointPrivateServiceConnectionArgs{
		Name:               pulumi.String(fmt.Sprintf("%s-connection", spec.Name)),
		IsManualConnection: pulumi.Bool(conn.GetIsManualConnection()),
	}
	if conn.PrivateConnectionResourceId != nil {
		connectionArgs.PrivateConnectionResourceId = pulumi.String(conn.PrivateConnectionResourceId.GetValue())
	}
	if conn.ConnectionAlias != "" {
		connectionArgs.PrivateConnectionResourceAlias = pulumi.String(conn.ConnectionAlias)
	}
	if len(conn.SubresourceNames) > 0 {
		connectionArgs.SubresourceNames = pulumi.ToStringArray(conn.SubresourceNames)
	}
	// request_message only ever accompanies a manual connection; the spec
	// guarantees the pairing, so send it only when non-empty.
	if conn.RequestMessage != "" {
		connectionArgs.RequestMessage = pulumi.String(conn.RequestMessage)
	}

	endpointArgs := &privatelink.EndpointArgs{
		Name:                     pulumi.String(spec.Name),
		Location:                 pulumi.String(spec.Region),
		ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
		SubnetId:                 pulumi.String(spec.SubnetId.GetValue()),
		PrivateServiceConnection: connectionArgs,
		Tags:                     pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.CustomNetworkInterfaceName != "" {
		endpointArgs.CustomNetworkInterfaceName = pulumi.String(spec.CustomNetworkInterfaceName)
	}

	// The DNS zone group registers the endpoint's private IP as an A record
	// in each referenced zone. Without it, the service FQDN resolves to the
	// PUBLIC IP -- so this is created whenever any zone is referenced. The
	// group name is an internal handle derived from the endpoint name.
	if len(spec.PrivateDnsZoneIds) > 0 {
		zoneIds := make([]string, 0, len(spec.PrivateDnsZoneIds))
		for _, z := range spec.PrivateDnsZoneIds {
			zoneIds = append(zoneIds, z.GetValue())
		}
		endpointArgs.PrivateDnsZoneGroup = &privatelink.EndpointPrivateDnsZoneGroupArgs{
			Name:              pulumi.String(fmt.Sprintf("%s-dns-zone-group", spec.Name)),
			PrivateDnsZoneIds: pulumi.ToStringArray(zoneIds),
		}
	}

	// Static IP configurations pin sub-resources to fixed addresses; when
	// empty, Azure allocates dynamically from the subnet (the common case).
	if len(spec.IpConfigurations) > 0 {
		ipConfigs := privatelink.EndpointIpConfigurationArray{}
		for _, ic := range spec.IpConfigurations {
			icArgs := privatelink.EndpointIpConfigurationArgs{
				Name:             pulumi.String(ic.Name),
				PrivateIpAddress: pulumi.String(ic.PrivateIpAddress),
			}
			if ic.SubresourceName != "" {
				icArgs.SubresourceName = pulumi.String(ic.SubresourceName)
			}
			if ic.MemberName != "" {
				icArgs.MemberName = pulumi.String(ic.MemberName)
			}
			ipConfigs = append(ipConfigs, icArgs)
		}
		endpointArgs.IpConfigurations = ipConfigs
	}

	createdEndpoint, err := privatelink.NewEndpoint(ctx,
		spec.Name,
		endpointArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private endpoint %s", spec.Name)
	}

	// Application security group membership is expressed member-side in
	// Azure's model, as its own association resource. ASG references resolve
	// to literal ARM IDs before the module runs.
	for i, asg := range spec.ApplicationSecurityGroupIds {
		if _, err := privatelink.NewApplicationSecurityGroupAssociation(ctx,
			fmt.Sprintf("%s-asg-%d", spec.Name, i),
			&privatelink.ApplicationSecurityGroupAssociationArgs{
				PrivateEndpointId:          createdEndpoint.ID(),
				ApplicationSecurityGroupId: pulumi.String(asg.GetValue()),
			},
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to associate application security group %d with private endpoint %s", i, spec.Name)
		}
	}

	ctx.Export(OpPrivateEndpointId, createdEndpoint.ID())
	ctx.Export(OpPrivateEndpointName, createdEndpoint.Name)

	// The private IP lives on the connection block; the auto-created NIC's
	// id lives in NetworkInterfaces. Both are known only after creation.
	privateIpAddress := createdEndpoint.PrivateServiceConnection.ApplyT(func(c privatelink.EndpointPrivateServiceConnection) string {
		if c.PrivateIpAddress != nil {
			return *c.PrivateIpAddress
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpPrivateIpAddress, privateIpAddress)

	networkInterfaceId := createdEndpoint.NetworkInterfaces.ApplyT(func(nics []privatelink.EndpointNetworkInterface) string {
		if len(nics) > 0 && nics[0].Id != nil {
			return *nics[0].Id
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpNetworkInterfaceId, networkInterfaceId)

	return nil
}
