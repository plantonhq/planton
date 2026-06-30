package module

import (
	awsrdsclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsrdscluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRdsCluster *awsrdsclusterv1.AwsRdsCluster
	Labels        map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsrdsclusterv1.AwsRdsClusterStackInput) *Locals {
	locals := &Locals{}

	locals.AwsRdsCluster = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsRdsCluster.Metadata.Org,
		"planton.org/environment":   locals.AwsRdsCluster.Metadata.Env,
		"planton.org/resource-kind": "AwsRdsCluster",
		"planton.org/resource-id":   locals.AwsRdsCluster.Metadata.Id,
	}

	return locals
}
