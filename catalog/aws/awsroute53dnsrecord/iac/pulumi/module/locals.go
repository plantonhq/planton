package module

import (
	awsroute53dnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53dnsrecord/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRoute53DnsRecord *awsroute53dnsrecordv1alpha1.AwsRoute53DnsRecord
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsroute53dnsrecordv1alpha1.AwsRoute53DnsRecordIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRoute53DnsRecord = iacInput.Target
	return locals
}
