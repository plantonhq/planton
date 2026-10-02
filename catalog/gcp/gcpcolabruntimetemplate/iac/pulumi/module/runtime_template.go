package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	gcpcolabruntimetemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabruntimetemplate/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/colab"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// runtimeTemplate creates the Colab Enterprise runtime template. The display
// name, the key, and the software update in place; everything else (labels
// included) is fixed at creation. Optional+Computed settings are sent only
// when set -- the Terraform module's posture.
func runtimeTemplate(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpColabRuntimeTemplate.Spec
	metadata := locals.GcpColabRuntimeTemplate.Metadata

	// Enable the Vertex AI API (Colab Enterprise's API) first so a fresh
	// project works on the first deploy. DisableOnDestroy stays false.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpcolt-aiplatform.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	// The template id and display name default to metadata.name --
	// identical to the Terraform module.
	templateId := spec.RuntimeTemplateId
	if templateId == "" {
		templateId = metadata.Name
	}
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = metadata.Name
	}

	args := &colab.RuntimeTemplateArgs{
		Location:    pulumi.String(spec.Location),
		Name:        pulumi.String(templateId),
		DisplayName: pulumi.String(displayName),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if len(spec.NetworkTags) > 0 {
		args.NetworkTags = pulumi.ToStringArray(spec.NetworkTags)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	if m := spec.MachineSpec; m != nil {
		machine := &colab.RuntimeTemplateMachineSpecArgs{}
		if m.MachineType != "" {
			machine.MachineType = pulumi.String(m.MachineType)
		}
		if m.AcceleratorType != "" {
			machine.AcceleratorType = pulumi.String(m.AcceleratorType)
		}
		if m.AcceleratorCount != 0 {
			machine.AcceleratorCount = pulumi.Int(int(m.AcceleratorCount))
		}
		args.MachineSpec = machine
	}
	// Google types the disk size as a decimal string.
	if d := spec.DataPersistentDiskSpec; d != nil {
		disk := &colab.RuntimeTemplateDataPersistentDiskSpecArgs{}
		if d.DiskType != "" {
			disk.DiskType = pulumi.String(d.DiskType)
		}
		if d.DiskSizeGb != 0 {
			disk.DiskSizeGb = pulumi.String(strconv.FormatInt(d.DiskSizeGb, 10))
		}
		args.DataPersistentDiskSpec = disk
	}
	if n := spec.NetworkSpec; n != nil {
		network := &colab.RuntimeTemplateNetworkSpecArgs{
			EnableInternetAccess: pulumi.Bool(n.EnableInternetAccess),
		}
		if n.Network.GetValue() != "" {
			network.Network = pulumi.String(n.Network.GetValue())
		}
		// A GcpSubnetwork reference arrives as a compute self-link; Colab
		// takes the relative path -- the same trim as the Terraform module.
		if n.Subnetwork.GetValue() != "" {
			network.Subnetwork = pulumi.String(strings.TrimPrefix(n.Subnetwork.GetValue(), "https://www.googleapis.com/compute/v1/"))
		}
		args.NetworkSpec = network
	}
	// The spec lifts Google's one-leaf wrappers; each block is sent only when
	// its field is set.
	if spec.IdleTimeout != "" {
		args.IdleShutdownConfig = &colab.RuntimeTemplateIdleShutdownConfigArgs{IdleTimeout: pulumi.String(spec.IdleTimeout)}
	}
	if spec.EucDisabled {
		args.EucConfig = &colab.RuntimeTemplateEucConfigArgs{EucDisabled: pulumi.Bool(true)}
	}
	if spec.EnableSecureBoot {
		args.ShieldedVmConfig = &colab.RuntimeTemplateShieldedVmConfigArgs{EnableSecureBoot: pulumi.Bool(true)}
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &colab.RuntimeTemplateEncryptionSpecArgs{KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue())}
	}
	if s := spec.SoftwareConfig; s != nil {
		args.SoftwareConfig = buildSoftwareConfig(s)
	}

	createdTemplate, err := colab.NewRuntimeTemplate(ctx, metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create colab runtime template")
	}

	ctx.Export(OpName, createdTemplate.ID())
	ctx.Export(OpRuntimeTemplateId, createdTemplate.Name)
	ctx.Export(OpLocation, createdTemplate.Location)
	return nil
}

func buildSoftwareConfig(s *gcpcolabruntimetemplatev1alpha1.GcpColabRuntimeTemplateSoftwareConfig) *colab.RuntimeTemplateSoftwareConfigArgs {
	args := &colab.RuntimeTemplateSoftwareConfigArgs{}
	if len(s.Env) > 0 {
		envs := colab.RuntimeTemplateSoftwareConfigEnvArray{}
		for _, e := range s.Env {
			env := &colab.RuntimeTemplateSoftwareConfigEnvArgs{Name: pulumi.String(e.Name)}
			if e.Value != "" {
				env.Value = pulumi.String(e.Value)
			}
			envs = append(envs, env)
		}
		args.Envs = envs
	}
	if p := s.PostStartupScriptConfig; p != nil {
		script := &colab.RuntimeTemplateSoftwareConfigPostStartupScriptConfigArgs{}
		if p.PostStartupScript != "" {
			script.PostStartupScript = pulumi.String(p.PostStartupScript)
		}
		if p.PostStartupScriptUrl != "" {
			script.PostStartupScriptUrl = pulumi.String(p.PostStartupScriptUrl)
		}
		if p.PostStartupScriptBehavior != "" {
			script.PostStartupScriptBehavior = pulumi.String(p.PostStartupScriptBehavior)
		}
		args.PostStartupScriptConfig = script
	}
	if i := s.ColabImage; i != nil {
		image := &colab.RuntimeTemplateSoftwareConfigColabImageArgs{}
		if i.ReleaseName != "" {
			image.ReleaseName = pulumi.String(i.ReleaseName)
		}
		args.ColabImage = image
	}
	return args
}
