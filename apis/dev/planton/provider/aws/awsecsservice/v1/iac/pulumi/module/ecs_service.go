package module

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/plantonhq/planton/internal/valuefrom"
	"k8s.io/utils/pointer"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/appautoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	awsecsservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsecsservice/v1"
)

// fallback when user doesn't supply alb.priority
const defaultAlbPriority = 100

// service creates and wires up the ECS Task Definition and AWS ECS Service resources.
func ecsService(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsEcsService.Spec
	serviceName := locals.AwsEcsService.Metadata.Name

	// ---------------------------------------------------------------------
	// CloudWatch log-group setup (enabled by default)
	// ---------------------------------------------------------------------
	loggingEnabled := true
	if spec.Container.Logging != nil {
		loggingEnabled = spec.Container.Logging.Enabled
	}

	logGroupName := fmt.Sprintf("/ecs/%s", serviceName)
	var logGroup *cloudwatch.LogGroup
	if loggingEnabled {
		var err error
		logGroup, err = cloudwatch.NewLogGroup(ctx,
			"log-group",
			&cloudwatch.LogGroupArgs{
				Name:            pulumi.String(logGroupName),
				RetentionInDays: pulumi.Int(30),
				Tags:            pulumi.ToStringMap(locals.AwsTags),
			},
			pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create CloudWatch log group")
		}
	}

	awsRegion, err := aws.GetRegion(ctx, nil, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "unable to detect AWS region")
	}

	containerDefs, err := buildContainerDefinitions(
		serviceName,
		spec.Container.Image.Repo,
		spec.Container.Image.Tag,
		spec.Container.Port,
		spec.Container.Env,
		loggingEnabled,
		logGroupName,
		awsRegion.Name,
	)
	if err != nil {
		return errors.Wrap(err, "failed to build container definitions JSON")
	}

	taskDefinitionArgs := &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(serviceName),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String(fmt.Sprintf("%d", spec.Container.Cpu)),
		Memory:                  pulumi.String(fmt.Sprintf("%d", spec.Container.Memory)),
		NetworkMode:             pulumi.String("awsvpc"),
		ContainerDefinitions:    pulumi.String(containerDefs),
		Tags:                    pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Iam != nil {
		if spec.Iam.TaskExecutionRoleArn.GetValue() != "" {
			taskDefinitionArgs.ExecutionRoleArn = pulumi.String(spec.Iam.TaskExecutionRoleArn.GetValue())
		}
		if spec.Iam.TaskRoleArn.GetValue() != "" {
			taskDefinitionArgs.TaskRoleArn = pulumi.String(spec.Iam.TaskRoleArn.GetValue())
		}
	}

	taskDef, err := ecs.NewTaskDefinition(ctx,
		"task-def",
		taskDefinitionArgs,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "unable to create ECS task definition")
	}

	serviceArgs := &ecs.ServiceArgs{
		Name:           pulumi.String(serviceName),
		Cluster:        pulumi.String(spec.ClusterArn.GetValue()),
		LaunchType:     pulumi.String("FARGATE"),
		DesiredCount:   pulumi.Int(int(spec.Container.Replicas)),
		TaskDefinition: taskDef.Arn,
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			Subnets:        pulumi.ToStringArray(valuefrom.ToStringArray(spec.Network.Subnets)),
			SecurityGroups: pulumi.ToStringArray(valuefrom.ToStringArray(spec.Network.SecurityGroups)),
		},
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Set health check grace period if ALB is enabled
	if spec.Alb != nil && spec.Alb.Enabled {
		gracePeriod := int32(60) // default
		if spec.HealthCheckGracePeriodSeconds != nil {
			gracePeriod = *spec.HealthCheckGracePeriodSeconds
		}
		serviceArgs.HealthCheckGracePeriodSeconds = pulumi.Int(int(gracePeriod))
	}

	var loadBalancerDNS pulumi.StringInput = pulumi.String("")
	var createdTargetGroupArn pulumi.StringInput = pulumi.String("")

	// Guard against nil dereference when alb block is omitted
	if spec.Alb != nil && spec.Alb.Enabled && spec.Container.Port != 0 {
		if len(spec.Network.Subnets) == 0 {
			return errors.New("at least one subnet is required for ALB usage")
		}
		if spec.Alb.Arn.GetValue() == "" {
			return errors.New("alb.arn is required when alb.enabled = true")
		}

		firstSubnetID := spec.Network.Subnets[0].GetValue()
		subnetLookup, err := ec2.LookupSubnet(ctx, &ec2.LookupSubnetArgs{
			Id: &firstSubnetID,
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to lookup subnet for ALB target group")
		}

		// ---------------- health‑check handling --------------------
		var protocol string = "HTTP"
		if spec.Alb.HealthCheck != nil && spec.Alb.HealthCheck.Protocol != "" {
			protocol = strings.ToUpper(spec.Alb.HealthCheck.Protocol)
		}

		targetGroup, err := lb.NewTargetGroup(ctx,
			"tg",
			&lb.TargetGroupArgs{
				Port:        pulumi.Int(int(spec.Container.Port)),
				Protocol:    pulumi.String(protocol),
				TargetType:  pulumi.String("ip"),
				VpcId:       pulumi.String(subnetLookup.VpcId),
				HealthCheck: healthCheckArgs(spec.Alb.HealthCheck, protocol),
				Tags:        pulumi.ToStringMap(locals.AwsTags),
			}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create ALB target group")
		}
		createdTargetGroupArn = targetGroup.Arn

		serviceArgs.LoadBalancers = ecs.ServiceLoadBalancerArray{
			&ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: targetGroup.Arn,
				ContainerName:  pulumi.String(serviceName),
				ContainerPort:  pulumi.Int(int(spec.Container.Port)),
			},
		}

		foundAlb, err := lb.LookupLoadBalancer(ctx, &lb.LookupLoadBalancerArgs{
			Arn: pointer.String(spec.Alb.Arn.GetValue()),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to find ALB by ARN")
		}
		loadBalancerDNS = pulumi.String(foundAlb.DnsName)

		listenerPort := int(spec.Alb.ListenerPort)

		foundListener, err := lb.LookupListener(ctx, &lb.LookupListenerArgs{
			LoadBalancerArn: &foundAlb.Arn,
			Port:            &listenerPort,
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to find ALB listener on the given port")
		}

		conditions := lb.ListenerRuleConditionArray{}

		if spec.Alb.RoutingType == "path" {
			if spec.Alb.Path == "" {
				return errors.New("alb.path must be set if routingType is 'path'")
			}
			conditions = lb.ListenerRuleConditionArray{
				&lb.ListenerRuleConditionArgs{
					PathPattern: &lb.ListenerRuleConditionPathPatternArgs{
						Values: pulumi.StringArray{
							pulumi.String(spec.Alb.Path),
						},
					},
				},
			}
		}

		if spec.Alb.RoutingType == "hostname" {
			if spec.Alb.Hostname == "" {
				return errors.New("alb.hostname must be set if routingType is 'hostname'")
			}
			conditions = lb.ListenerRuleConditionArray{
				&lb.ListenerRuleConditionArgs{
					HostHeader: &lb.ListenerRuleConditionHostHeaderArgs{
						Values: pulumi.StringArray{
							pulumi.String(spec.Alb.Hostname),
						},
					},
				},
			}
		}

		if len(conditions) > 0 {
			// ---------------- choose priority -------------------
			priority := defaultAlbPriority
			if spec.Alb.ListenerPriority != nil && *spec.Alb.ListenerPriority != 0 {
				priority = int(*spec.Alb.ListenerPriority)
			}

			_, err := lb.NewListenerRule(ctx,
				"listener-rule",
				&lb.ListenerRuleArgs{
					ListenerArn: pulumi.String(foundListener.Arn),
					Actions: lb.ListenerRuleActionArray{
						&lb.ListenerRuleActionArgs{
							Type:           pulumi.String("forward"),
							TargetGroupArn: targetGroup.Arn,
						},
					},
					Conditions: conditions,
					Priority:   pulumi.Int(priority),
					Tags:       pulumi.ToStringMap(locals.AwsTags),
				}, pulumi.Provider(provider))
			if err != nil {
				return errors.Wrap(err, "failed to create listener rule for path/hostname-based routing")
			}
		}
	}

	awsEcsService, err := ecs.NewService(ctx,
		"service", serviceArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "unable to create ECS service")
	}

	// Configure autoscaling if enabled
	if spec.Autoscaling != nil && spec.Autoscaling.Enabled {
		if err := configureAutoscaling(ctx, locals, provider, awsEcsService, spec); err != nil {
			return errors.Wrap(err, "failed to configure autoscaling")
		}
	}

	ctx.Export(OpAwsEcsServiceName, awsEcsService.Name)
	ctx.Export(OpEcsClusterName, pulumi.String(spec.ClusterArn.GetValue()))
	ctx.Export(OpLoadBalancerDnsName, loadBalancerDNS)
	ctx.Export(OpServiceArn, awsEcsService.Arn)
	ctx.Export(OpTargetGroupArn, createdTargetGroupArn)

	var serviceUrl pulumi.StringInput = pulumi.String("")
	if spec.Alb != nil &&
		strings.ToLower(spec.Alb.RoutingType) == "hostname" &&
		spec.Alb.Enabled && spec.Alb.Hostname != "" {
		serviceUrl = pulumi.String(fmt.Sprintf("http://%s", spec.Alb.Hostname))
	}
	ctx.Export(OpServiceUrl, serviceUrl)
	ctx.Export(OpServiceDiscoveryName, pulumi.String(""))

	if loggingEnabled && logGroup != nil {
		ctx.Export(OpCloudwatchLogGroupName, logGroup.Name)
		ctx.Export(OpCloudwatchLogGroupArn, logGroup.Arn)
	}

	return nil
}

// buildContainerDefinitions constructs a JSON array of container definitions.
// It honours env.variables, env.secrets, env.s3_files and optional log configuration.
func buildContainerDefinitions(
	serviceName, repo, tag string,
	port int32,
	env *awsecsservicev1.AwsEcsServiceContainerEnv,
	loggingEnabled bool,
	logGroupName, region string,
) (string, error) {

	// -------- environment (key-value) ----------
	envVars := []map[string]string{}
	if env != nil {
		keys := []string{}
		for k := range env.Variables {
			keys = append(keys, k)
		}
		for k := range env.Secrets {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := ""
			if v, ok := env.Variables[k]; ok {
				val = v
			}
			if v, ok := env.Secrets[k]; ok {
				val = v
			}
			envVars = append(envVars, map[string]string{
				"name":  k,
				"value": val,
			})
		}
	}

	// -------- environmentFiles (S3) ------------
	envFiles := []map[string]string{}
	if env != nil {
		for _, uri := range env.S3Files {
			if uri != "" {
				envFiles = append(envFiles, map[string]string{
					"type":  "s3",
					"value": uri,
				})
			}
		}
	}

	// -------- container base -------------------
	container := map[string]interface{}{
		"name":      serviceName,
		"image":     fmt.Sprintf("%s:%s", repo, tag),
		"essential": true,
	}

	if len(envVars) > 0 {
		container["environment"] = envVars
	}
	if len(envFiles) > 0 {
		container["environmentFiles"] = envFiles
	}

	if port != 0 {
		container["portMappings"] = []map[string]int32{
			{
				"containerPort": port,
			},
		}
	}

	if loggingEnabled {
		container["logConfiguration"] = map[string]interface{}{
			"logDriver": "awslogs",
			"options": map[string]string{
				"awslogs-group":         logGroupName,
				"awslogs-region":        region,
				"awslogs-stream-prefix": serviceName,
			},
		}
	}

	containerDefinitions := []interface{}{container}
	encoded, err := json.Marshal(containerDefinitions)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode container definitions")
	}
	return string(encoded), nil
}

// healthCheckArgs converts the optional AwsEcsServiceHealthCheck block into
// Pulumi lb.TargetGroupHealthCheckArgs, applying sensible defaults when the
// block is absent.
func healthCheckArgs(
	hc *awsecsservicev1.AwsEcsServiceHealthCheck,
	protocol string,
) *lb.TargetGroupHealthCheckArgs {

	// ---------------- default behaviour ------------------
	if hc == nil {
		return &lb.TargetGroupHealthCheckArgs{
			Path: pulumi.String("/"),
		}
	}

	args := &lb.TargetGroupHealthCheckArgs{}

	// Path is only valid for HTTP/HTTPS
	if (protocol == "HTTP" || protocol == "HTTPS") && hc.Path != "" {
		args.Path = pulumi.String(hc.Path)
	} else if protocol == "HTTP" || protocol == "HTTPS" {
		args.Path = pulumi.String("/")
	}

	if hc.Port != "" {
		args.Port = pulumi.String(hc.Port)
	}
	if hc.Interval != 0 {
		args.Interval = pulumi.Int(int(hc.Interval))
	}
	if hc.Timeout != 0 {
		args.Timeout = pulumi.Int(int(hc.Timeout))
	}
	if hc.HealthyThreshold != 0 {
		args.HealthyThreshold = pulumi.Int(int(hc.HealthyThreshold))
	}
	if hc.UnhealthyThreshold != 0 {
		args.UnhealthyThreshold = pulumi.Int(int(hc.UnhealthyThreshold))
	}

	return args
}

// configureAutoscaling sets up AWS Application Auto Scaling with target tracking policies
// for the ECS service based on CPU and/or memory utilization.
func configureAutoscaling(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	ecsService *ecs.Service,
	spec *awsecsservicev1.AwsEcsServiceSpec,
) error {
	serviceName := locals.AwsEcsService.Metadata.Name
	autoscalingConfig := spec.Autoscaling

	// Extract cluster name from ARN (format: arn:aws:ecs:region:account:cluster/name)
	clusterArn := spec.ClusterArn.GetValue()
	clusterName := clusterArn
	if idx := strings.LastIndex(clusterArn, "/"); idx != -1 {
		clusterName = clusterArn[idx+1:]
	}

	// Create the scalable target
	target, err := appautoscaling.NewTarget(ctx,
		"autoscaling-target",
		&appautoscaling.TargetArgs{
			MaxCapacity:       pulumi.Int(int(autoscalingConfig.MaxTasks)),
			MinCapacity:       pulumi.Int(int(autoscalingConfig.MinTasks)),
			ResourceId:        pulumi.Sprintf("service/%s/%s", clusterName, serviceName),
			ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
			ServiceNamespace:  pulumi.String("ecs"),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{ecsService}))
	if err != nil {
		return errors.Wrap(err, "failed to create autoscaling target")
	}

	// Create CPU-based scaling policy if target_cpu_percent is set
	if autoscalingConfig.TargetCpuPercent != nil && *autoscalingConfig.TargetCpuPercent > 0 {
		_, err := appautoscaling.NewPolicy(ctx,
			"cpu-scaling-policy",
			&appautoscaling.PolicyArgs{
				PolicyType:        pulumi.String("TargetTrackingScaling"),
				ResourceId:        target.ResourceId,
				ScalableDimension: target.ScalableDimension,
				ServiceNamespace:  target.ServiceNamespace,
				TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
					TargetValue: pulumi.Float64(float64(*autoscalingConfig.TargetCpuPercent)),
					PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
						PredefinedMetricType: pulumi.String("ECSServiceAverageCPUUtilization"),
					},
					ScaleInCooldown:  pulumi.Int(300),
					ScaleOutCooldown: pulumi.Int(60),
				},
			},
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{target}))
		if err != nil {
			return errors.Wrap(err, "failed to create CPU scaling policy")
		}
	}

	// Create memory-based scaling policy if target_memory_percent is set
	if autoscalingConfig.TargetMemoryPercent != nil && *autoscalingConfig.TargetMemoryPercent > 0 {
		_, err := appautoscaling.NewPolicy(ctx,
			"memory-scaling-policy",
			&appautoscaling.PolicyArgs{
				PolicyType:        pulumi.String("TargetTrackingScaling"),
				ResourceId:        target.ResourceId,
				ScalableDimension: target.ScalableDimension,
				ServiceNamespace:  target.ServiceNamespace,
				TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
					TargetValue: pulumi.Float64(float64(*autoscalingConfig.TargetMemoryPercent)),
					PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
						PredefinedMetricType: pulumi.String("ECSServiceAverageMemoryUtilization"),
					},
					ScaleInCooldown:  pulumi.Int(300),
					ScaleOutCooldown: pulumi.Int(60),
				},
			},
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{target}))
		if err != nil {
			return errors.Wrap(err, "failed to create memory scaling policy")
		}
	}

	return nil
}
