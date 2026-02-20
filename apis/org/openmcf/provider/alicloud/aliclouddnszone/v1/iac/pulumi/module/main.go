package module

import (
	"github.com/pkg/errors"
	aliclouddnszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/aliclouddnszone/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *aliclouddnszonev1.AlicloudDnsZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudDnsZone.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	domain, err := dns.NewAlidnsDomain(ctx, spec.DomainName, &dns.AlidnsDomainArgs{
		DomainName:      pulumi.String(spec.DomainName),
		GroupId:         optionalString(spec.GroupId),
		Remark:          optionalString(spec.Remark),
		ResourceGroupId: optionalString(spec.ResourceGroupId),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS domain %s", spec.DomainName)
	}

	ctx.Export(OpDomainId, domain.DomainId)
	ctx.Export(OpDomainName, domain.DomainName)
	ctx.Export(OpDnsServers, domain.DnsServers)
	ctx.Export(OpGroupName, domain.GroupName)
	ctx.Export(OpPunyCode, domain.PunyCode)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
