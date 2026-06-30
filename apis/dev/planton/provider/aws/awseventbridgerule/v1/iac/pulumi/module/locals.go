package module

import (
	"encoding/json"
	"strconv"

	awseventbridgerulev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awseventbridgerule/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/types/known/structpb"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awseventbridgerulev1.AwsEventBridgeRule
	Spec    *awseventbridgerulev1.AwsEventBridgeRuleSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awseventbridgerulev1.AwsEventBridgeRuleStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsEventBridgeRule.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}

// serializeStruct converts a google.protobuf.Struct to a JSON string.
func serializeStruct(s *structpb.Struct) (string, error) {
	if s == nil {
		return "", nil
	}
	m := s.AsMap()
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
