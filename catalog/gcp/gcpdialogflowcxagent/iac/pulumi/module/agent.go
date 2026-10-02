package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// agent enables the Dialogflow API and creates the agent itself; Google
// creates the default start flow and the default playbook with it. The
// folded webhooks, tools, versions, environments, and generative settings
// are created by their own functions, parented to the agent returned here.
func agent(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*diagflow.CxAgent, error) {
	spec := locals.GcpDialogflowCxAgent.Spec
	resourceName := locals.GcpDialogflowCxAgent.Metadata.Name

	// Enable the Dialogflow API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one agent must
	// never disable the API for everything else in the project.
	dialogflowApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("dialogflow.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		dialogflowApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdDialogflowApi, err := projects.NewService(ctx,
		"gcpdfcx-dialogflow.googleapis.com", dialogflowApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable dialogflow.googleapis.com api")
	}

	args := &diagflow.CxAgentArgs{
		Location:               pulumi.String(spec.Location),
		DisplayName:            pulumi.String(locals.DisplayName),
		DefaultLanguageCode:    pulumi.String(spec.DefaultLanguageCode),
		TimeZone:               pulumi.String(spec.TimeZone),
		Description:            optionalString(spec.Description),
		AvatarUri:              optionalString(spec.AvatarUri),
		SupportedLanguageCodes: optionalStringArray(spec.SupportedLanguageCodes),
		SecuritySettings:       optionalString(spec.SecuritySettings.GetValue()),
		// Optional booleans are sent only when true (Google's default is
		// false) -- the Terraform module's rule.
		EnableMultiLanguageTraining: optionalTrue(spec.EnableMultiLanguageTraining),
		EnableSpellCorrection:       optionalTrue(spec.EnableSpellCorrection),
		Locked:                      optionalTrue(spec.Locked),
		// Client-side: also delete the linked Vertex AI Search engine on
		// destroy.
		DeleteChatEngineOnDestroy: optionalTrue(spec.DeleteChatEngineOnDestroy),
		// Engine-side destroy stance, fanned to every folded child.
		DeletionPolicy: deletionPolicy(locals),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.StartWithDefaultPlaybook {
		args.StartPlaybook = pulumi.String(defaultPlaybook)
	}

	// Optional and computed: an unset block leaves Google's values alone,
	// and each nested block is sent only when declared.
	if advanced := spec.AdvancedSettings; advanced != nil {
		advancedArgs := &diagflow.CxAgentAdvancedSettingsArgs{}
		if destination := advanced.AudioExportGcsDestination; destination != nil {
			advancedArgs.AudioExportGcsDestination = &diagflow.CxAgentAdvancedSettingsAudioExportGcsDestinationArgs{
				Uri: optionalString(destination.Uri),
			}
		}
		if dtmf := advanced.DtmfSettings; dtmf != nil {
			dtmfArgs := &diagflow.CxAgentAdvancedSettingsDtmfSettingsArgs{
				Enabled:     optionalTrue(dtmf.Enabled),
				FinishDigit: optionalString(dtmf.FinishDigit),
			}
			if dtmf.MaxDigits > 0 {
				dtmfArgs.MaxDigits = pulumi.Int(int(dtmf.MaxDigits))
			}
			advancedArgs.DtmfSettings = dtmfArgs
		}
		if logging := advanced.LoggingSettings; logging != nil {
			advancedArgs.LoggingSettings = &diagflow.CxAgentAdvancedSettingsLoggingSettingsArgs{
				EnableConsentBasedRedaction: optionalTrue(logging.EnableConsentBasedRedaction),
				EnableInteractionLogging:    optionalTrue(logging.EnableInteractionLogging),
				EnableStackdriverLogging:    optionalTrue(logging.EnableStackdriverLogging),
			}
		}
		if speech := advanced.SpeechSettings; speech != nil {
			speechArgs := &diagflow.CxAgentAdvancedSettingsSpeechSettingsArgs{
				Models:                     optionalStringMap(speech.Models),
				NoSpeechTimeout:            optionalString(speech.NoSpeechTimeout),
				UseTimeoutBasedEndpointing: optionalTrue(speech.UseTimeoutBasedEndpointing),
			}
			if speech.EndpointerSensitivity > 0 {
				speechArgs.EndpointerSensitivity = pulumi.Int(int(speech.EndpointerSensitivity))
			}
			advancedArgs.SpeechSettings = speechArgs
		}
		args.AdvancedSettings = advancedArgs
	}

	// The spec lifts these four blocks' single leaves.
	if spec.EnableAnswerFeedback {
		args.AnswerFeedbackSettings = &diagflow.CxAgentAnswerFeedbackSettingsArgs{
			EnableAnswerFeedback: pulumi.Bool(true),
		}
	}
	if spec.DefaultEndUserMetadata != "" {
		args.PersonalizationSettings = &diagflow.CxAgentPersonalizationSettingsArgs{
			DefaultEndUserMetadata: pulumi.String(spec.DefaultEndUserMetadata),
		}
	}
	if spec.EnableSpeechAdaptation {
		args.SpeechToTextSettings = &diagflow.CxAgentSpeechToTextSettingsArgs{
			EnableSpeechAdaptation: pulumi.Bool(true),
		}
	}
	if spec.SynthesizeSpeechConfigs != "" {
		args.TextToSpeechSettings = &diagflow.CxAgentTextToSpeechSettingsArgs{
			SynthesizeSpeechConfigs: pulumi.String(spec.SynthesizeSpeechConfigs),
		}
	}

	if certificate := spec.ClientCertificateSettings; certificate != nil {
		args.ClientCertificateSettings = &diagflow.CxAgentClientCertificateSettingsArgs{
			SslCertificate: pulumi.String(certificate.SslCertificate),
			PrivateKey:     pulumi.String(certificate.PrivateKey),
			Passphrase:     optionalString(certificate.Passphrase),
		}
	}
	if genAppBuilder := spec.GenAppBuilderSettings; genAppBuilder != nil {
		args.GenAppBuilderSettings = &diagflow.CxAgentGenAppBuilderSettingsArgs{
			Engine: pulumi.String(genAppBuilder.Engine.GetValue()),
		}
	}
	if git := spec.GitIntegrationSettings; git != nil {
		gitArgs := &diagflow.CxAgentGitIntegrationSettingsArgs{}
		if github := git.GithubSettings; github != nil {
			gitArgs.GithubSettings = &diagflow.CxAgentGitIntegrationSettingsGithubSettingsArgs{
				DisplayName:    optionalString(github.DisplayName),
				RepositoryUri:  optionalString(github.RepositoryUri),
				TrackingBranch: optionalString(github.TrackingBranch),
				Branches:       optionalStringArray(github.Branches),
				AccessToken:    optionalSecret(github.AccessToken),
			}
		}
		args.GitIntegrationSettings = gitArgs
	}

	createdAgent, err := diagflow.NewCxAgent(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdDialogflowApi}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dialogflow cx agent")
	}

	ctx.Export(OpName, createdAgent.ID().ToStringOutput())
	ctx.Export(OpAgentId, createdAgent.Name)
	ctx.Export(OpLocation, createdAgent.Location)
	ctx.Export(OpStartFlow, createdAgent.StartFlow)
	return createdAgent, nil
}
