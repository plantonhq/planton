package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureprivateendpointv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureprivateendpoint/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatelink"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureprivateendpointv1.AzurePrivateEndpointStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateEndpoint.Spec

	// Build PrivateServiceConnection block
	privateServiceConnectionName := fmt.Sprintf("%s-connection", locals.AzurePrivateEndpoint.Metadata.Name)
	privateServiceConnection := &privatelink.EndpointPrivateServiceConnectionArgs{
		Name:                        pulumi.String(privateServiceConnectionName),
		IsManualConnection:          pulumi.Bool(false),
		PrivateConnectionResourceId: pulumi.String(spec.PrivateConnectionResourceId.GetValue()),
		SubresourceNames:            pulumi.ToStringArray(spec.SubresourceNames),
	}

	// Build endpoint args
	endpointArgs := &privatelink.EndpointArgs{
		Name:                     pulumi.String(spec.Name),
		Location:                 pulumi.String(spec.Region),
		ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
		SubnetId:                 pulumi.String(spec.SubnetId.GetValue()),
		PrivateServiceConnection: privateServiceConnection,
		Tags:                     pulumi.ToStringMap(locals.AzureTags),
	}

	// Conditionally add PrivateDnsZoneGroup if PrivateDnsZoneId is provided
	if spec.PrivateDnsZoneId != nil {
		dnsZoneGroupName := fmt.Sprintf("%s-dns-zone-group", locals.AzurePrivateEndpoint.Metadata.Name)
		endpointArgs.PrivateDnsZoneGroup = &privatelink.EndpointPrivateDnsZoneGroupArgs{
			Name:              pulumi.String(dnsZoneGroupName),
			PrivateDnsZoneIds: pulumi.StringArray{pulumi.String(spec.PrivateDnsZoneId.GetValue())},
		}
	}

	// Create the Private Endpoint
	endpoint, err := privatelink.NewEndpoint(ctx,
		spec.Name,
		endpointArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private endpoint %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpPrivateEndpointId, endpoint.ID())

	// Export private IP address from PrivateServiceConnection
	privateIpAddress := endpoint.PrivateServiceConnection.ApplyT(func(conn privatelink.EndpointPrivateServiceConnection) string {
		if conn.PrivateIpAddress != nil {
			return *conn.PrivateIpAddress
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpPrivateIpAddress, privateIpAddress)

	// Export network interface ID from NetworkInterfaces
	networkInterfaceId := endpoint.NetworkInterfaces.ApplyT(func(nics []privatelink.EndpointNetworkInterface) string {
		if len(nics) > 0 && nics[0].Id != nil {
			return *nics[0].Id
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpNetworkInterfaceId, networkInterfaceId)

	return nil
}
