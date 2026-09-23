package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createdEngine is what the three arms have in common, handed to the
// folded companions (controls, serving config, widget config, assistants)
// so they address the engine and depend on it whichever resource built it.
type createdEngine struct {
	// Resource is the engine resource, for Parent/DependsOn.
	Resource pulumi.Resource
	// Name is the engine's full resource name.
	Name pulumi.StringOutput
	// EngineId is the engine's id as Google returns it.
	EngineId pulumi.StringOutput
	// DialogflowAgent is the Dialogflow CX agent a CHAT engine answers
	// through; an empty string on the other arms.
	DialogflowAgent pulumi.StringOutput
}

// engine enables the Discovery Engine API (and, on the chat arm, the
// Dialogflow API the created agent needs) and builds exactly one of the
// three provider resources the spec's engine_type selects. The arms share
// engine_id, display_name, location, data_store_ids, industry_vertical,
// and common_config; each carries its own configuration block, mapped in
// its own file. Identical to the Terraform module's count-gated resources.
func engine(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*createdEngine, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec

	// Enable the Discovery Engine API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one engine
	// must never disable the API for everything else in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("discoveryengine.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpvseng-discoveryengine.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable discoveryengine.googleapis.com api")
	}
	dependsOn := []pulumi.Resource{createdApi}

	// A chat engine creates (or links) a Dialogflow CX agent, which needs
	// the Dialogflow API on -- enabled by the module so the proof lane and
	// a fresh project never have to remember it.
	if locals.EngineType == engineTypeChat {
		dialogflowApiArgs := &projects.ServiceArgs{
			Service:                  pulumi.String("dialogflow.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}
		if spec.ProjectId.GetValue() != "" {
			dialogflowApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		createdDialogflowApi, err := projects.NewService(ctx,
			"gcpvseng-dialogflow.googleapis.com", dialogflowApiArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to enable dialogflow.googleapis.com api")
		}
		dependsOn = append(dependsOn, createdDialogflowApi)
	}

	switch locals.EngineType {
	case engineTypeChat:
		return chatEngine(ctx, locals, gcpProvider, dependsOn)
	case engineTypeRecommendation:
		return recommendationEngine(ctx, locals, gcpProvider, dependsOn)
	default:
		return searchEngine(ctx, locals, gcpProvider, dependsOn)
	}
}
