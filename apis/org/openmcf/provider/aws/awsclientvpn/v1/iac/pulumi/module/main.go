package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsclientvpnv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsclientvpn/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2clientvpn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources – entry‑point invoked by the OpenMCF engine
func Resources(ctx *pulumi.Context, stackInput *awsclientvpnv1.AwsClientVpnStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsClientVpn.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsClientVpn.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Authentication (only certificate‑based for now)
	authOptions := ec2clientvpn.EndpointAuthenticationOptionArray{
		&ec2clientvpn.EndpointAuthenticationOptionArgs{
			Type:                    pulumi.String("certificate-authentication"),
			RootCertificateChainArn: pulumi.String(locals.AwsClientVpn.Spec.ServerCertificateArn.GetValue()),
		},
	}

	// Connection‑log options
	logOptions := &ec2clientvpn.EndpointConnectionLogOptionsArgs{
		Enabled: pulumi.Bool(locals.AwsClientVpn.Spec.LogGroupName != ""),
	}
	if locals.AwsClientVpn.Spec.LogGroupName != "" {
		logOptions.CloudwatchLogGroup = pulumi.String(locals.AwsClientVpn.Spec.LogGroupName)
	}

	// VPN port (default 443)
	vpnPort := 443
	if locals.AwsClientVpn.Spec.VpnPort != nil && *locals.AwsClientVpn.Spec.VpnPort != 0 {
		vpnPort = int(*locals.AwsClientVpn.Spec.VpnPort)
	}

	// Transport protocol - get directly from enum (values match AWS API strings)
	proto := "tcp" // default to tcp
	if locals.AwsClientVpn.Spec.TransportProtocol != 0 {
		proto = locals.AwsClientVpn.Spec.TransportProtocol.String()
	}

	var securityGroupIds pulumi.StringArray

	for _, sg := range locals.AwsClientVpn.Spec.SecurityGroups {
		securityGroupIds = append(securityGroupIds, pulumi.String(sg.GetValue()))
	}

	// --------------------------------------------------- Endpoint resource
	createdClientVpnEndpoint, err := ec2clientvpn.NewEndpoint(ctx,
		locals.AwsClientVpn.Metadata.Name,
		&ec2clientvpn.EndpointArgs{
			VpcId:                 pulumi.String(locals.AwsClientVpn.Spec.VpcId.GetValue()),
			Description:           pulumi.String(locals.AwsClientVpn.Spec.Description),
			ServerCertificateArn:  pulumi.String(locals.AwsClientVpn.Spec.ServerCertificateArn.GetValue()),
			ClientCidrBlock:       pulumi.String(locals.AwsClientVpn.Spec.ClientCidrBlock),
			SplitTunnel:           pulumi.Bool(!locals.AwsClientVpn.Spec.DisableSplitTunnel),
			VpnPort:               pulumi.Int(vpnPort),
			TransportProtocol:     pulumi.String(proto),
			SecurityGroupIds:      securityGroupIds,
			AuthenticationOptions: authOptions,
			ConnectionLogOptions:  logOptions,
			DnsServers:            pulumi.ToStringArray(locals.AwsClientVpn.Spec.DnsServers),
			Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
				stringmaps.AddEntry(locals.AwsTags, "Name", locals.AwsClientVpn.Metadata.Name)),
		}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create endpoint")
	}
	ctx.Export(OpClientVpnEndpointId, createdClientVpnEndpoint.ID())
	ctx.Export(OpEndpointDnsName, createdClientVpnEndpoint.DnsName)

	// --------------------------------------- Subnet associations
	for _, subnetRef := range locals.AwsClientVpn.Spec.Subnets {
		subnetID := subnetRef.GetValue()

		createdAssociation, err := ec2clientvpn.NewNetworkAssociation(ctx,
			fmt.Sprintf("assoc-%s-%s", locals.AwsClientVpn.Metadata.Name, subnetID),
			&ec2clientvpn.NetworkAssociationArgs{
				ClientVpnEndpointId: createdClientVpnEndpoint.ID(),
				SubnetId:            pulumi.String(subnetID),
			}, pulumi.Provider(provider), pulumi.Parent(createdClientVpnEndpoint))
		if err != nil {
			return errors.Wrapf(err, "associate subnet %s", subnetID)
		}
		ctx.Export(fmt.Sprintf("%s.%s", OpSubnetAssociationIds, subnetID), createdAssociation.ID())
	}

	// --------------------------------------- Authorization rules
	for idx, cidr := range locals.AwsClientVpn.Spec.CidrAuthorizationRules {
		_, err = ec2clientvpn.NewAuthorizationRule(ctx,
			fmt.Sprintf("auth-rule-%d", idx),
			&ec2clientvpn.AuthorizationRuleArgs{
				ClientVpnEndpointId: createdClientVpnEndpoint.ID(),
				TargetNetworkCidr:   pulumi.String(cidr),
				AuthorizeAllGroups:  pulumi.Bool(true),
			}, pulumi.Provider(provider), pulumi.Parent(createdClientVpnEndpoint))
		if err != nil {
			return errors.Wrapf(err, "authorization rule %q", cidr)
		}
	}

	return nil
}
