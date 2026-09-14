package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionNetworkEndpointGroup provisions the regional network endpoint group —
// the bridge that lets a backend service target a Cloud Run/Functions/App
// Engine workload, a Private Service Connect endpoint, or an external origin
// instead of a group of VMs.
//
// The whole resource is immutable in GCP (every field is ForceNew): any change
// destroys and recreates the NEG. Because an in-use NEG cannot be deleted,
// callers that recreate one should create the replacement first
// (create-before-destroy) to avoid a resourceInUseByAnotherResource error.
//
// The endpoint type gates which nested block is sent; the spec's CEL rules
// enforce the "exactly one serverless block for SERVERLESS, none otherwise"
// and PSC/internet coherence before deploy, so no defensive branching lives
// here — the module sends whatever the spec set.
func regionNetworkEndpointGroup(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpRegionNetworkEndpointGroup.Spec

	// Enable the Compute Engine API so a fresh project can host the NEG.
	// disable_on_destroy stays false (the provider default): tearing down one
	// NEG must never disable the API for everything else in the project.
	// Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"regionneg-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	args := &compute.RegionNetworkEndpointGroupArgs{
		Name:   pulumi.String(locals.NetworkEndpointGroupName),
		Region: pulumi.String(spec.Region),
	}

	// Honor the spec contract: an empty project_id falls back to the provider's
	// default project. Leaving Project unset lets the gcp provider resolve its
	// own project (configuration or the GOOGLE_PROJECT / GOOGLE_CLOUD_PROJECT
	// environment chain); an empty string would be sent verbatim and rejected.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// network_endpoint_type has a proto default of SERVERLESS; empty falls
	// through to the same GCP API default.
	if spec.GetNetworkEndpointType() != "" {
		args.NetworkEndpointType = pulumi.String(spec.GetNetworkEndpointType())
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// network/subnetwork arrive as resolved self-links (or literals) — the CLI
	// tfvars converter and the Pulumi ref accessor both flatten StringValueOrRef
	// to a plain string. Only the non-serverless endpoint types use them.
	if spec.Network.GetValue() != "" {
		args.Network = pulumi.String(spec.Network.GetValue())
	}
	if spec.Subnetwork.GetValue() != "" {
		args.Subnetwork = pulumi.String(spec.Subnetwork.GetValue())
	}
	if spec.PscTargetService != "" {
		args.PscTargetService = pulumi.String(spec.PscTargetService)
	}
	if spec.PscData != nil && spec.PscData.ProducerPort != "" {
		args.PscData = &compute.RegionNetworkEndpointGroupPscDataArgs{
			ProducerPort: pulumi.String(spec.PscData.ProducerPort),
		}
	}

	// The serverless target is a oneof, so at most one arm is set by
	// construction and a SERVERLESS NEG carries exactly one (the spec's CEL).
	// cloud_run.service is a resolved string (reference or literal Cloud Run
	// service name).
	if spec.GetCloudRun() != nil {
		cloudRun := &compute.RegionNetworkEndpointGroupCloudRunArgs{}
		if spec.GetCloudRun().Service.GetValue() != "" {
			cloudRun.Service = pulumi.String(spec.GetCloudRun().Service.GetValue())
		}
		if spec.GetCloudRun().Tag != "" {
			cloudRun.Tag = pulumi.String(spec.GetCloudRun().Tag)
		}
		if spec.GetCloudRun().UrlMask != "" {
			cloudRun.UrlMask = pulumi.String(spec.GetCloudRun().UrlMask)
		}
		args.CloudRun = cloudRun
	}
	if spec.GetCloudFunction() != nil {
		cloudFunction := &compute.RegionNetworkEndpointGroupCloudFunctionArgs{}
		if spec.GetCloudFunction().Function.GetValue() != "" {
			cloudFunction.Function = pulumi.String(spec.GetCloudFunction().Function.GetValue())
		}
		if spec.GetCloudFunction().UrlMask != "" {
			cloudFunction.UrlMask = pulumi.String(spec.GetCloudFunction().UrlMask)
		}
		args.CloudFunction = cloudFunction
	}
	if spec.GetAppEngine() != nil {
		// The App Engine block may be empty (routes to the default app), so it
		// is always emitted when present even with all sub-fields unset.
		appEngine := &compute.RegionNetworkEndpointGroupAppEngineArgs{}
		if spec.GetAppEngine().Service != "" {
			appEngine.Service = pulumi.String(spec.GetAppEngine().Service)
		}
		if spec.GetAppEngine().Version != "" {
			appEngine.Version = pulumi.String(spec.GetAppEngine().Version)
		}
		if spec.GetAppEngine().UrlMask != "" {
			appEngine.UrlMask = pulumi.String(spec.GetAppEngine().UrlMask)
		}
		args.AppEngine = appEngine
	}
	// Empty defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdNeg, err := compute.NewRegionNetworkEndpointGroup(ctx, "region-network-endpoint-group", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create region network endpoint group")
	}

	ctx.Export(OpSelfLink, createdNeg.SelfLink)
	ctx.Export(OpNetworkEndpointGroupName, createdNeg.Name)
	ctx.Export(OpNetworkEndpointType, createdNeg.NetworkEndpointType)
	// Plain region name from spec — createdNeg.Region may be a self-link on
	// newer provider lines, which breaks gcloud-style API callers.
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
