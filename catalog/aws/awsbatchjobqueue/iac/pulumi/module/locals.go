package module

import (
	"strconv"

	awsbatchjobqueuev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchjobqueue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsBatchJobQueue *awsbatchjobqueuev1alpha1.AwsBatchJobQueue
	AwsTags          map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsbatchjobqueuev1alpha1.AwsBatchJobQueueIacInput) *Locals {
	locals := &Locals{}
	locals.AwsBatchJobQueue = iacInput.Target

	// Resource-identity tags follow the catalog convention.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsBatchJobQueue.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsBatchJobQueue.Metadata.Org,
		awstagkeys.Environment:  locals.AwsBatchJobQueue.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBatchJobQueue.String(),
		awstagkeys.ResourceId:   locals.AwsBatchJobQueue.Metadata.Id,
	}

	return locals
}
