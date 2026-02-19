package module

import (
	aliclouddnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/aliclouddnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudDnsRecord *aliclouddnsrecordv1.AlicloudDnsRecord
}

func initializeLocals(ctx *pulumi.Context, stackInput *aliclouddnsrecordv1.AlicloudDnsRecordStackInput) *Locals {
	return &Locals{
		AlicloudDnsRecord: stackInput.Target,
	}
}
