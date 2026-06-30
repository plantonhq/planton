package module

// Output keys for the AwsNetworkLoadBalancer module. These constants match
// the field names in AwsNetworkLoadBalancerStackOutputs.
const (
	OpLoadBalancerArn          = "load_balancer_arn"
	OpLoadBalancerName         = "load_balancer_name"
	OpLoadBalancerDnsName      = "load_balancer_dns_name"
	OpLoadBalancerHostedZoneId = "load_balancer_hosted_zone_id"
	OpListenerArns             = "listener_arns"
	OpTargetGroupArns          = "target_group_arns"
)
