package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// servingConfig applies the spec's control lists to the engine's default
// serving config (`google_discovery_engine_serving_config`). Google creates
// "default_search" with the engine; the provider's create is a PATCH on it
// and its delete is a no-op, so this block configures rather than owns.
// The resource depends on every control so the ids it lists exist when
// the PATCH lands (the spec already requires each id to name a declared
// control). Returns an empty name when the spec declares no serving
// config.
func servingConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	created *createdEngine, createdControls []pulumi.Resource) (pulumi.StringOutput, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec
	cfg := spec.ServingConfig
	if cfg == nil {
		return pulumi.String("").ToStringOutput(), nil
	}

	args := &discoveryengine.ServingConfigArgs{
		EngineId:     created.EngineId,
		Location:     pulumi.String(spec.Location),
		CollectionId: pulumi.String(locals.CollectionId),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if len(cfg.BoostControlIds) > 0 {
		args.BoostControlIds = pulumi.ToStringArray(cfg.BoostControlIds)
	}
	if len(cfg.FilterControlIds) > 0 {
		args.FilterControlIds = pulumi.ToStringArray(cfg.FilterControlIds)
	}
	if len(cfg.PromoteControlIds) > 0 {
		args.PromoteControlIds = pulumi.ToStringArray(cfg.PromoteControlIds)
	}
	if len(cfg.RedirectControlIds) > 0 {
		args.RedirectControlIds = pulumi.ToStringArray(cfg.RedirectControlIds)
	}
	if len(cfg.SynonymsControlIds) > 0 {
		args.SynonymsControlIds = pulumi.ToStringArray(cfg.SynonymsControlIds)
	}

	// Depending on the engine alone would not order the PATCH after the
	// controls it lists, so the serving config depends on every control.
	createdServingConfig, err := discoveryengine.NewServingConfig(ctx,
		locals.GcpVertexAiSearchEngine.Metadata.Name+"-serving-config", args,
		pulumi.Provider(gcpProvider),
		pulumi.Parent(created.Resource),
		pulumi.DependsOn(createdControls))
	if err != nil {
		return pulumi.StringOutput{}, errors.Wrap(err, "failed to configure discovery engine serving config")
	}
	return createdServingConfig.Name, nil
}
