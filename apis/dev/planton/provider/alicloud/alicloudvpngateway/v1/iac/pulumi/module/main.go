package module

import (
	"github.com/pkg/errors"
	alicloudvpngatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudvpngateway/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudvpngatewayv1.AliCloudVpnGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudVpnGateway.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	gatewayArgs := &vpn.GatewayArgs{
		VpnGatewayName: pulumi.String(spec.VpnGatewayName),
		VpcId:          pulumi.String(spec.VpcId.GetValue()),
		VswitchId:      pulumi.StringPtr(spec.VswitchId.GetValue()),
		Bandwidth:      pulumi.Int(spec.Bandwidth),
		PaymentType:    pulumi.StringPtr(paymentType(spec)),
		Tags:           pulumi.ToStringMap(locals.Tags),
	}

	if spec.Description != "" {
		gatewayArgs.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.EnableSsl != nil {
		gatewayArgs.EnableSsl = pulumi.BoolPtr(*spec.EnableSsl)
	}

	if spec.SslConnections != nil {
		gatewayArgs.SslConnections = pulumi.IntPtr(int(*spec.SslConnections))
	}

	if spec.ResourceGroupId != "" {
		gatewayArgs.ResourceGroupId = pulumi.StringPtr(spec.ResourceGroupId)
	}

	gateway, err := vpn.NewGateway(ctx, spec.VpnGatewayName, gatewayArgs,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create VPN gateway %s", spec.VpnGatewayName)
	}

	connectionIds := pulumi.StringMap{}

	for _, conn := range spec.Connections {
		connId, err := vpnConnection(ctx, alicloudProvider, gateway, conn)
		if err != nil {
			return err
		}
		connectionIds[conn.Name] = connId
	}

	ctx.Export(OpVpnGatewayId, gateway.ID())
	ctx.Export(OpInternetIp, gateway.InternetIp)
	ctx.Export(OpSslVpnInternetIp, gateway.SslVpnInternetIp)
	ctx.Export(OpConnectionIds, connectionIds)

	return nil
}
