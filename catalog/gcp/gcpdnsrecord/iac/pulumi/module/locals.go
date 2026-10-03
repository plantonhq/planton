package module

import (
	gcpdnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdnsrecord/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpDnsRecord *gcpdnsrecordv1alpha1.GcpDnsRecord

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	ManagedZone string
	RecordType  string
	Name        string
	Values      []string
	TtlSeconds  int
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpdnsrecordv1alpha1.GcpDnsRecordIacInput) *Locals {
	locals := &Locals{}

	locals.GcpDnsRecord = iacInput.Target

	target := iacInput.Target

	locals.ProjectId = target.Spec.ProjectId.GetValue()
	locals.ManagedZone = target.Spec.ManagedZone.GetValue()
	locals.RecordType = target.Spec.Type
	locals.Name = target.Spec.Name.GetValue()
	for _, v := range target.Spec.Values {
		locals.Values = append(locals.Values, v.GetValue())
	}

	// TTL defaults to 300 when unset. An explicit 0 is preserved (optional
	// field presence distinguishes "unset" from "no caching").
	if target.Spec.TtlSeconds != nil {
		locals.TtlSeconds = int(target.Spec.GetTtlSeconds())
	} else {
		locals.TtlSeconds = 300
	}

	return locals
}
