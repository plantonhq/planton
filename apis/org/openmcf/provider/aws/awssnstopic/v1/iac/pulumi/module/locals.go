package module

import (
	"strconv"
	"strings"

	awssnstopicv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssnstopic/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target    *awssnstopicv1.AwsSnsTopic
	Spec      *awssnstopicv1.AwsSnsTopicSpec
	AwsTags   map[string]string
	TopicName string // Derived topic name; includes `.fifo` suffix for FIFO topics.
}

func initializeLocals(ctx *pulumi.Context, in *awssnstopicv1.AwsSnsTopicStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	// Derive the topic name. FIFO topics must end with `.fifo`.
	topicName := in.Target.Metadata.Name
	if in.Target.Spec.FifoTopic && !strings.HasSuffix(topicName, ".fifo") {
		topicName = topicName + ".fifo"
	}
	locals.TopicName = topicName

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsSnsTopic.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
