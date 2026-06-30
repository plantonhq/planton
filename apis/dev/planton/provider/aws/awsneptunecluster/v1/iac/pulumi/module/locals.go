package module

import (
	awsneptuneclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsneptunecluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsNeptuneCluster *awsneptuneclusterv1.AwsNeptuneCluster
	Labels            map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsneptuneclusterv1.AwsNeptuneClusterStackInput) *Locals {
	locals := &Locals{}

	locals.AwsNeptuneCluster = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsNeptuneCluster.Metadata.Org,
		"planton.org/environment":   locals.AwsNeptuneCluster.Metadata.Env,
		"planton.org/resource-kind": "AwsNeptuneCluster",
		"planton.org/resource-id":   locals.AwsNeptuneCluster.Metadata.Id,
	}

	return locals
}
