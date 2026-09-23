package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	// pulumi-gcp files Google's Document AI resources under the
	// essentialcontacts package (its documentai package holds only the
	// warehouse schema); the types are the google_document_ai_processor and
	// google_document_ai_processor_default_version bridges.
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/essentialcontacts"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// processor creates the Document AI processor and, when the spec names one,
// the default version binding. Every processor argument is immutable.
func processor(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDocumentAiProcessor.Spec
	metadata := locals.GcpDocumentAiProcessor.Metadata

	// Enable the Document AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one processor must
	// never disable the API for everything else in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("documentai.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpdocai-documentai.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable documentai.googleapis.com api")
	}

	// Google requires a display name; the spec defaults it to metadata.name
	// -- identical to the Terraform module.
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = metadata.Name
	}

	args := &essentialcontacts.DocumentAiProcessorArgs{
		Location:    pulumi.String(spec.Location),
		Type:        pulumi.String(spec.Type),
		DisplayName: pulumi.String(displayName),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdProcessor, err := essentialcontacts.NewDocumentAiProcessor(ctx,
		metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create document ai processor")
	}

	// The default version: the spec carries the short id and the full path
	// is composed under this processor -- the Terraform module composes it
	// the same way. Destroy is a no-op in Google (there is no "unset").
	if spec.DefaultVersion != "" {
		processorId := createdProcessor.ID().ToStringOutput()
		_, err := essentialcontacts.NewDocumentAiProcessorDefaultVersion(ctx,
			metadata.Name+"-default-version",
			&essentialcontacts.DocumentAiProcessorDefaultVersionArgs{
				Processor: processorId,
				Version:   pulumi.Sprintf("%s/processorVersions/%s", processorId, spec.DefaultVersion),
			},
			pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to set the processor's default version")
		}
	}

	ctx.Export(OpName, createdProcessor.ID())
	ctx.Export(OpProcessorId, createdProcessor.Name)
	ctx.Export(OpLocation, createdProcessor.Location)
	ctx.Export(OpProcessEndpoint, pulumi.Sprintf("https://%s-documentai.googleapis.com/v1/%s:process",
		createdProcessor.Location, createdProcessor.ID()))
	return nil
}
