package module

import (
	gcpdnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpdnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpDnsRecord *gcpdnsrecordv1.GcpDnsRecord
	ProjectId    string
	ManagedZone  string
	RecordType   string
	Name         string
	Values       []string
	TtlSeconds   int
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcpdnsrecordv1.GcpDnsRecordStackInput) *Locals {
	locals := &Locals{}

	locals.GcpDnsRecord = stackInput.Target

	target := stackInput.Target

	// Extract project ID from StringValueOrRef
	locals.ProjectId = target.Spec.ProjectId.GetValue()

	// Extract managed zone from StringValueOrRef
	locals.ManagedZone = target.Spec.ManagedZone.GetValue()
	locals.RecordType = target.Spec.Type.String()
	locals.Name = target.Spec.Name
	locals.Values = target.Spec.Values

	// Get TTL with default of 300 if not set
	locals.TtlSeconds = int(target.Spec.GetTtlSeconds())
	if locals.TtlSeconds == 0 {
		locals.TtlSeconds = 300
	}

	return locals
}
