package module

import (
	aliclouddnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/aliclouddnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudDnsRecord *aliclouddnsrecordv1.AliCloudDnsRecord
}

func initializeLocals(ctx *pulumi.Context, stackInput *aliclouddnsrecordv1.AliCloudDnsRecordStackInput) *Locals {
	return &Locals{
		AliCloudDnsRecord: stackInput.Target,
	}
}
