package module

import (
	"github.com/pkg/errors"
	gcpvertexaifeatureonlinestorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// featureOnlineStore creates the online store -- exactly one of a managed
// Bigtable instance or Optimized serving (the spec's CEL) -- and the
// feature views it serves (feature_views.go).
//
// Send posture (parity with the Terraform module): the Optional+Computed
// levers -- the Bigtable zone, the CPU utilization target, and the
// dedicated serving endpoint -- are sent only when set, so Google's
// defaults stay in charge and a manifest that never mentions them
// re-plans clean on either engine.
func featureOnlineStore(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiFeatureOnlineStore.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one online store
	// must never disable the API for everything else in the project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpvfos-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiFeatureOnlineStoreArgs{
		// The provider names the axis `region`; the spec keeps the Vertex
		// family's single word, `location`.
		Region:       pulumi.String(spec.Location),
		Name:         pulumi.String(spec.FeatureOnlineStoreId),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
		ForceDestroy: pulumi.Bool(spec.ForceDestroy),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Bigtable != nil {
		args.Bigtable = buildBigtable(spec.Bigtable)
	}
	// The spec's bool is the honest form of Google's empty optimized block.
	if spec.Optimized {
		args.Optimized = &vertex.AiFeatureOnlineStoreOptimizedArgs{}
	}
	if endpoint := spec.DedicatedServingEndpoint; endpoint != nil {
		endpointArgs := &vertex.AiFeatureOnlineStoreDedicatedServingEndpointArgs{}
		if psc := endpoint.PrivateServiceConnectConfig; psc != nil {
			pscArgs := &vertex.AiFeatureOnlineStoreDedicatedServingEndpointPrivateServiceConnectConfigArgs{
				EnablePrivateServiceConnect: pulumi.Bool(psc.EnablePrivateServiceConnect),
			}
			if len(psc.ProjectAllowlist) > 0 {
				pscArgs.ProjectAllowlists = pulumi.ToStringArray(psc.ProjectAllowlist)
			}
			endpointArgs.PrivateServiceConnectConfig = pscArgs
		}
		args.DedicatedServingEndpoint = endpointArgs
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiFeatureOnlineStoreEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Engine-side destroy stance, fanned to every feature view: PREVENT
	// fails destroys, ABANDON removes from management without deleting.
	// Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdStore, err := vertex.NewAiFeatureOnlineStore(ctx,
		locals.GcpVertexAiFeatureOnlineStore.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai feature online store")
	}

	featureViewNames, err := featureViews(ctx, locals, gcpProvider, createdStore)
	if err != nil {
		return err
	}

	ctx.Export(OpName, createdStore.ID().ToStringOutput())
	ctx.Export(OpFeatureOnlineStoreId, createdStore.Name)
	ctx.Export(OpLocation, createdStore.Region)
	ctx.Export(OpPublicEndpointDomainName, createdStore.DedicatedServingEndpoint.PublicEndpointDomainName().Elem())
	ctx.Export(OpServiceAttachment, createdStore.DedicatedServingEndpoint.ServiceAttachment().Elem())
	ctx.Export(OpFeatureViewNames, featureViewNames.ToStringArrayOutput())
	return nil
}

// buildBigtable maps the managed Bigtable settings; the zone and the CPU
// target are sent only when set.
func buildBigtable(bigtable *gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStoreBigtable) *vertex.AiFeatureOnlineStoreBigtableArgs {
	autoScaling := &vertex.AiFeatureOnlineStoreBigtableAutoScalingArgs{
		MinNodeCount: pulumi.Int(int(bigtable.AutoScaling.GetMinNodeCount())),
		MaxNodeCount: pulumi.Int(int(bigtable.AutoScaling.GetMaxNodeCount())),
	}
	if bigtable.AutoScaling.CpuUtilizationTarget != nil {
		autoScaling.CpuUtilizationTarget = pulumi.Int(int(bigtable.AutoScaling.GetCpuUtilizationTarget()))
	}
	args := &vertex.AiFeatureOnlineStoreBigtableArgs{AutoScaling: autoScaling}
	if bigtable.EnableDirectBigtableAccess {
		args.EnableDirectBigtableAccess = pulumi.Bool(true)
	}
	if bigtable.Zone != "" {
		args.Zone = pulumi.String(bigtable.Zone)
	}
	return args
}
