package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// securitySettings creates the Dialogflow CX security settings -- the
// redaction, retention, audio-export, and Insights-export policy agents in
// the same project and location reference by name.
func securitySettings(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDialogflowCxSecuritySettings.Spec
	resourceName := locals.GcpDialogflowCxSecuritySettings.Metadata.Name

	// Enable the Dialogflow API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one set of security
	// settings must never disable the API for every agent in the project.
	dialogflowApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("dialogflow.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		dialogflowApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdDialogflowApi, err := projects.NewService(ctx,
		"gcpdfcxs-dialogflow.googleapis.com", dialogflowApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable dialogflow.googleapis.com api")
	}

	args := &diagflow.CxSecuritySettingsArgs{
		Location:    pulumi.String(spec.Location),
		DisplayName: pulumi.String(locals.DisplayName),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.RedactionStrategy != "" {
		args.RedactionStrategy = pulumi.String(spec.RedactionStrategy)
	}
	if spec.RedactionScope != "" {
		args.RedactionScope = pulumi.String(spec.RedactionScope)
	}
	if spec.InspectTemplate != "" {
		args.InspectTemplate = pulumi.String(spec.InspectTemplate)
	}
	if spec.DeidentifyTemplate != "" {
		args.DeidentifyTemplate = pulumi.String(spec.DeidentifyTemplate)
	}
	if len(spec.PurgeDataTypes) > 0 {
		args.PurgeDataTypes = pulumi.ToStringArray(spec.PurgeDataTypes)
	}
	if spec.RetentionStrategy != "" {
		args.RetentionStrategy = pulumi.String(spec.RetentionStrategy)
	}
	// 0 means Google's default TTL, which is also what an unset window
	// means; send the window only when one is declared.
	if spec.RetentionWindowDays > 0 {
		args.RetentionWindowDays = pulumi.Int(int(spec.RetentionWindowDays))
	}

	// Setting gcs_bucket makes Google grant the Dialogflow service agent
	// Storage Object Creator on the bucket.
	if audio := spec.AudioExportSettings; audio != nil {
		audioArgs := &diagflow.CxSecuritySettingsAudioExportSettingsArgs{}
		if audio.GcsBucket.GetValue() != "" {
			audioArgs.GcsBucket = pulumi.String(audio.GcsBucket.GetValue())
		}
		if audio.AudioExportPattern != "" {
			audioArgs.AudioExportPattern = pulumi.String(audio.AudioExportPattern)
		}
		if audio.AudioFormat != "" {
			audioArgs.AudioFormat = pulumi.String(audio.AudioFormat)
		}
		if audio.EnableAudioRedaction {
			audioArgs.EnableAudioRedaction = pulumi.Bool(true)
		}
		args.AudioExportSettings = audioArgs
	}

	// The spec lifts the block's one required flag; the block is sent only
	// when export is on -- the Terraform module's rule.
	if spec.EnableInsightsExport {
		args.InsightsExportSettings = &diagflow.CxSecuritySettingsInsightsExportSettingsArgs{
			EnableInsightsExport: pulumi.Bool(true),
		}
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := diagflow.NewCxSecuritySettings(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdDialogflowApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx security settings")
	}

	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpSecuritySettingsId, created.Name)
	ctx.Export(OpLocation, created.Location)
	return nil
}
