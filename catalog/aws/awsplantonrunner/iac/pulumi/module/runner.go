package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// containerName is the single container of the task -- a stable name so
// operators tailing logs always find the same stream prefix.
const containerName = "planton-runner"

// The runner's listening ports: the gRPC/CloudOps server and the health
// server. Both are private to the task ENI -- the security group admits no
// inbound traffic; the CloudOps channel reaches the runner through its own
// outbound-initiated tunnel.
const (
	grpcPort   = 50051
	healthPort = 8093
)

// runnerCompute provisions the compute stack of the appliance -- log
// group, dedicated cluster, task definition, and the Fargate service that
// keeps exactly one runner running -- and exports the component's
// outputs.
func runnerCompute(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSecret *secretsmanager.Secret,
	createdExecutionRole *iam.Role,
	taskRoleArn pulumi.StringOutput,
	createdSecurityGroup *ec2.SecurityGroup,
) error {
	spec := locals.AwsPlantonRunner.Spec
	runnerName := locals.AwsPlantonRunner.Metadata.Name

	// The group name is decided here (not read back from the resource) so
	// the container-definitions JSON below can embed it as a plain string;
	// the task definition depends on the group explicitly so a task never
	// starts before its log destination exists.
	logGroupName := fmt.Sprintf("/ecs/%s", runnerName)
	createdLogGroup, err := cloudwatch.NewLogGroup(ctx,
		"log-group",
		&cloudwatch.LogGroupArgs{
			Name: pulumi.String(logGroupName),
			// The runner's logs are the audit trail of every operation it
			// executed -- retention is explicit, never CloudWatch's
			// never-expire default.
			RetentionInDays: pulumi.Int(int(spec.GetLogRetentionDays())),
			Tags:            pulumi.ToStringMap(locals.AwsTags),
		},
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create log group")
	}

	// A dedicated cluster per runner: clusters are free scheduling
	// boundaries (no cost until tasks run), and a dedicated one keeps the
	// appliance self-contained -- its teardown removes everything, and no
	// shared-cluster coupling can block or complicate it.
	createdCluster, err := ecs.NewCluster(ctx,
		"cluster",
		&ecs.ClusterArgs{
			Name: pulumi.String(runnerName),
			Tags: pulumi.ToStringMap(locals.AwsTags),
		},
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create ECS cluster")
	}

	// The container definition embeds the secret's ARN, which is only
	// known after the secret exists -- so the (otherwise static) JSON
	// document is assembled inside an Apply over that one output.
	containerDefinitions := createdSecret.Arn.ApplyT(func(secretArn string) (string, error) {
		return buildContainerDefinitions(locals, logGroupName, secretArn)
	}).(pulumi.StringOutput)

	createdTaskDefinition, err := ecs.NewTaskDefinition(ctx,
		"task-definition",
		&ecs.TaskDefinitionArgs{
			Family:                  pulumi.String(runnerName),
			RequiresCompatibilities: pulumi.ToStringArray([]string{"FARGATE"}),
			NetworkMode:             pulumi.String("awsvpc"),
			// Fargate takes task-level sizing as strings (they predate
			// typed sizes); the spec's CEL already guarantees a valid
			// cpu/memory pairing.
			Cpu:                  pulumi.String(fmt.Sprintf("%d", spec.GetCpu())),
			Memory:               pulumi.String(fmt.Sprintf("%d", spec.GetMemory())),
			ExecutionRoleArn:     createdExecutionRole.Arn,
			TaskRoleArn:          taskRoleArn,
			ContainerDefinitions: containerDefinitions,
			Tags:                 pulumi.ToStringMap(locals.AwsTags),
		},
		pulumi.Provider(provider),
		// A missing log group fails at task START (not registration) --
		// this explicit edge is what prevents that failure mode.
		pulumi.DependsOn([]pulumi.Resource{createdLogGroup}))
	if err != nil {
		return errors.Wrap(err, "failed to register task definition")
	}

	// The service's security groups: always the created outbound-only
	// group, plus any referenced extras (groups a private target trusts).
	securityGroupIds := pulumi.StringArray{createdSecurityGroup.ID().ToStringOutput()}
	for _, extraGroupId := range valuefrom.ToStringArray(spec.SecurityGroups) {
		securityGroupIds = append(securityGroupIds, pulumi.String(extraGroupId))
	}

	// Exactly one runner per registration: the registration's work queue
	// serializes operations, so scaling execution capacity means more
	// runners, never more copies of this one. The service deliberately
	// does NOT gate on steady state and carries no deployment circuit
	// breaker: ECS reports the service ACTIVE independently of task
	// health, and a runner whose control plane is momentarily unreachable
	// must still deploy and destroy cleanly -- its readiness contract is
	// the work queue, not ECS task liveness.
	createdService, err := ecs.NewService(ctx,
		"service",
		&ecs.ServiceArgs{
			Name:           pulumi.String(runnerName),
			Cluster:        createdCluster.Arn,
			TaskDefinition: createdTaskDefinition.Arn,
			DesiredCount:   pulumi.Int(1),
			LaunchType:     pulumi.String("FARGATE"),
			NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
				Subnets:        pulumi.ToStringArray(valuefrom.ToStringArray(spec.Subnets)),
				SecurityGroups: securityGroupIds,
				AssignPublicIp: pulumi.BoolPtr(spec.AssignPublicIp),
			},
			Tags: pulumi.ToStringMap(locals.AwsTags),
		},
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create ECS service")
	}

	ctx.Export(OpServiceArn, createdService.Arn)
	ctx.Export(OpServiceName, createdService.Name)
	ctx.Export(OpClusterArn, createdCluster.Arn)
	ctx.Export(OpTaskDefinitionArn, createdTaskDefinition.Arn)
	ctx.Export(OpLogGroupName, createdLogGroup.Name)
	ctx.Export(OpSecurityGroupId, createdSecurityGroup.ID())
	ctx.Export(OpExecutionRoleArn, createdExecutionRole.Arn)
	ctx.Export(OpTaskRoleArn, taskRoleArn)
	ctx.Export(OpTokenSecretArn, createdSecret.Arn)
	ctx.Export(OpRegion, pulumi.String(spec.Region))
	ctx.Export(OpRunnerName, pulumi.String(locals.RegistrationName))

	return nil
}

// buildContainerDefinitions renders the single-container definition the
// runner runs with. The document is deterministic (fixed field order via
// the struct maps below and Go's sorted-key JSON encoding), so the
// registered revision only changes when the configuration actually does.
func buildContainerDefinitions(locals *Locals, logGroupName string, secretArn string) (string, error) {
	spec := locals.AwsPlantonRunner.Spec

	// The runner's environment contract: the name it registers itself
	// under, the control-plane endpoint when one is declared (omitted,
	// the runner's built-in hosted default applies), and the base process
	// settings. No EXECUTION_MODE: only the control plane knows whether
	// its work queue is reachable, so the runner derives its mode from
	// the identity the join returns -- a mode knob here would silently
	// strip capability.
	environment := []map[string]string{
		{"name": "PLANTON_RUNNER_NAME", "value": locals.RegistrationName},
		{"name": "PORT", "value": fmt.Sprintf("%d", grpcPort)},
		{"name": "LOG_LEVEL", "value": "info"},
	}
	if spec.GetControlPlaneEndpoint() != "" {
		environment = append(environment,
			map[string]string{"name": "PLANTON_RUNNER_ENDPOINT", "value": spec.GetControlPlaneEndpoint()})
	}

	definition := []map[string]interface{}{
		{
			"name":      containerName,
			"image":     fmt.Sprintf("%s:%s", spec.GetImageRepository(), spec.GetRunnerVersion()),
			"command":   []string{"start"},
			"essential": true,
			"portMappings": []map[string]interface{}{
				{"containerPort": grpcPort, "protocol": "tcp"},
				{"containerPort": healthPort, "protocol": "tcp"},
			},
			"environment": environment,
			// The runner token arrives as an env var resolved by the ECS
			// agent at task start via the execution role -- never as
			// plaintext in this document. The token is only read at join,
			// so rotating the secret needs no task restart: running tasks
			// keep serving on their minted identity, and the next task
			// replacement joins with the new value.
			"secrets": []map[string]string{
				{"name": "PLANTON_RUNNER_TOKEN", "valueFrom": secretArn},
			},
			"logConfiguration": map[string]interface{}{
				"logDriver": "awslogs",
				"options": map[string]string{
					"awslogs-group":         logGroupName,
					"awslogs-region":        spec.Region,
					"awslogs-stream-prefix": "runner",
				},
			},
		},
	}

	encoded, err := json.Marshal(definition)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode container definitions")
	}
	return string(encoded), nil
}
