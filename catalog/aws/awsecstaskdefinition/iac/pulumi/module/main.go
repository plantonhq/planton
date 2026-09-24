package module

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	awsecstaskdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecstaskdefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// defaultLogRetentionDays applies to the auto-created log group when the
// spec leaves logging.retention_days unset.
const defaultLogRetentionDays = 30

// Resources registers the task definition (and, by default, the shared
// CloudWatch log group its containers log to).
//
// The family name comes from metadata.name. AWS task definitions are
// immutable: every apply that changes anything registers a NEW revision of
// the family, and the task_definition_arn output carries the revision --
// which is exactly what lets a referencing ECS service roll on each change.
func Resources(ctx *pulumi.Context, stackInput *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEcsTaskDefinition.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsEcsTaskDefinition.Spec
	family := locals.AwsEcsTaskDefinition.Metadata.Name

	// ---------------------------------------------------------------------
	// Default logging: one shared log group, one stream prefix per container.
	//
	// The group name is decided here (not read back from the resource) so the
	// container-definitions JSON below can embed it as a plain string --
	// container definitions are one opaque JSON document to AWS, so weaving a
	// Pulumi output into it would force the whole document through an Apply.
	// The task definition instead depends on the group resource explicitly:
	// a missing group fails at task START (not registration), which is the
	// failure mode this ordering prevents.
	// ---------------------------------------------------------------------
	loggingDisabled := spec.Logging != nil && spec.Logging.Disabled
	referencedLogGroup := ""
	if spec.Logging != nil {
		referencedLogGroup = spec.Logging.GetLogGroup().GetValue()
	}

	logGroupName := ""
	var logGroup *cloudwatch.LogGroup
	if !loggingDisabled {
		if referencedLogGroup != "" {
			// An existing group owns its own retention and lifecycle; the
			// module only points the awslogs driver at it.
			logGroupName = referencedLogGroup
		} else {
			logGroupName = fmt.Sprintf("/ecs/%s", family)
			retention := int32(defaultLogRetentionDays)
			if spec.Logging != nil && spec.Logging.RetentionDays != nil {
				retention = *spec.Logging.RetentionDays
			}
			logGroup, err = cloudwatch.NewLogGroup(ctx,
				"log-group",
				&cloudwatch.LogGroupArgs{
					Name:            pulumi.String(logGroupName),
					RetentionInDays: pulumi.Int(int(retention)),
					Tags:            pulumi.ToStringMap(locals.AwsTags),
				},
				pulumi.Provider(provider))
			if err != nil {
				return errors.Wrap(err, "failed to create CloudWatch log group")
			}
		}
	}

	storedSecrets, err := storeSecretEnvironment(ctx, family, spec, locals.AwsTags, provider)
	if err != nil {
		return errors.Wrap(err, "failed to store the secret_environment values")
	}

	containerDefs, err := containerDefinitionsInput(spec, logGroupName, storedSecrets)
	if err != nil {
		return errors.Wrap(err, "failed to build container definitions JSON")
	}

	args := &ecs.TaskDefinitionArgs{
		Family:               pulumi.String(family),
		ContainerDefinitions: containerDefs,
		Tags:                 pulumi.ToStringMap(locals.AwsTags),
	}

	// Empty means FARGATE -- the serverless default this catalog leads with.
	compatibilities := spec.RequiresCompatibilities
	if len(compatibilities) == 0 {
		compatibilities = []string{"FARGATE"}
	}
	args.RequiresCompatibilities = pulumi.ToStringArray(compatibilities)

	// awsvpc is both the Fargate requirement and the modern EC2 posture, so
	// it is the default rather than AWS's launch-type-dependent one.
	networkMode := spec.NetworkMode
	if networkMode == "" {
		networkMode = "awsvpc"
	}
	args.NetworkMode = pulumi.String(networkMode)

	// AWS takes task-level sizing as strings (they predate typed sizes).
	// Zero means unset -- valid only for EC2/EXTERNAL, which the spec's CEL
	// already guarantees.
	if spec.Cpu > 0 {
		args.Cpu = pulumi.String(fmt.Sprintf("%d", spec.Cpu))
	}
	if spec.Memory > 0 {
		args.Memory = pulumi.String(fmt.Sprintf("%d", spec.Memory))
	}

	// Two roles by design: the agent's setup identity (pull images, fetch
	// secrets, write logs) and the application's runtime identity stay
	// separate so neither accumulates the other's permissions.
	if spec.ExecutionRole.GetValue() != "" {
		args.ExecutionRoleArn = pulumi.StringPtr(spec.ExecutionRole.GetValue())
	}
	if spec.TaskRole.GetValue() != "" {
		args.TaskRoleArn = pulumi.StringPtr(spec.TaskRole.GetValue())
	}

	// The FIS opt-in that lets chaos experiments target this task's
	// containers; false stays unsent (Optional+Computed at the provider --
	// sending false forces a diff on definitions registered before the
	// field existed).
	if spec.EnableFaultInjection {
		args.EnableFaultInjection = pulumi.BoolPtr(true)
	}

	// Namespace sharing is EC2-surface (CEL guards the Fargate rules);
	// empty stays unsent so Docker's per-container default stays in charge.
	if spec.IpcMode != "" {
		args.IpcMode = pulumi.StringPtr(spec.IpcMode)
	}
	if spec.PidMode != "" {
		args.PidMode = pulumi.StringPtr(spec.PidMode)
	}

	// Task-level placement rules, evaluated when tasks schedule onto EC2
	// container instances.
	if len(spec.PlacementConstraints) > 0 {
		constraints := make(ecs.TaskDefinitionPlacementConstraintArray, 0, len(spec.PlacementConstraints))
		for _, constraint := range spec.PlacementConstraints {
			constraints = append(constraints, &ecs.TaskDefinitionPlacementConstraintArgs{
				Type:       pulumi.String(constraint.Type),
				Expression: pulumi.StringPtr(constraint.Expression),
			})
		}
		args.PlacementConstraints = constraints
	}

	if spec.RuntimePlatform != nil {
		runtimePlatform := &ecs.TaskDefinitionRuntimePlatformArgs{}
		if spec.RuntimePlatform.CpuArchitecture != "" {
			runtimePlatform.CpuArchitecture = pulumi.StringPtr(spec.RuntimePlatform.CpuArchitecture)
		}
		if spec.RuntimePlatform.OperatingSystemFamily != "" {
			runtimePlatform.OperatingSystemFamily = pulumi.StringPtr(spec.RuntimePlatform.OperatingSystemFamily)
		}
		args.RuntimePlatform = runtimePlatform
	}

	if spec.EphemeralStorageGib > 0 {
		args.EphemeralStorage = &ecs.TaskDefinitionEphemeralStorageArgs{
			SizeInGib: pulumi.Int(int(spec.EphemeralStorageGib)),
		}
	}

	if len(spec.Volumes) > 0 {
		volumes := make(ecs.TaskDefinitionVolumeArray, 0, len(spec.Volumes))
		for _, volume := range spec.Volumes {
			volumeArgs := &ecs.TaskDefinitionVolumeArgs{
				Name: pulumi.String(volume.Name),
			}
			// A launch-time volume declares only its name here -- the
			// consuming service's volume_configuration supplies the
			// managed-EBS backing. false stays unsent (Optional+Computed
			// at the provider).
			if volume.ConfigureAtLaunch {
				volumeArgs.ConfigureAtLaunch = pulumi.BoolPtr(true)
			}
			if volume.Efs != nil {
				efs := &ecs.TaskDefinitionVolumeEfsVolumeConfigurationArgs{
					// References (AwsElasticFileSystem / AwsEfsAccessPoint)
					// arrive pre-resolved; GetValue() reads literal or
					// resolved value alike.
					FileSystemId: pulumi.String(volume.Efs.FileSystemId.GetValue()),
					// Transit encryption is always on: AWS requires it with
					// access points or IAM auth, and there is no good reason
					// to mount EFS unencrypted in transit without them.
					TransitEncryption: pulumi.StringPtr("ENABLED"),
				}
				if volume.Efs.RootDirectory != "" {
					efs.RootDirectory = pulumi.StringPtr(volume.Efs.RootDirectory)
				}
				if volume.Efs.TransitEncryptionPort != 0 {
					efs.TransitEncryptionPort = pulumi.IntPtr(int(volume.Efs.TransitEncryptionPort))
				}
				if volume.Efs.AccessPointId.GetValue() != "" || volume.Efs.IamAuthorization {
					authorization := &ecs.TaskDefinitionVolumeEfsVolumeConfigurationAuthorizationConfigArgs{}
					if volume.Efs.AccessPointId.GetValue() != "" {
						authorization.AccessPointId = pulumi.StringPtr(volume.Efs.AccessPointId.GetValue())
					}
					if volume.Efs.IamAuthorization {
						authorization.Iam = pulumi.StringPtr("ENABLED")
					}
					efs.AuthorizationConfig = authorization
				}
				volumeArgs.EfsVolumeConfiguration = efs
			}
			if volume.HostPath != "" {
				volumeArgs.HostPath = pulumi.StringPtr(volume.HostPath)
			}
			if volume.Docker != nil {
				docker := &ecs.TaskDefinitionVolumeDockerVolumeConfigurationArgs{}
				if volume.Docker.Autoprovision {
					docker.Autoprovision = pulumi.BoolPtr(true)
				}
				if volume.Docker.Driver != "" {
					docker.Driver = pulumi.StringPtr(volume.Docker.Driver)
				}
				if len(volume.Docker.DriverOpts) > 0 {
					docker.DriverOpts = pulumi.ToStringMap(volume.Docker.DriverOpts)
				}
				if len(volume.Docker.Labels) > 0 {
					docker.Labels = pulumi.ToStringMap(volume.Docker.Labels)
				}
				if volume.Docker.Scope != "" {
					docker.Scope = pulumi.StringPtr(volume.Docker.Scope)
				}
				volumeArgs.DockerVolumeConfiguration = docker
			}
			if volume.S3Files != nil {
				s3files := &ecs.TaskDefinitionVolumeS3filesVolumeConfigurationArgs{
					FileSystemArn: pulumi.String(volume.S3Files.FileSystemArn.GetValue()),
				}
				if volume.S3Files.AccessPointArn != "" {
					s3files.AccessPointArn = pulumi.StringPtr(volume.S3Files.AccessPointArn)
				}
				if volume.S3Files.RootDirectory != "" {
					s3files.RootDirectory = pulumi.StringPtr(volume.S3Files.RootDirectory)
				}
				if volume.S3Files.TransitEncryptionPort != 0 {
					s3files.TransitEncryptionPort = pulumi.IntPtr(int(volume.S3Files.TransitEncryptionPort))
				}
				volumeArgs.S3filesVolumeConfiguration = s3files
			}
			volumes = append(volumes, volumeArgs)
		}
		args.Volumes = volumes
	}

	// Keep old revisions registered on destroy when other consumers (a
	// scheduled task, a manual RunTask) may still reference them.
	if spec.SkipDestroy {
		args.SkipDestroy = pulumi.BoolPtr(true)
	}

	var resourceOptions []pulumi.ResourceOption
	resourceOptions = append(resourceOptions, pulumi.Provider(provider))
	if logGroup != nil {
		// See the logging comment above: the group must exist before any
		// task launches from this revision.
		resourceOptions = append(resourceOptions, pulumi.DependsOn([]pulumi.Resource{logGroup}))
	}

	created, err := ecs.NewTaskDefinition(ctx, "task-definition", args, resourceOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to register ECS task definition")
	}

	ctx.Export(OpTaskDefinitionArn, created.Arn)
	ctx.Export(OpArnWithoutRevision, created.ArnWithoutRevision)
	ctx.Export(OpFamily, created.Family)
	ctx.Export(OpRevision, created.Revision)

	// Log-group outputs are only populated when this module owns the group;
	// a referenced group publishes its own outputs on its own resource.
	if logGroup != nil {
		ctx.Export(OpLogGroupName, logGroup.Name)
		ctx.Export(OpLogGroupArn, logGroup.Arn)
	} else {
		ctx.Export(OpLogGroupName, pulumi.String(logGroupName))
		ctx.Export(OpLogGroupArn, pulumi.String(""))
	}

	return nil
}

// containerDefinitionsInput is the container-definitions document as the
// task definition's input. Without stored secrets it is a plain string (see
// the logging comment in Resources: the document is one opaque JSON value to
// AWS). With them, each stored secret's valueFrom is an output of the secret
// it names, so the document is rendered inside one Apply over those outputs --
// the only way to weave them in, paid only by tasks that carry them.
func containerDefinitionsInput(
	spec *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionSpec,
	logGroupName string,
	storedSecrets []storedSecretEnv,
) (pulumi.StringInput, error) {
	if len(storedSecrets) == 0 {
		document, err := buildContainerDefinitions(spec, logGroupName, nil)
		if err != nil {
			return nil, err
		}
		return pulumi.String(document), nil
	}
	valueFroms := make([]interface{}, 0, len(storedSecrets))
	for _, secret := range storedSecrets {
		valueFroms = append(valueFroms, secret.valueFrom)
	}
	return pulumi.All(valueFroms...).ApplyT(func(resolved []interface{}) (string, error) {
		byContainer := map[string]map[string]string{}
		for i, secret := range storedSecrets {
			if byContainer[secret.container] == nil {
				byContainer[secret.container] = map[string]string{}
			}
			byContainer[secret.container][secret.name] = resolved[i].(string)
		}
		return buildContainerDefinitions(spec, logGroupName, byContainer)
	}).(pulumi.StringOutput), nil
}

// buildContainerDefinitions renders the spec's structured containers into
// the container-definitions JSON document the ECS API takes. Maps are
// emitted in sorted key order so the document -- and therefore the
// registered revision -- is deterministic across applies. storedSecrets
// carries, per container, the valueFrom of each secret_environment entry.
func buildContainerDefinitions(
	spec *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionSpec,
	logGroupName string,
	storedSecrets map[string]map[string]string,
) (string, error) {
	definitions := make([]map[string]interface{}, 0, len(spec.Containers))

	for _, container := range spec.Containers {
		definition := map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
			// AWS defaults essential to true when omitted; emitting it
			// explicitly keeps the registered document identical to intent.
			"essential": container.Essential == nil || *container.Essential,
		}

		if container.Cpu > 0 {
			definition["cpu"] = container.Cpu
		}
		if container.Memory > 0 {
			definition["memory"] = container.Memory
		}
		if container.MemoryReservation > 0 {
			definition["memoryReservation"] = container.MemoryReservation
		}

		if len(container.PortMappings) > 0 {
			portMappings := make([]map[string]interface{}, 0, len(container.PortMappings))
			for _, portMapping := range container.PortMappings {
				entry := map[string]interface{}{
					"containerPort": portMapping.ContainerPort,
					"protocol":      valueOrDefault(portMapping.Protocol, "tcp"),
				}
				if portMapping.Name != "" {
					entry["name"] = portMapping.Name
				}
				if portMapping.AppProtocol != "" {
					entry["appProtocol"] = portMapping.AppProtocol
				}
				portMappings = append(portMappings, entry)
			}
			definition["portMappings"] = portMappings
		}

		if len(container.EntryPoint) > 0 {
			definition["entryPoint"] = container.EntryPoint
		}
		if len(container.Command) > 0 {
			definition["command"] = container.Command
		}
		if container.WorkingDirectory != "" {
			definition["workingDirectory"] = container.WorkingDirectory
		}

		if len(container.Environment) > 0 {
			definition["environment"] = sortedNameValueList(container.Environment, "value")
		}
		// Secrets are name -> ARN pairs -- the author's own, plus the pinned
		// ARNs of the secret_environment values this module stored; the agent
		// resolves them at task start via the execution role, so no secret
		// material passes through here.
		secrets := map[string]string{}
		for name, valueFrom := range container.Secrets {
			secrets[name] = valueFrom
		}
		for name, valueFrom := range storedSecrets[container.Name] {
			secrets[name] = valueFrom
		}
		if len(secrets) > 0 {
			definition["secrets"] = sortedNameValueList(secrets, "valueFrom")
		}
		if len(container.EnvironmentFiles) > 0 {
			environmentFiles := make([]map[string]string, 0, len(container.EnvironmentFiles))
			for _, s3Arn := range container.EnvironmentFiles {
				environmentFiles = append(environmentFiles, map[string]string{
					"value": s3Arn,
					"type":  "s3",
				})
			}
			definition["environmentFiles"] = environmentFiles
		}

		if container.HealthCheck != nil {
			healthCheck := map[string]interface{}{
				"command": container.HealthCheck.Command,
			}
			if container.HealthCheck.IntervalSeconds > 0 {
				healthCheck["interval"] = container.HealthCheck.IntervalSeconds
			}
			if container.HealthCheck.TimeoutSeconds > 0 {
				healthCheck["timeout"] = container.HealthCheck.TimeoutSeconds
			}
			if container.HealthCheck.Retries > 0 {
				healthCheck["retries"] = container.HealthCheck.Retries
			}
			if container.HealthCheck.StartPeriodSeconds > 0 {
				healthCheck["startPeriod"] = container.HealthCheck.StartPeriodSeconds
			}
			definition["healthCheck"] = healthCheck
		}

		if len(container.DependsOn) > 0 {
			dependencies := make([]map[string]string, 0, len(container.DependsOn))
			for _, dependency := range container.DependsOn {
				dependencies = append(dependencies, map[string]string{
					"containerName": dependency.ContainerName,
					"condition":     dependency.Condition,
				})
			}
			definition["dependsOn"] = dependencies
		}

		if len(container.MountPoints) > 0 {
			mountPoints := make([]map[string]interface{}, 0, len(container.MountPoints))
			for _, mountPoint := range container.MountPoints {
				mountPoints = append(mountPoints, map[string]interface{}{
					"sourceVolume":  mountPoint.SourceVolume,
					"containerPath": mountPoint.ContainerPath,
					"readOnly":      mountPoint.ReadOnly,
				})
			}
			definition["mountPoints"] = mountPoints
		}

		// Log configuration precedence: the container's own block wins;
		// otherwise the task-level default wires awslogs into the shared
		// group with the container's name as the stream prefix.
		if container.LogConfiguration != nil {
			logConfiguration := map[string]interface{}{
				"logDriver": container.LogConfiguration.LogDriver,
			}
			if len(container.LogConfiguration.Options) > 0 {
				logConfiguration["options"] = sortedStringMap(container.LogConfiguration.Options)
			}
			if len(container.LogConfiguration.SecretOptions) > 0 {
				logConfiguration["secretOptions"] = sortedNameValueList(container.LogConfiguration.SecretOptions, "valueFrom")
			}
			definition["logConfiguration"] = logConfiguration
		} else if logGroupName != "" {
			definition["logConfiguration"] = map[string]interface{}{
				"logDriver": "awslogs",
				"options": map[string]string{
					"awslogs-group":         logGroupName,
					"awslogs-region":        spec.Region,
					"awslogs-stream-prefix": container.Name,
				},
			}
		}

		if container.FirelensConfiguration != nil {
			firelens := map[string]interface{}{
				"type": container.FirelensConfiguration.Type,
			}
			if len(container.FirelensConfiguration.Options) > 0 {
				firelens["options"] = sortedStringMap(container.FirelensConfiguration.Options)
			}
			definition["firelensConfiguration"] = firelens
		}

		if container.RepositoryCredentialsSecretArn != "" {
			definition["repositoryCredentials"] = map[string]string{
				"credentialsParameter": container.RepositoryCredentialsSecretArn,
			}
		}

		if container.User != "" {
			definition["user"] = container.User
		}
		if container.ReadonlyRootFilesystem {
			definition["readonlyRootFilesystem"] = true
		}
		if container.Privileged {
			definition["privileged"] = true
		}
		// initProcessEnabled lives under linuxParameters in the ECS API.
		if container.InitProcessEnabled {
			definition["linuxParameters"] = map[string]interface{}{
				"initProcessEnabled": true,
			}
		}
		if container.GpuCount > 0 {
			definition["resourceRequirements"] = []map[string]string{
				{
					"type":  "GPU",
					"value": fmt.Sprintf("%d", container.GpuCount),
				},
			}
		}

		if len(container.Ulimits) > 0 {
			ulimits := make([]map[string]interface{}, 0, len(container.Ulimits))
			for _, ulimit := range container.Ulimits {
				ulimits = append(ulimits, map[string]interface{}{
					"name":      ulimit.Name,
					"softLimit": ulimit.SoftLimit,
					"hardLimit": ulimit.HardLimit,
				})
			}
			definition["ulimits"] = ulimits
		}

		if len(container.DockerLabels) > 0 {
			definition["dockerLabels"] = sortedStringMap(container.DockerLabels)
		}

		if container.StartTimeoutSeconds > 0 {
			definition["startTimeout"] = container.StartTimeoutSeconds
		}
		if container.StopTimeoutSeconds > 0 {
			definition["stopTimeout"] = container.StopTimeoutSeconds
		}

		if container.RestartPolicy != nil && container.RestartPolicy.Enabled {
			restartPolicy := map[string]interface{}{
				"enabled": true,
			}
			if len(container.RestartPolicy.IgnoredExitCodes) > 0 {
				restartPolicy["ignoredExitCodes"] = container.RestartPolicy.IgnoredExitCodes
			}
			if container.RestartPolicy.RestartAttemptPeriodSeconds > 0 {
				restartPolicy["restartAttemptPeriod"] = container.RestartPolicy.RestartAttemptPeriodSeconds
			}
			definition["restartPolicy"] = restartPolicy
		}

		definitions = append(definitions, definition)
	}

	encoded, err := json.Marshal(definitions)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode container definitions")
	}
	return string(encoded), nil
}

// sortedNameValueList renders a proto map as the ECS API's list-of-objects
// shape ({"name": k, "<valueKey>": v}), sorted by name for a deterministic
// document.
func sortedNameValueList(entries map[string]string, valueKey string) []map[string]string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	list := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		list = append(list, map[string]string{
			"name":   key,
			valueKey: entries[key],
		})
	}
	return list
}

// sortedStringMap copies a map so json.Marshal emits it with sorted keys
// (Go's encoder sorts map keys, so a plain copy suffices; the helper exists
// to make the determinism intent explicit at call sites).
func sortedStringMap(entries map[string]string) map[string]string {
	copied := make(map[string]string, len(entries))
	for key, value := range entries {
		copied[key] = value
	}
	return copied
}

// valueOrDefault returns the value when set, else the fallback.
func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
