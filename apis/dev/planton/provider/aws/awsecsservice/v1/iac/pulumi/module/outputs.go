package module

// These constants define the output keys we export from our Pulumi module.
// They correspond to fields in AwsEcsServiceStackOutputs, so that other parts
// of Planton can consume them consistently.
const (
	OpAwsEcsServiceName      = "aws_ecs_service_name"
	OpEcsClusterName         = "ecs_cluster_name"
	OpLoadBalancerDnsName    = "load_balancer_dns_name"
	OpServiceUrl             = "service_url"
	OpServiceDiscoveryName   = "service_discovery_name"
	OpCloudwatchLogGroupName = "cloudwatch_log_group_name"
	OpCloudwatchLogGroupArn  = "cloudwatch_log_group_arn"
	OpServiceArn             = "service_arn"
	OpTargetGroupArn         = "target_group_arn"
)
