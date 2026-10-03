package module

import (
	"encoding/json"
	"strconv"

	awseventbridgerulev1alpha1 "github.com/plantonhq/planton/catalog/aws/awseventbridgerule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/types/known/structpb"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target  *awseventbridgerulev1alpha1.AwsEventBridgeRule
	Spec    *awseventbridgerulev1alpha1.AwsEventBridgeRuleSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awseventbridgerulev1alpha1.AwsEventBridgeRuleIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEventBridgeRule.String(),
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
