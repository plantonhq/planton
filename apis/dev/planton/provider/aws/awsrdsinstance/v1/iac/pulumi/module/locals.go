package module

import (
	awsrdsinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsrdsinstance/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals captures convenient references and computed labels for the module.
type Locals struct {
	AwsRdsInstance *awsrdsinstancev1.AwsRdsInstance
	Labels         map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsrdsinstancev1.AwsRdsInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AwsRdsInstance = in.Target

	// Use Metadata.Id if available, otherwise fall back to Metadata.Name
	resourceId := locals.AwsRdsInstance.Metadata.Id
	if resourceId == "" {
		resourceId = locals.AwsRdsInstance.Metadata.Name
	}

	locals.Labels = map[string]string{
		"planton.org/resource":      "true",
		"planton.org/organization":  locals.AwsRdsInstance.Metadata.Org,
		"planton.org/environment":   locals.AwsRdsInstance.Metadata.Env,
		"planton.org/resource-kind": "AwsRdsInstance",
		"planton.org/resource-id":   resourceId,
	}
	return locals
}
