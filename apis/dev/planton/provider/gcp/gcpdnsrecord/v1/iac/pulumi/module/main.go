package module

import (
	"github.com/pkg/errors"
	gcpdnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpdnsrecord/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpdnsrecordv1.GcpDnsRecordStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Create GCP provider using credentials from the input
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Create the DNS record set
	createdRecordSet, err := dns.NewRecordSet(ctx,
		locals.GcpDnsRecord.Metadata.Name,
		&dns.RecordSetArgs{
			Project:     pulumi.String(locals.ProjectId),
			ManagedZone: pulumi.String(locals.ManagedZone),
			Name:        pulumi.String(locals.Name),
			Type:        pulumi.String(locals.RecordType),
			Ttl:         pulumi.IntPtr(locals.TtlSeconds),
			Rrdatas:     pulumi.ToStringArray(locals.Values),
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS record %s", locals.Name)
	}

	// Export outputs
	ctx.Export(OpFqdn, createdRecordSet.Name)
	ctx.Export(OpRecordType, createdRecordSet.Type)
	ctx.Export(OpManagedZone, createdRecordSet.ManagedZone)
	ctx.Export(OpProjectId, createdRecordSet.Project)
	ctx.Export(OpTtlSeconds, createdRecordSet.Ttl)

	return nil
}
