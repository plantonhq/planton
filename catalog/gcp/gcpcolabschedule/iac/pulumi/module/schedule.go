package module

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	gcpcolabschedulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabschedule/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/colab"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const computeSelfLinkPrefix = "https://www.googleapis.com/compute/v1/"

var numericProject = regexp.MustCompile(`^[0-9]+$`)

// schedule creates the Vertex AI schedule. desired_state is a client-side
// control the provider enforces with pause and resume calls. Exactly one
// request is rendered; the notebook request is immutable, the pipeline
// request updates in place. Each request's parent is left to the provider,
// which fills projects/{project}/locations/{location} -- the Terraform
// module's posture.
func schedule(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpColabSchedule.Spec
	metadata := locals.GcpColabSchedule.Metadata

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpcols-aiplatform.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	// Google requires display names on the schedule and its notebook run;
	// both default to metadata.name through the schedule's display name --
	// identical to the Terraform module.
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = metadata.Name
	}

	// Google types the three run counts as decimal strings.
	args := &colab.ScheduleArgs{
		Location:              pulumi.String(spec.Location),
		DisplayName:           pulumi.String(displayName),
		Cron:                  pulumi.String(spec.Cron),
		MaxConcurrentRunCount: pulumi.String(strconv.FormatInt(spec.MaxConcurrentRunCount, 10)),
		AllowQueueing:         pulumi.Bool(spec.AllowQueueing),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.StartTime != "" {
		args.StartTime = pulumi.String(spec.StartTime)
	}
	if spec.EndTime != "" {
		args.EndTime = pulumi.String(spec.EndTime)
	}
	if spec.MaxRunCount != 0 {
		args.MaxRunCount = pulumi.String(strconv.FormatInt(spec.MaxRunCount, 10))
	}
	if spec.MaxConcurrentActiveRunCount != 0 {
		args.MaxConcurrentActiveRunCount = pulumi.String(strconv.FormatInt(spec.MaxConcurrentActiveRunCount, 10))
	}
	if spec.DesiredState != "" {
		args.DesiredState = pulumi.String(spec.DesiredState)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	if job := spec.NotebookExecutionJob; job != nil {
		args.CreateNotebookExecutionJobRequest = &colab.ScheduleCreateNotebookExecutionJobRequestArgs{
			NotebookExecutionJob: buildNotebookExecutionJob(job, displayName),
		}
	}
	if job := spec.PipelineJob; job != nil {
		pipelineJob, err := buildPipelineJob(ctx, gcpProvider, job)
		if err != nil {
			return err
		}
		args.CreatePipelineJobRequest = &colab.ScheduleCreatePipelineJobRequestArgs{
			PipelineJob: pipelineJob,
		}
	}

	createdSchedule, err := colab.NewSchedule(ctx, metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create colab schedule")
	}

	ctx.Export(OpName, createdSchedule.ID())
	ctx.Export(OpScheduleId, createdSchedule.Name)
	ctx.Export(OpLocation, createdSchedule.Location)
	return nil
}

func buildNotebookExecutionJob(job *gcpcolabschedulev1alpha1.GcpColabScheduleNotebookExecutionJob, scheduleDisplayName string) *colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobArgs {
	displayName := job.DisplayName
	if displayName == "" {
		displayName = scheduleDisplayName
	}
	args := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobArgs{
		DisplayName:  pulumi.String(displayName),
		GcsOutputUri: pulumi.String(job.GcsOutputUri.GetValue()),
	}
	if v := job.NotebookRuntimeTemplateResourceName.GetValue(); v != "" {
		args.NotebookRuntimeTemplateResourceName = pulumi.String(v)
	}
	if job.ExecutionUser != "" {
		args.ExecutionUser = pulumi.String(job.ExecutionUser)
	}
	if v := job.ServiceAccount.GetValue(); v != "" {
		args.ServiceAccount = pulumi.String(v)
	}
	if job.ExecutionTimeout != "" {
		args.ExecutionTimeout = pulumi.String(job.ExecutionTimeout)
	}
	if job.KernelName != "" {
		args.KernelName = pulumi.String(job.KernelName)
	}
	if len(job.Labels) > 0 {
		args.Labels = pulumi.ToStringMap(job.Labels)
	}
	if s := job.GcsNotebookSource; s != nil {
		source := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobGcsNotebookSourceArgs{
			Uri: pulumi.String(s.Uri),
		}
		if s.Generation != "" {
			source.Generation = pulumi.String(s.Generation)
		}
		args.GcsNotebookSource = source
	}
	if s := job.DataformRepositorySource; s != nil {
		source := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobDataformRepositorySourceArgs{
			DataformRepositoryResourceName: pulumi.String(s.DataformRepositoryResourceName),
		}
		if s.CommitSha != "" {
			source.CommitSha = pulumi.String(s.CommitSha)
		}
		args.DataformRepositorySource = source
	}
	if env := job.CustomEnvironmentSpec; env != nil {
		args.CustomEnvironmentSpec = buildCustomEnvironmentSpec(env)
	}
	// Google's workbench_runtime is an empty marker block; the spec carries
	// it as a bool.
	if job.WorkbenchRuntime {
		args.WorkbenchRuntime = &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobWorkbenchRuntimeArgs{}
	}
	if v := job.KmsKeyName.GetValue(); v != "" {
		args.EncryptionSpec = &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobEncryptionSpecArgs{
			KmsKeyName: pulumi.String(v),
		}
	}
	return args
}

func buildCustomEnvironmentSpec(env *gcpcolabschedulev1alpha1.GcpColabScheduleCustomEnvironmentSpec) *colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecArgs {
	args := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecArgs{}
	if m := env.MachineSpec; m != nil {
		machine := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecMachineSpecArgs{}
		if m.MachineType != "" {
			machine.MachineType = pulumi.String(m.MachineType)
		}
		if m.AcceleratorType != "" {
			machine.AcceleratorType = pulumi.String(m.AcceleratorType)
		}
		if m.AcceleratorCount != 0 {
			machine.AcceleratorCount = pulumi.Int(int(m.AcceleratorCount))
		}
		if m.GpuPartitionSize != "" {
			machine.GpuPartitionSize = pulumi.String(m.GpuPartitionSize)
		}
		if m.TpuTopology != "" {
			machine.TpuTopology = pulumi.String(m.TpuTopology)
		}
		if r := m.ReservationAffinity; r != nil {
			affinity := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecMachineSpecReservationAffinityArgs{
				ReservationAffinityType: pulumi.String(r.ReservationAffinityType),
				UseReservationPool:      pulumi.Bool(r.UseReservationPool),
			}
			if r.Key != "" {
				affinity.Key = pulumi.String(r.Key)
			}
			if len(r.Values) > 0 {
				affinity.Values = pulumi.ToStringArray(r.Values)
			}
			machine.ReservationAffinity = affinity
		}
		args.MachineSpec = machine
	}
	if n := env.NetworkSpec; n != nil {
		network := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecNetworkSpecArgs{
			EnableInternetAccess: pulumi.Bool(n.EnableInternetAccess),
		}
		if v := n.Network.GetValue(); v != "" {
			network.Network = pulumi.String(v)
		}
		// A GcpSubnetwork reference arrives as a compute self-link; Vertex
		// AI takes the relative path -- the same trim as the Terraform module.
		if v := n.Subnetwork.GetValue(); v != "" {
			network.Subnetwork = pulumi.String(strings.TrimPrefix(v, computeSelfLinkPrefix))
		}
		args.NetworkSpec = network
	}
	// Google types the disk size as a decimal string.
	if d := env.PersistentDiskSpec; d != nil {
		disk := &colab.ScheduleCreateNotebookExecutionJobRequestNotebookExecutionJobCustomEnvironmentSpecPersistentDiskSpecArgs{}
		if d.DiskType != "" {
			disk.DiskType = pulumi.String(d.DiskType)
		}
		if d.DiskSizeGb != 0 {
			disk.DiskSizeGb = pulumi.String(strconv.FormatInt(d.DiskSizeGb, 10))
		}
		args.PersistentDiskSpec = disk
	}
	return args
}

func buildPipelineJob(ctx *pulumi.Context, gcpProvider *gcp.Provider, job *gcpcolabschedulev1alpha1.GcpColabSchedulePipelineJob) (*colab.ScheduleCreatePipelineJobRequestPipelineJobArgs, error) {
	args := &colab.ScheduleCreatePipelineJobRequestPipelineJobArgs{
		PreflightValidations: pulumi.Bool(job.PreflightValidations),
	}
	if job.DisplayName != "" {
		args.DisplayName = pulumi.String(job.DisplayName)
	}
	if job.PipelineSpec != "" {
		args.PipelineSpec = pulumi.String(job.PipelineSpec)
	}
	if job.TemplateUri != "" {
		args.TemplateUri = pulumi.String(job.TemplateUri)
	}
	if v := job.ServiceAccount.GetValue(); v != "" {
		args.ServiceAccount = pulumi.String(v)
	}
	if v := job.Network.GetValue(); v != "" {
		network, err := resolveNetwork(ctx, gcpProvider, v)
		if err != nil {
			return nil, err
		}
		args.Network = pulumi.String(network)
	}
	if len(job.ReservedIpRanges) > 0 {
		args.ReservedIpRanges = pulumi.ToStringArray(job.ReservedIpRanges)
	}
	if len(job.Labels) > 0 {
		args.Labels = pulumi.ToStringMap(job.Labels)
	}
	if rc := job.RuntimeConfig; rc != nil {
		runtimeConfig := &colab.ScheduleCreatePipelineJobRequestPipelineJobRuntimeConfigArgs{
			GcsOutputDirectory: pulumi.String(rc.GcsOutputDirectory),
		}
		if rc.FailurePolicy != "" {
			runtimeConfig.FailurePolicy = pulumi.String(rc.FailurePolicy)
		}
		if len(rc.ParameterValues) > 0 {
			runtimeConfig.ParameterValues = pulumi.ToStringMap(rc.ParameterValues)
		}
		args.RuntimeConfig = runtimeConfig
	}
	if v := job.KmsKeyName.GetValue(); v != "" {
		args.EncryptionSpec = &colab.ScheduleCreatePipelineJobRequestPipelineJobEncryptionSpecArgs{
			KmsKeyName: pulumi.String(v),
		}
	}
	if psc := job.PscInterfaceConfig; psc != nil {
		pscArgs := &colab.ScheduleCreatePipelineJobRequestPipelineJobPscInterfaceConfigArgs{}
		if psc.NetworkAttachment != "" {
			pscArgs.NetworkAttachment = pulumi.String(psc.NetworkAttachment)
		}
		if len(psc.DnsPeeringConfigs) > 0 {
			peerings := colab.ScheduleCreatePipelineJobRequestPipelineJobPscInterfaceConfigDnsPeeringConfigArray{}
			for _, p := range psc.DnsPeeringConfigs {
				peerings = append(peerings, &colab.ScheduleCreatePipelineJobRequestPipelineJobPscInterfaceConfigDnsPeeringConfigArgs{
					Domain:        pulumi.String(p.Domain),
					TargetNetwork: pulumi.String(p.TargetNetwork.GetValue()),
					TargetProject: pulumi.String(p.TargetProject.GetValue()),
				})
			}
			pscArgs.DnsPeeringConfigs = peerings
		}
		args.PscInterfaceConfig = pscArgs
	}
	return args, nil
}

// resolveNetwork turns a GcpVpcNetwork self-link or a literal path into
// the projects/{NUMBER}/global/networks/{name} form a pipeline job
// requires. When the project segment is not already a number, one project
// lookup resolves it -- the same guarded read the Terraform module
// performs.
func resolveNetwork(ctx *pulumi.Context, gcpProvider *gcp.Provider, value string) (string, error) {
	path := strings.TrimPrefix(value, computeSelfLinkPrefix)
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
