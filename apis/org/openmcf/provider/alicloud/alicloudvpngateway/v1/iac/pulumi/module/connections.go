package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudvpngatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudvpngateway/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func vpnConnection(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	gateway *vpn.Gateway,
	conn *alicloudvpngatewayv1.AlicloudVpnConnection,
) (pulumi.StringOutput, error) {
	cgArgs := &vpn.CustomerGatewayArgs{
		CustomerGatewayName: pulumi.StringPtr(fmt.Sprintf("%s-cg", conn.Name)),
		IpAddress:           pulumi.String(conn.CustomerGatewayIp),
	}

	if conn.CustomerGatewayAsn != "" {
		cgArgs.Asn = pulumi.StringPtr(conn.CustomerGatewayAsn)
	}

	customerGateway, err := vpn.NewCustomerGateway(ctx, fmt.Sprintf("%s-cg", conn.Name), cgArgs,
		pulumi.Provider(provider),
	)
	if err != nil {
		return pulumi.StringOutput{}, errors.Wrapf(err, "failed to create customer gateway for connection %s", conn.Name)
	}

	connArgs := &vpn.ConnectionArgs{
		VpnGatewayId:      gateway.ID(),
		CustomerGatewayId: customerGateway.ID(),
		VpnConnectionName: pulumi.StringPtr(conn.Name),
		LocalSubnets:      pulumi.ToStringArray(conn.LocalSubnets),
		RemoteSubnets:     pulumi.ToStringArray(conn.RemoteSubnets),
	}

	if conn.EnableDpd != nil {
		connArgs.EnableDpd = pulumi.BoolPtr(*conn.EnableDpd)
	}

	if conn.EnableNatTraversal != nil {
		connArgs.EnableNatTraversal = pulumi.BoolPtr(*conn.EnableNatTraversal)
	}

	if conn.EffectImmediately != nil {
		connArgs.EffectImmediately = pulumi.BoolPtr(*conn.EffectImmediately)
	}

	if conn.IkeConfig != nil {
		connArgs.IkeConfig = buildIkeConfig(conn.IkeConfig)
	}

	if conn.IpsecConfig != nil {
		connArgs.IpsecConfig = buildIpsecConfig(conn.IpsecConfig)
	}

	if conn.HealthCheckConfig != nil {
		connArgs.HealthCheckConfig = buildHealthCheckConfig(conn.HealthCheckConfig)
	}

	vpnConn, err := vpn.NewConnection(ctx, conn.Name, connArgs,
		pulumi.Provider(provider),
		pulumi.Parent(gateway),
		pulumi.DependsOn([]pulumi.Resource{customerGateway}),
	)
	if err != nil {
		return pulumi.StringOutput{}, errors.Wrapf(err, "failed to create VPN connection %s", conn.Name)
	}

	return vpnConn.ID().ToStringOutput(), nil
}

func buildIkeConfig(cfg *alicloudvpngatewayv1.AlicloudIkeConfig) vpn.ConnectionIkeConfigPtrInput {
	ike := &vpn.ConnectionIkeConfigArgs{}

	if cfg.Psk != "" {
		ike.Psk = pulumi.StringPtr(cfg.Psk)
	}
	if cfg.IkeVersion != nil {
		ike.IkeVersion = pulumi.StringPtr(*cfg.IkeVersion)
	}
	if cfg.IkeMode != nil {
		ike.IkeMode = pulumi.StringPtr(*cfg.IkeMode)
	}
	if cfg.IkeEncAlg != nil {
		ike.IkeEncAlg = pulumi.StringPtr(*cfg.IkeEncAlg)
	}
	if cfg.IkeAuthAlg != nil {
		ike.IkeAuthAlg = pulumi.StringPtr(*cfg.IkeAuthAlg)
	}
	if cfg.IkePfs != nil {
		ike.IkePfs = pulumi.StringPtr(*cfg.IkePfs)
	}
	if cfg.IkeLifetime != nil {
		ike.IkeLifetime = pulumi.IntPtr(int(*cfg.IkeLifetime))
	}

	return ike
}

func buildIpsecConfig(cfg *alicloudvpngatewayv1.AlicloudIpsecConfig) vpn.ConnectionIpsecConfigPtrInput {
	ipsec := &vpn.ConnectionIpsecConfigArgs{}

	if cfg.IpsecEncAlg != nil {
		ipsec.IpsecEncAlg = pulumi.StringPtr(*cfg.IpsecEncAlg)
	}
	if cfg.IpsecAuthAlg != nil {
		ipsec.IpsecAuthAlg = pulumi.StringPtr(*cfg.IpsecAuthAlg)
	}
	if cfg.IpsecPfs != nil {
		ipsec.IpsecPfs = pulumi.StringPtr(*cfg.IpsecPfs)
	}
	if cfg.IpsecLifetime != nil {
		ipsec.IpsecLifetime = pulumi.IntPtr(int(*cfg.IpsecLifetime))
	}

	return ipsec
}

func buildHealthCheckConfig(cfg *alicloudvpngatewayv1.AlicloudVpnHealthCheckConfig) vpn.ConnectionHealthCheckConfigPtrInput {
	hc := &vpn.ConnectionHealthCheckConfigArgs{}

	if cfg.Enable != nil {
		hc.Enable = pulumi.BoolPtr(*cfg.Enable)
	}
	if cfg.Sip != "" {
		hc.Sip = pulumi.StringPtr(cfg.Sip)
	}
	if cfg.Dip != "" {
		hc.Dip = pulumi.StringPtr(cfg.Dip)
	}
	if cfg.Interval != nil {
		hc.Interval = pulumi.IntPtr(int(*cfg.Interval))
	}
	if cfg.Retry != nil {
		hc.Retry = pulumi.IntPtr(int(*cfg.Retry))
	}

	return hc
}
