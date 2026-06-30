package module

import (
	awsredshiftclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsredshiftcluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRedshiftCluster *awsredshiftclusterv1.AwsRedshiftCluster
	Labels             map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsredshiftclusterv1.AwsRedshiftClusterStackInput) *Locals {
	locals := &Locals{}

	locals.AwsRedshiftCluster = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsRedshiftCluster.Metadata.Org,
		"planton.org/environment":   locals.AwsRedshiftCluster.Metadata.Env,
		"planton.org/resource-kind": "AwsRedshiftCluster",
		"planton.org/resource-id":   locals.AwsRedshiftCluster.Metadata.Id,
	}

	return locals
}
