package module

import (
	ocidnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocidnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciDnsRecord *ocidnsrecordv1.OciDnsRecord
	ResourceName string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocidnsrecordv1.OciDnsRecordStackInput) *Locals {
	return &Locals{
		OciDnsRecord: stackInput.Target,
		ResourceName: stackInput.Target.Metadata.Name,
	}
}
