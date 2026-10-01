package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpcloudfunctionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudfunctionsv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The APIs a Gen 2 function deploy exercises: Cloud Functions (the control
// plane), Cloud Build (turns source into a container), Cloud Run (serves
// the function), Artifact Registry (stores the built image), and Eventarc
// (delivers event triggers).
var requiredApis = []string{
	"cloudfunctions.googleapis.com",
	"cloudbuild.googleapis.com",
	"run.googleapis.com",
	"artifactregistry.googleapis.com",
	"eventarc.googleapis.com",
}

// function provisions the Cloud Functions (Gen 2) function — Cloud Build
// containerizes the source with buildpacks; Cloud Run serves it.
func function(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*cloudfunctionsv2.Function, error) {
	spec := locals.GcpCloudFunction.Spec

	// Enable the required APIs before deploying so a fresh project works
	// first try. disable_on_destroy=false: turning an API off on teardown
	// is a project-wide blast radius no single resource should own.
	createdServices := make([]pulumi.Resource, 0, len(requiredApis))
	for _, api := range requiredApis {
		serviceArgs := &projects.ServiceArgs{
			Service:                  pulumi.String(api),
			DisableDependentServices: pulumi.BoolPtr(false),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}
		if spec.ProjectId.GetValue() != "" {
			serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		createdService, err := projects.NewService(ctx,
			fmt.Sprintf("function-%s", api), serviceArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to enable %s api", api)
		}
		createdServices = append(createdServices, createdService)
	}

	// Secret values the service config carries are stored in Secret Manager
	// before the function exists, and the function reads them by reference.
	// The function is created after the grants: instances read every
	// referenced secret at start, so a deploy that races its grant fails.
	storedSecrets, err := envsecrets.Store(ctx, secretPlacement(locals), secretVariables(spec), gcpProvider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store the environment's secret values")
	}

	// Secret Manager entries require an explicit project id on every entry
	// even when the function rides the ambient project — resolve the
	// effective project once (mirrors the Terraform module's
	// data.google_project lookup).
	effectiveProject := spec.ProjectId.GetValue()
	if effectiveProject == "" && needsEffectiveProject(spec) {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve ambient project from provider config")
		}
		effectiveProject = clientConfig.Project
	}

	args := &cloudfunctionsv2.FunctionArgs{
		Name:        pulumi.String(locals.FunctionName),
		Location:    pulumi.String(spec.Region),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
		BuildConfig: buildConfig(spec),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	// CMEK: encrypts the built container image and source artifacts.
	// Requires a customer-managed docker_repository and encrypter/decrypter
	// grants for the Cloud Functions + Artifact Registry service agents.
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}
	// DELETE (provider default) removes the function on destroy; PREVENT
	// fails the destroy; ABANDON leaves it serving and consuming events.
	// Sent only when set — mirrors the Terraform module.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	if spec.ServiceConfig != nil {
		args.ServiceConfig = serviceConfig(spec, effectiveProject, storedSecrets.Refs)
	}
	if isEventTrigger(spec) {
		args.EventTrigger = eventTrigger(spec)
	}

	createdFunction, err := cloudfunctionsv2.NewFunction(ctx,
		locals.GcpCloudFunction.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(append(createdServices, storedSecrets.Grants...)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloud function")
	}

	// Public invocation for HTTP functions: Gen 2 functions are served by
	// Cloud Run, so "allow unauthenticated" is run.invoker for allUsers on
	// the UNDERLYING Cloud Run service (which shares the function's name),
	// not a Cloud Functions IAM binding.
	if !isEventTrigger(spec) && spec.ServiceConfig != nil && spec.ServiceConfig.AllowUnauthenticated {
		invokerArgs := &cloudrunv2.ServiceIamMemberArgs{
			Location: pulumi.String(spec.Region),
			Name:     createdFunction.Name,
			Role:     pulumi.String("roles/run.invoker"),
			Member:   pulumi.String("allUsers"),
		}
		if spec.ProjectId.GetValue() != "" {
			invokerArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		_, err = cloudrunv2.NewServiceIamMember(ctx,
			"public-invoker",
			invokerArgs,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdFunction),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to grant public invoker on the underlying cloud run service")
		}
	}

	return createdFunction, nil
}

// isEventTrigger reports whether the function is event-driven (versus the
// HTTP default).
func isEventTrigger(spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec) bool {
	return spec.Trigger != nil &&
		spec.Trigger.TriggerType == gcpcloudfunctionv1alpha1.GcpCloudFunctionTriggerType_EVENT_TRIGGER &&
		spec.Trigger.EventTrigger != nil
}

// needsEffectiveProject reports whether any Secret Manager entry naming a
// secret the author owns omits its project — the only place the module must
// materialize the ambient project (a stored value's project comes from the
// secret the module created).
func needsEffectiveProject(spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec) bool {
	if spec.ServiceConfig == nil {
		return false
	}
	for _, sev := range spec.ServiceConfig.SecretEnvironmentVariables {
		if sev.Secret != "" && sev.ProjectId == "" {
			return true
		}
	}
	for _, sv := range spec.ServiceConfig.SecretVolumes {
		if sv.ProjectId == "" {
			return true
		}
	}
	return false
}

func buildConfig(spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec) *cloudfunctionsv2.FunctionBuildConfigArgs {
	bc := spec.BuildConfig

	sourceArgs := &cloudfunctionsv2.FunctionBuildConfigSourceArgs{}
	// Exactly one source arm — enforced pre-deploy by the spec's CEL rule.
	if bc.Source.StorageSource != nil {
		storageSource := &cloudfunctionsv2.FunctionBuildConfigSourceStorageSourceArgs{
			Bucket: pulumi.String(bc.Source.StorageSource.Bucket.GetValue()),
			Object: pulumi.String(bc.Source.StorageSource.Object),
		}
		if bc.Source.StorageSource.Generation != nil {
			storageSource.Generation = pulumi.Int(int(bc.Source.StorageSource.GetGeneration()))
		}
		sourceArgs.StorageSource = storageSource
	}
	if bc.Source.RepoSource != nil {
		repoSource := &cloudfunctionsv2.FunctionBuildConfigSourceRepoSourceArgs{
			RepoName: pulumi.StringPtr(bc.Source.RepoSource.RepoName),
		}
		if bc.Source.RepoSource.BranchName != "" {
			repoSource.BranchName = pulumi.String(bc.Source.RepoSource.BranchName)
		}
		if bc.Source.RepoSource.TagName != "" {
			repoSource.TagName = pulumi.String(bc.Source.RepoSource.TagName)
		}
		if bc.Source.RepoSource.CommitSha != "" {
			repoSource.CommitSha = pulumi.String(bc.Source.RepoSource.CommitSha)
		}
		if bc.Source.RepoSource.Dir != "" {
			repoSource.Dir = pulumi.String(bc.Source.RepoSource.Dir)
		}
		if bc.Source.RepoSource.InvertRegex {
			repoSource.InvertRegex = pulumi.Bool(true)
		}
		if bc.Source.RepoSource.ProjectId != "" {
			repoSource.ProjectId = pulumi.String(bc.Source.RepoSource.ProjectId)
		}
		sourceArgs.RepoSource = repoSource
	}

	buildArgs := &cloudfunctionsv2.FunctionBuildConfigArgs{
		Runtime:    pulumi.String(bc.Runtime),
		EntryPoint: pulumi.String(bc.EntryPoint),
		Source:     sourceArgs,
	}

	if len(bc.BuildEnvironmentVariables) > 0 {
		buildArgs.EnvironmentVariables = pulumi.ToStringMap(bc.BuildEnvironmentVariables)
	}
	// Build identity: the fully-qualified service account resource name
	// (projects/*/serviceAccounts/*) Cloud Build runs as.
	if bc.ServiceAccount.GetValue() != "" {
		buildArgs.ServiceAccount = pulumi.String(bc.ServiceAccount.GetValue())
	}
	if bc.WorkerPool.GetValue() != "" {
		buildArgs.WorkerPool = pulumi.String(bc.WorkerPool.GetValue())
	}
	if bc.DockerRepository.GetValue() != "" {
		buildArgs.DockerRepository = pulumi.String(bc.DockerRepository.GetValue())
	}

	// Runtime base-image patching: AUTOMATIC is the proto zero value AND
	// the API default, so it sends nothing (indistinguishable from unset —
	// and the API behaves identically either way). Only the non-default
	// ON_DEPLOY choice sends a block; the Terraform module does the same.
	if bc.UpdatePolicy == gcpcloudfunctionv1alpha1.GcpCloudFunctionBuildUpdatePolicy_ON_DEPLOY {
		buildArgs.OnDeployUpdatePolicy = &cloudfunctionsv2.FunctionBuildConfigOnDeployUpdatePolicyArgs{}
	}

	return buildArgs
}

func serviceConfig(
	spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec,
	effectiveProject string,
	secretRefs map[envsecrets.Key]envsecrets.Ref,
) *cloudfunctionsv2.FunctionServiceConfigArgs {
	sc := spec.ServiceConfig

	serviceArgs := &cloudfunctionsv2.FunctionServiceConfigArgs{}

	// The Gen 2 API takes memory as a quantity string ("256M", "1Gi"); the
	// spec carries it verbatim. Unset defers to the API default (256M).
	if sc.AvailableMemory != "" {
		serviceArgs.AvailableMemory = pulumi.String(sc.AvailableMemory)
	}
	if sc.AvailableCpu != "" {
		serviceArgs.AvailableCpu = pulumi.String(sc.AvailableCpu)
	}
	if sc.TimeoutSeconds > 0 {
		serviceArgs.TimeoutSeconds = pulumi.Int(int(sc.TimeoutSeconds))
	}
	// GCP defaults concurrency to 1 (every request its own instance);
	// values above 1 require at least 1 CPU.
	if sc.MaxInstanceRequestConcurrency > 0 {
		serviceArgs.MaxInstanceRequestConcurrency = pulumi.Int(int(sc.MaxInstanceRequestConcurrency))
	}
	if sc.Scaling != nil {
		serviceArgs.MinInstanceCount = pulumi.Int(int(sc.Scaling.MinInstanceCount))
		serviceArgs.MaxInstanceCount = pulumi.Int(int(sc.Scaling.MaxInstanceCount))
	}
	// Runtime identity: bare service-account email.
	if sc.ServiceAccountEmail.GetValue() != "" {
		serviceArgs.ServiceAccountEmail = pulumi.String(sc.ServiceAccountEmail.GetValue())
	}
	if len(sc.EnvironmentVariables) > 0 {
		serviceArgs.EnvironmentVariables = pulumi.ToStringMap(sc.EnvironmentVariables)
	}

	// Secret Manager references resolved at instance start — material never
	// appears in configuration. An entry is a secret the author owns or a
	// value this module stored (one of the two); a stored value reads the
	// exact version the module created. The API requires an explicit
	// project on every entry; default to the function's effective project.
	if len(sc.SecretEnvironmentVariables) > 0 {
		secretEnvs := cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArray{}
		for _, sev := range sc.SecretEnvironmentVariables {
			if ref, stored := secretRefs[envsecrets.Key{Name: sev.Key}]; stored {
				secretEnvs = append(secretEnvs, &cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArgs{
					Key:       pulumi.String(sev.Key),
					Secret:    ref.Secret,
					Version:   ref.Version,
					ProjectId: ref.Project,
				})
				continue
			}
			project := sev.ProjectId
			if project == "" {
				project = effectiveProject
			}
			version := sev.Version
			if version == "" {
				version = "latest"
			}
			secretEnvs = append(secretEnvs, &cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArgs{
				Key:       pulumi.String(sev.Key),
				Secret:    pulumi.String(sev.Secret),
				Version:   pulumi.String(version),
				ProjectId: pulumi.String(project),
			})
		}
		serviceArgs.SecretEnvironmentVariables = secretEnvs
	}

	if len(sc.SecretVolumes) > 0 {
		secretVolumes := cloudfunctionsv2.FunctionServiceConfigSecretVolumeArray{}
		for _, sv := range sc.SecretVolumes {
			project := sv.ProjectId
			if project == "" {
				project = effectiveProject
			}
			volumeArgs := &cloudfunctionsv2.FunctionServiceConfigSecretVolumeArgs{
				MountPath: pulumi.String(sv.MountPath),
				Secret:    pulumi.String(sv.Secret),
				ProjectId: pulumi.String(project),
			}
			if len(sv.Versions) > 0 {
				versions := cloudfunctionsv2.FunctionServiceConfigSecretVolumeVersionArray{}
				for _, v := range sv.Versions {
					versions = append(versions, &cloudfunctionsv2.FunctionServiceConfigSecretVolumeVersionArgs{
						Version: pulumi.String(v.Version),
						Path:    pulumi.String(v.Path),
					})
				}
				volumeArgs.Versions = versions
			}
			secretVolumes = append(secretVolumes, volumeArgs)
		}
		serviceArgs.SecretVolumes = secretVolumes
	}

	if sc.VpcConnector.GetValue() != "" {
		serviceArgs.VpcConnector = pulumi.String(sc.VpcConnector.GetValue())
		// Egress settings only make sense with a connector attached (sending
		// them without one is an API error) — but WITH a connector they are
		// sent explicitly on every apply, PRIVATE_RANGES_ONLY included: the
		// API materializes that default server-side and echoes it into state,
		// so an omitted value re-plans as a perpetual remove-diff (live-caught
		// by the idempotency gate). Same explicit-send grain as
		// direct_vpc_egress below.
		serviceArgs.VpcConnectorEgressSettings = pulumi.String(sc.VpcConnectorEgressSettings.String())
	}

	// Direct VPC egress: the connectorless path (mutually exclusive with
	// vpc_connector — the spec's CEL enforces it pre-deploy). The egress
	// mode is sent explicitly whenever the interface is present — the
	// field is Optional+Computed on the provider, so an ALL_TRAFFIC ->
	// PRIVATE_RANGES_ONLY spec edit would otherwise silently keep the old
	// live value. The API takes VPC_EGRESS_-prefixed values; the spec
	// carries the bare enum names.
	if sc.DirectVpcNetworkInterface != nil {
		interfaceArgs := &cloudfunctionsv2.FunctionServiceConfigDirectVpcNetworkInterfaceArgs{}
		if sc.DirectVpcNetworkInterface.Network.GetValue() != "" {
			interfaceArgs.Network = pulumi.StringPtr(sc.DirectVpcNetworkInterface.Network.GetValue())
		}
		if sc.DirectVpcNetworkInterface.Subnetwork.GetValue() != "" {
			interfaceArgs.Subnetwork = pulumi.StringPtr(sc.DirectVpcNetworkInterface.Subnetwork.GetValue())
		}
		if len(sc.DirectVpcNetworkInterface.Tags) > 0 {
			interfaceArgs.Tags = pulumi.ToStringArray(sc.DirectVpcNetworkInterface.Tags)
		}
		serviceArgs.DirectVpcNetworkInterfaces = cloudfunctionsv2.FunctionServiceConfigDirectVpcNetworkInterfaceArray{interfaceArgs}
		serviceArgs.DirectVpcEgress = pulumi.StringPtr("VPC_EGRESS_" + sc.DirectVpcEgress.String())
	}
	if sc.IngressSettings != gcpcloudfunctionv1alpha1.GcpCloudFunctionIngressSetting_ALLOW_ALL {
		serviceArgs.IngressSettings = pulumi.String(sc.IngressSettings.String())
	}

	// true (the API default) sends 100% of traffic to the latest ready
	// revision; false holds traffic for manual canary/rollback on the
	// underlying Cloud Run service.
	allTraffic := true
	if sc.AllTrafficOnLatestRevision != nil {
		allTraffic = sc.GetAllTrafficOnLatestRevision()
	}
	serviceArgs.AllTrafficOnLatestRevision = pulumi.Bool(allTraffic)

	if sc.BinaryAuthorizationPolicy != "" {
		serviceArgs.BinaryAuthorizationPolicy = pulumi.String(sc.BinaryAuthorizationPolicy)
	}

	return serviceArgs
}

func eventTrigger(spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec) *cloudfunctionsv2.FunctionEventTriggerArgs {
	et := spec.Trigger.EventTrigger

	triggerArgs := &cloudfunctionsv2.FunctionEventTriggerArgs{
		EventType: pulumi.String(et.EventType),
	}

	if et.PubsubTopic.GetValue() != "" {
		triggerArgs.PubsubTopic = pulumi.String(et.PubsubTopic.GetValue())
	}
	// If unset, GCP uses the function's region; multi-region sources
	// (Storage multi-region buckets) use "us"/"eu".
	if et.TriggerRegion != "" {
		triggerArgs.TriggerRegion = pulumi.String(et.TriggerRegion)
	}
	if et.RetryPolicy != gcpcloudfunctionv1alpha1.GcpCloudFunctionRetryPolicy_RETRY_POLICY_DO_NOT_RETRY {
		triggerArgs.RetryPolicy = pulumi.String(et.RetryPolicy.String())
	}
	if et.ServiceAccountEmail.GetValue() != "" {
		triggerArgs.ServiceAccountEmail = pulumi.String(et.ServiceAccountEmail.GetValue())
	}

	if len(et.EventFilters) > 0 {
		filters := cloudfunctionsv2.FunctionEventTriggerEventFilterArray{}
		for _, filter := range et.EventFilters {
			filterArgs := &cloudfunctionsv2.FunctionEventTriggerEventFilterArgs{
				Attribute: pulumi.String(filter.Attribute),
				Value:     pulumi.String(filter.Value),
			}
			if filter.Operator != nil && *filter.Operator != "" {
				filterArgs.Operator = pulumi.String(*filter.Operator)
			}
			filters = append(filters, filterArgs)
		}
		triggerArgs.EventFilters = filters
	}

	return triggerArgs
}
