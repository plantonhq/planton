package module

import (
	"strconv"
	"strings"

	awssqsqueuev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssqsqueue/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target    *awssqsqueuev1.AwsSqsQueue
	Spec      *awssqsqueuev1.AwsSqsQueueSpec
	AwsTags   map[string]string
	QueueName string // Derived queue name; includes `.fifo` suffix for FIFO queues.
}

func initializeLocals(ctx *pulumi.Context, in *awssqsqueuev1.AwsSqsQueueStackInput) *Locals {
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
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsSqsQueue.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
