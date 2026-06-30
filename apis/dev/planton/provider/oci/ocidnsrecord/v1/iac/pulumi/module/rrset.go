package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func rrset(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciDnsRecord.Spec

	items := dns.RrsetItemArray{}
	for _, item := range spec.Items {
		items = append(items, &dns.RrsetItemArgs{
			Domain: pulumi.String(spec.Domain),
			Rtype:  pulumi.String(spec.Rtype),
			Rdata:  pulumi.String(item.Rdata),
			Ttl:    pulumi.Int(int(item.Ttl)),
		})
	}

	args := &dns.RrsetArgs{
		ZoneNameOrId: pulumi.String(spec.ZoneNameOrId.GetValue()),
		Domain:       pulumi.String(spec.Domain),
		Rtype:        pulumi.String(spec.Rtype),
		Items:        items,
	}

	if spec.ViewId != nil {
		args.ViewId = pulumi.String(spec.ViewId.GetValue())
	}

	_, err := dns.NewRrset(ctx, locals.ResourceName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create dns rrset")
	}

	return nil
}
