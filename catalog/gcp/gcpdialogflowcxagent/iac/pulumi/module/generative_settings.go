package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// generativeSettings creates the agent's generative settings for each
// declared language, keyed by language. Google keeps one per agent and
// language: create is a PATCH that overwrites Google's defaults, and
// destroy only stops managing it (the resource has no deletion_policy).
func generativeSettings(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdAgent *diagflow.CxAgent) error {
	resourceName := locals.GcpDialogflowCxAgent.Metadata.Name

	names := pulumi.StringArray{}
	for _, settings := range locals.GcpDialogflowCxAgent.Spec.GenerativeSettings {
		args := &diagflow.CxGenerativeSettingsArgs{
			Parent:       createdAgent.ID().ToStringOutput(),
			LanguageCode: pulumi.String(settings.LanguageCode),
		}
		if fallback := settings.FallbackSettings; fallback != nil {
			fallbackArgs := &diagflow.CxGenerativeSettingsFallbackSettingsArgs{
				SelectedPrompt: optionalString(fallback.SelectedPrompt),
			}
			if len(fallback.PromptTemplates) > 0 {
				templates := diagflow.CxGenerativeSettingsFallbackSettingsPromptTemplateArray{}
				for _, template := range fallback.PromptTemplates {
					templates = append(templates, &diagflow.CxGenerativeSettingsFallbackSettingsPromptTemplateArgs{
						DisplayName: optionalString(template.DisplayName),
						Frozen:      optionalTrue(template.Frozen),
						PromptText:  optionalString(template.PromptText),
					})
				}
				fallbackArgs.PromptTemplates = templates
			}
			args.FallbackSettings = fallbackArgs
		}
		if safety := settings.GenerativeSafetySettings; safety != nil {
			safetyArgs := &diagflow.CxGenerativeSettingsGenerativeSafetySettingsArgs{
				DefaultBannedPhraseMatchStrategy: optionalString(safety.DefaultBannedPhraseMatchStrategy),
			}
			if len(safety.BannedPhrases) > 0 {
				phrases := diagflow.CxGenerativeSettingsGenerativeSafetySettingsBannedPhraseArray{}
				for _, phrase := range safety.BannedPhrases {
					phrases = append(phrases, &diagflow.CxGenerativeSettingsGenerativeSafetySettingsBannedPhraseArgs{
						LanguageCode: pulumi.String(phrase.LanguageCode),
						Text:         pulumi.String(phrase.Text),
					})
				}
				safetyArgs.BannedPhrases = phrases
			}
			args.GenerativeSafetySettings = safetyArgs
		}
		if knowledge := settings.KnowledgeConnectorSettings; knowledge != nil {
			args.KnowledgeConnectorSettings = &diagflow.CxGenerativeSettingsKnowledgeConnectorSettingsArgs{
				Agent:                    optionalString(knowledge.Agent),
				AgentIdentity:            optionalString(knowledge.AgentIdentity),
				AgentScope:               optionalString(knowledge.AgentScope),
				Business:                 optionalString(knowledge.Business),
				BusinessDescription:      optionalString(knowledge.BusinessDescription),
				DisableDataStoreFallback: optionalTrue(knowledge.DisableDataStoreFallback),
			}
		}
		if model := settings.LlmModelSettings; model != nil {
			args.LlmModelSettings = &diagflow.CxGenerativeSettingsLlmModelSettingsArgs{
				Model:      optionalString(model.Model),
				PromptText: optionalString(model.PromptText),
			}
		}

		created, err := diagflow.NewCxGenerativeSettings(ctx,
			fmt.Sprintf("%s-generative-settings-%s", resourceName, settings.LanguageCode), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdAgent))
		if err != nil {
			return errors.Wrapf(err, "failed to create generative settings for %s", settings.LanguageCode)
		}
		names = append(names, created.ID().ToStringOutput())
	}

	ctx.Export(OpGenerativeSettingsNames, names.ToStringArrayOutput())
	return nil
}
