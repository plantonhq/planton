package module

import (
	"strconv"
	"strings"

	awssqsqueuev1alpha1 "github.com/plantonhq/planton/catalog/aws/awssqsqueue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target    *awssqsqueuev1alpha1.AwsSqsQueue
	Spec      *awssqsqueuev1alpha1.AwsSqsQueueSpec
	AwsTags   map[string]string
	QueueName string // Derived queue name; includes `.fifo` suffix for FIFO queues.
}

func initializeLocals(ctx *pulumi.Context, in *awssqsqueuev1alpha1.AwsSqsQueueIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	// Derive the queue name. FIFO queues must end with `.fifo`.
	queueName := in.Target.Metadata.Name
	if in.Target.Spec.FifoQueue && !strings.HasSuffix(queueName, ".fifo") {
		queueName = queueName + ".fifo"
	}
	locals.QueueName = queueName

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSqsQueue.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
