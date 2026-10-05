package module

import (
	"sort"
	"strconv"

	awsbedrockagentcoreevaluationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoreevaluation/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluation
	Spec   *awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluationSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluationIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockAgentCoreEvaluation.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}

func svrsToStringArray(in []*foreignkeyv1.StringValueOrRef) pulumi.StringArray {
	out := pulumi.StringArray{}
	for _, ref := range in {
		out = append(out, pulumi.String(ref.GetValue()))
	}
	return out
}

func sortedEvaluators(in []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluator) []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluator {
	out := append([]*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreEvaluator{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedHarnesses(in []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreHarness) []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreHarness {
	out := append([]*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreHarness{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedOnlineConfigs(in []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreOnlineEvaluationConfig) []*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreOnlineEvaluationConfig {
	out := append([]*awsbedrockagentcoreevaluationv1alpha1.AwsBedrockAgentCoreOnlineEvaluationConfig{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
