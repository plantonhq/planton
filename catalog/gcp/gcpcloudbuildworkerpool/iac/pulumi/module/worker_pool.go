package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudbuild"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var numericProject = regexp.MustCompile(`^[0-9]+$`)

// workerPool enables the Cloud Build API and creates the private pool.
func workerPool(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudBuildWorkerPool.Spec
	resourceName := locals.GcpCloudBuildWorkerPool.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Build API. DisableOnDestroy stays false: tearing down one
	// pool must never disable the API for every other build in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("cloudbuild.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcbwp-cloudbuild.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable cloudbuild.googleapis.com")
	}

	// The pool ID defaults to metadata.name -- identical to the Terraform
	// module.
	workerPoolId := spec.WorkerPoolId
	if workerPoolId == "" {
		workerPoolId = resourceName
	}

	args := &cloudbuild.WorkerPoolArgs{
		Location: pulumi.String(spec.Location),
		Name:     pulumi.String(workerPoolId),
	}
	if project != "" {
		args.Project = pulumi.StringPtr(project)
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if n := spec.NetworkConfig; n != nil {
		peeredNetwork, err := resolvePeeredNetwork(ctx, gcpProvider, n.GetPeeredNetwork().GetValue())
		if err != nil {
			return err
		}
		networkArgs := &cloudbuild.WorkerPoolNetworkConfigArgs{
			PeeredNetwork: pulumi.String(peeredNetwork),
		}
		if n.PeeredNetworkIpRange != "" {
			networkArgs.PeeredNetworkIpRange = pulumi.StringPtr(n.PeeredNetworkIpRange)
		}
		args.NetworkConfig = networkArgs
	}

	if p := spec.PrivateServiceConnect; p != nil {
		args.PrivateServiceConnect = &cloudbuild.WorkerPoolPrivateServiceConnectArgs{
			NetworkAttachment: pulumi.String(p.NetworkAttachment),
			RouteAllTraffic:   pulumi.BoolPtr(p.RouteAllTraffic),
		}
	}

	if w := spec.WorkerConfig; w != nil {
		workerArgs := &cloudbuild.WorkerPoolWorkerConfigArgs{}
		if w.MachineType != "" {
			workerArgs.MachineType = pulumi.StringPtr(w.MachineType)
		}
		if w.DiskSizeGb != 0 {
			workerArgs.DiskSizeGb = pulumi.IntPtr(int(w.DiskSizeGb))
		}
		if w.NoExternalIp != nil {
			workerArgs.NoExternalIp = pulumi.BoolPtr(w.GetNoExternalIp())
		}
		if w.EnableNestedVirtualization != nil {
			workerArgs.EnableNestedVirtualization = pulumi.BoolPtr(w.GetEnableNestedVirtualization())
		}
		args.WorkerConfig = workerArgs
	}

	createdPool, err := cloudbuild.NewWorkerPool(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create worker pool")
	}

	ctx.Export(OpName, createdPool.ID())
	ctx.Export(OpWorkerPoolId, createdPool.Name)
	ctx.Export(OpState, createdPool.State)
	ctx.Export(OpUid, createdPool.Uid)
	return nil
}

// resolvePeeredNetwork returns the peered network in the form Google
// requires, projects/{NUMBER}/global/networks/{name}: a self-link prefix is
// stripped and a project ID is resolved to its number through one project
// lookup -- the same rule as the Terraform module.
func resolvePeeredNetwork(ctx *pulumi.Context, gcpProvider *gcp.Provider, value string) (string, error) {
	path := strings.TrimPrefix(value, "https://www.googleapis.com/compute/v1/")
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return path, nil
	}
	project := segments[1]
	if numericProject.MatchString(project) {
		return path, nil
	}
	lookup, err := organizations.LookupProject(ctx,
		&organizations.LookupProjectArgs{ProjectId: pulumi.StringRef(project)},
		pulumi.Provider(gcpProvider))
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve the project number for network %s", value)
	}
	if lookup.Number == "" {
		return "", errors.Errorf("project %s has no number; cannot build the network path Google requires", project)
	}
	return "projects/" + lookup.Number + "/global/networks/" + segments[len(segments)-1], nil
}
