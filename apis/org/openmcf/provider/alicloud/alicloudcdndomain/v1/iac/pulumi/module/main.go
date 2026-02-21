package module

import (
	"github.com/pkg/errors"
	alicloudcdndomainv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudcdndomain/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudcdndomainv1.AliCloudCdnDomainStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudCdnDomain.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	args := &cdn.DomainNewArgs{
		DomainName: pulumi.String(spec.DomainName),
		CdnType:    pulumi.String(spec.CdnType),
		Sources:    buildSources(spec.Sources),
		Tags:       pulumi.ToStringMap(locals.Tags),
	}

	if spec.Scope != "" {
		args.Scope = pulumi.String(spec.Scope)
	}
	if spec.CheckUrl != "" {
		args.CheckUrl = pulumi.String(spec.CheckUrl)
	}
	if spec.ResourceGroupId != "" {
		args.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}
	if spec.CertificateConfig != nil {
		args.CertificateConfig = buildCertificateConfig(spec.CertificateConfig)
	}

	domain, err := cdn.NewDomainNew(ctx, spec.DomainName, args, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create CDN domain %s", spec.DomainName)
	}

	ctx.Export(OpDomainName, domain.DomainName)
	ctx.Export(OpCname, domain.Cname)
	ctx.Export(OpStatus, domain.Status)

	return nil
}

func buildSources(sources []*alicloudcdndomainv1.AliCloudCdnDomainSource) cdn.DomainNewSourceArray {
	var result cdn.DomainNewSourceArray
	for _, s := range sources {
		sourceArgs := cdn.DomainNewSourceArgs{
			Type:    pulumi.String(s.Type),
			Content: pulumi.String(s.Content),
		}
		if s.Port != 0 {
			sourceArgs.Port = pulumi.Int(int(s.Port))
		}
		if s.Priority != 0 {
			sourceArgs.Priority = pulumi.Int(int(s.Priority))
		}
		if s.Weight != 0 {
			sourceArgs.Weight = pulumi.Int(int(s.Weight))
		}
		result = append(result, sourceArgs)
	}
	return result
}

func buildCertificateConfig(cc *alicloudcdndomainv1.AliCloudCdnDomainCertificateConfig) cdn.DomainNewCertificateConfigPtrInput {
	args := &cdn.DomainNewCertificateConfigArgs{}

	if cc.CertName != "" {
		args.CertName = pulumi.String(cc.CertName)
	}
	if cc.CertType != "" {
		args.CertType = pulumi.String(cc.CertType)
	}
	if cc.CertId != "" {
		args.CertId = pulumi.String(cc.CertId)
	}
	if cc.CertRegion != "" {
		args.CertRegion = pulumi.String(cc.CertRegion)
	}
	if cc.ServerCertificate != "" {
		args.ServerCertificate = pulumi.String(cc.ServerCertificate)
	}
	if cc.PrivateKey != "" {
		args.PrivateKey = pulumi.String(cc.PrivateKey)
	}
	if cc.ServerCertificateStatus != "" {
		args.ServerCertificateStatus = pulumi.String(cc.ServerCertificateStatus)
	}

	return args
}
