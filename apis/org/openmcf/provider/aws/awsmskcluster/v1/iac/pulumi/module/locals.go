package module

import (
	awsmskclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmskcluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsMskCluster *awsmskclusterv1.AwsMskCluster
	Labels        map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsmskclusterv1.AwsMskClusterStackInput) *Locals {
	locals := &Locals{}

	locals.AwsMskCluster = in.Target

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsMskCluster.Metadata.Org,
		"planton.org/environment":   locals.AwsMskCluster.Metadata.Env,
		"planton.org/resource-kind": "AwsMskCluster",
		"planton.org/resource-id":   locals.AwsMskCluster.Metadata.Id,
	}

	return locals
}
