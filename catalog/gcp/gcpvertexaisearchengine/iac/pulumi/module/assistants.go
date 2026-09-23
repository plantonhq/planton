package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// assistants creates one `google_discovery_engine_assistant` per
// spec.assistants[] entry, keyed by assistant_id -- the Gemini Enterprise
// assistant's policy, generation settings, and web grounding. Optional
// levers are sent only when set. Names are exported in manifest order.
func assistants(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	created *createdEngine) (pulumi.StringArray, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec
	names := pulumi.StringArray{}

	for _, assistant := range spec.Assistants {
		args := &discoveryengine.AssistantArgs{
			AssistantId:  pulumi.String(assistant.AssistantId),
			DisplayName:  pulumi.String(assistant.DisplayName),
			EngineId:     created.EngineId,
			Location:     pulumi.String(spec.Location),
			CollectionId: pulumi.String(locals.CollectionId),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if assistant.Description != "" {
			args.Description = pulumi.String(assistant.Description)
		}
		if assistant.WebGroundingType != "" {
			args.WebGroundingType = pulumi.String(assistant.WebGroundingType)
		}
		if policy := assistant.CustomerPolicy; policy != nil {
			policyArgs := &discoveryengine.AssistantCustomerPolicyArgs{}
			if len(policy.BannedPhrases) > 0 {
				phrases := discoveryengine.AssistantCustomerPolicyBannedPhraseArray{}
				for _, phrase := range policy.BannedPhrases {
					phraseArgs := &discoveryengine.AssistantCustomerPolicyBannedPhraseArgs{
						Phrase:           pulumi.String(phrase.Phrase),
						IgnoreDiacritics: pulumi.BoolPtr(phrase.IgnoreDiacritics),
					}
					if phrase.MatchType != "" {
						phraseArgs.MatchType = pulumi.String(phrase.MatchType)
					}
					phrases = append(phrases, phraseArgs)
				}
				policyArgs.BannedPhrases = phrases
			}
			if armor := policy.ModelArmorConfig; armor != nil {
				armorArgs := &discoveryengine.AssistantCustomerPolicyModelArmorConfigArgs{
					UserPromptTemplate: pulumi.String(armor.UserPromptTemplate),
					ResponseTemplate:   pulumi.String(armor.ResponseTemplate),
				}
				if armor.FailureMode != "" {
					armorArgs.FailureMode = pulumi.String(armor.FailureMode)
				}
				policyArgs.ModelArmorConfig = armorArgs
			}
			args.CustomerPolicy = policyArgs
		}
		if gen := assistant.GenerationConfig; gen != nil {
			genArgs := &discoveryengine.AssistantGenerationConfigArgs{}
			if gen.DefaultLanguage != "" {
				genArgs.DefaultLanguage = pulumi.String(gen.DefaultLanguage)
			}
			if gen.AdditionalSystemInstruction != "" {
				genArgs.SystemInstruction = &discoveryengine.AssistantGenerationConfigSystemInstructionArgs{
					AdditionalSystemInstruction: pulumi.String(gen.AdditionalSystemInstruction),
				}
			}
			args.GenerationConfig = genArgs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdAssistant, err := discoveryengine.NewAssistant(ctx,
			fmt.Sprintf("%s-%s", locals.GcpVertexAiSearchEngine.Metadata.Name, assistant.AssistantId), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(created.Resource))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create assistant %s", assistant.AssistantId)
		}
		names = append(names, createdAssistant.Name)
	}
	return names, nil
}
