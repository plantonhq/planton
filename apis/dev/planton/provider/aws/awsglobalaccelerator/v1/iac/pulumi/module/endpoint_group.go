package module

import (
	"fmt"

	awsgav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsglobalaccelerator/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/globalaccelerator"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// EndpointGroupResult holds the created endpoint groups keyed by their
// composite key ("listener_name/group_name").
type EndpointGroupResult struct {
	EndpointGroups map[string]*globalaccelerator.EndpointGroup
}

// endpointGroups creates endpoint groups for all listeners, iterating over
// the nested spec structure. Each endpoint group is keyed by
// "listener_name/group_name" for output map construction.
func endpointGroups(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	listenerResult *ListenerResult,
) (*EndpointGroupResult, error) {
	result := &EndpointGroupResult{
		EndpointGroups: make(map[string]*globalaccelerator.EndpointGroup),
	}

	for _, listenerSpec := range locals.GlobalAccelerator.Spec.Listeners {
		listener, ok := listenerResult.Listeners[listenerSpec.Name]
		if !ok {
			continue
		}

		for _, groupSpec := range listenerSpec.EndpointGroups {
			compositeKey := fmt.Sprintf("%s/%s", listenerSpec.Name, groupSpec.Name)
			resourceName := fmt.Sprintf("%s-%s", listenerSpec.Name, groupSpec.Name)

			group, err := createEndpointGroup(ctx, provider, listener, groupSpec, resourceName)
			if err != nil {
				return nil, fmt.Errorf("failed to create endpoint group %s: %w", compositeKey, err)
			}
			result.EndpointGroups[compositeKey] = group
		}
	}

	return result, nil
}

// createEndpointGroup creates a single Global Accelerator endpoint group resource.
func createEndpointGroup(
	ctx *pulumi.Context,
	provider *aws.Provider,
	listener *globalaccelerator.Listener,
	spec *awsgav1.AwsGlobalAcceleratorEndpointGroup,
	resourceName string,
) (*globalaccelerator.EndpointGroup, error) {
	args := &globalaccelerator.EndpointGroupArgs{
		ListenerArn: listener.ID().ToStringOutput(),
	}

	if spec.EndpointGroupRegion != "" {
		args.EndpointGroupRegion = pulumi.StringPtr(spec.EndpointGroupRegion)
	}

	if spec.HealthCheckPort > 0 {
		args.HealthCheckPort = pulumi.IntPtr(int(spec.HealthCheckPort))
	}

	if spec.HealthCheckProtocol != nil {
		args.HealthCheckProtocol = pulumi.StringPtr(spec.GetHealthCheckProtocol())
	}

	if spec.HealthCheckPath != "" {
		args.HealthCheckPath = pulumi.StringPtr(spec.HealthCheckPath)
	}

	if spec.HealthCheckIntervalSeconds != nil {
		args.HealthCheckIntervalSeconds = pulumi.IntPtr(int(spec.GetHealthCheckIntervalSeconds()))
	}

	if spec.ThresholdCount != nil {
		args.ThresholdCount = pulumi.IntPtr(int(spec.GetThresholdCount()))
	}

	if spec.TrafficDialPercentage != 0 {
		args.TrafficDialPercentage = pulumi.Float64Ptr(spec.TrafficDialPercentage)
	}

	if len(spec.Endpoints) > 0 {
		endpointConfigs := make(globalaccelerator.EndpointGroupEndpointConfigurationArray, len(spec.Endpoints))
		for i, ep := range spec.Endpoints {
			epArgs := &globalaccelerator.EndpointGroupEndpointConfigurationArgs{
				EndpointId: pulumi.StringPtr(ep.EndpointId.GetValue()),
			}
			if ep.Weight > 0 {
				epArgs.Weight = pulumi.IntPtr(int(ep.Weight))
			}
			if ep.ClientIpPreservationEnabled {
				epArgs.ClientIpPreservationEnabled = pulumi.BoolPtr(true)
			}
			endpointConfigs[i] = epArgs
		}
		args.EndpointConfigurations = endpointConfigs
	}

	if len(spec.PortOverrides) > 0 {
		overrides := make(globalaccelerator.EndpointGroupPortOverrideArray, len(spec.PortOverrides))
		for i, po := range spec.PortOverrides {
			overrides[i] = &globalaccelerator.EndpointGroupPortOverrideArgs{
				ListenerPort: pulumi.Int(int(po.ListenerPort)),
				EndpointPort: pulumi.Int(int(po.EndpointPort)),
			}
		}
		args.PortOverrides = overrides
	}

	return globalaccelerator.NewEndpointGroup(ctx, resourceName, args, pulumi.Provider(provider))
}
