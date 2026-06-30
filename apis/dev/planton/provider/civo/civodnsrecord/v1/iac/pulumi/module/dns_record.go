package module

import (
	"strings"

	"github.com/pkg/errors"
	civodnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodnsrecord/v1"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsRecord provisions the Civo DNS record and exports outputs.
func dnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.DnsDomainRecord, error) {
	spec := locals.CivoDnsRecord.Spec

	// Determine TTL (default to 3600 if not specified).
	ttl := 3600
	if spec.Ttl > 0 {
		ttl = int(spec.Ttl)
	}

	// Build the record arguments.
	recordArgs := &civo.DnsDomainRecordArgs{
		DomainId: pulumi.String(locals.ZoneId),
		Name:     pulumi.String(spec.Name),
		Type:     pulumi.String(spec.Type.String()),
		Value:    pulumi.String(spec.Value),
		Ttl:      pulumi.Int(ttl),
	}

	// Set priority for MX/SRV records.
	if spec.Type == civodnsrecordv1.CivoDnsRecordSpec_MX ||
		spec.Type == civodnsrecordv1.CivoDnsRecordSpec_SRV {
		recordArgs.Priority = pulumi.Int(int(spec.Priority))
	}

	// Create the DNS record.
	createdRecord, err := civo.NewDnsDomainRecord(
		ctx,
		strings.ToLower(locals.CivoDnsRecord.Metadata.Name),
		recordArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo dns record")
	}

	// Export required outputs.
	ctx.Export(OpRecordId, createdRecord.ID())
	ctx.Export(OpHostname, createdRecord.Name)
	ctx.Export(OpRecordType, pulumi.String(spec.Type.String()))
	ctx.Export(OpAccountId, createdRecord.AccountId)

	return createdRecord, nil
}
