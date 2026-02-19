package module

import (
	"fmt"

	"github.com/pkg/errors"
	aliclouddnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/aliclouddnsrecord/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *aliclouddnsrecordv1.AlicloudDnsRecordStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudDnsRecord.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	resourceName := fmt.Sprintf("%s.%s", spec.Rr, spec.DomainName)

	args := &dns.AlidnsRecordArgs{
		DomainName: pulumi.String(spec.DomainName),
		Rr:         pulumi.String(spec.Rr),
		Type:       pulumi.String(spec.Type),
		Value:      pulumi.String(spec.Value),
	}

	if spec.Ttl > 0 {
		args.Ttl = pulumi.Int(int(spec.Ttl))
	}

	if spec.Priority > 0 {
		args.Priority = pulumi.Int(int(spec.Priority))
	}

	if spec.Line != "" {
		args.Line = pulumi.String(spec.Line)
	}

	if spec.Status != "" {
		args.Status = pulumi.String(spec.Status)
	}

	if spec.Remark != "" {
		args.Remark = pulumi.String(spec.Remark)
	}

	record, err := dns.NewAlidnsRecord(ctx, resourceName, args, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS record %s", resourceName)
	}

	ctx.Export(OpRecordId, record.ID())

	return nil
}
