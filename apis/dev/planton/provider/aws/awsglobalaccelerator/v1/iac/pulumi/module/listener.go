package module

import (
	"fmt"

	awsgav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsglobalaccelerator/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/globalaccelerator"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ListenerResult holds the created listeners keyed by their spec name.
type ListenerResult struct {
	Listeners map[string]*globalaccelerator.Listener
}

// listeners creates a Global Accelerator listener for each entry in the spec.
func listeners(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	accel *globalaccelerator.Accelerator,
) (*ListenerResult, error) {
	result := &ListenerResult{
		Listeners: make(map[string]*globalaccelerator.Listener),
	}

	for _, listenerSpec := range locals.GlobalAccelerator.Spec.Listeners {
		listener, err := createListener(ctx, provider, accel, listenerSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener %s: %w", listenerSpec.Name, err)
		}
		result.Listeners[listenerSpec.Name] = listener
	}

	return result, nil
}

// createListener creates a single Global Accelerator listener resource.
func createListener(
	ctx *pulumi.Context,
	provider *aws.Provider,
	accel *globalaccelerator.Accelerator,
	spec *awsgav1.AwsGlobalAcceleratorListener,
) (*globalaccelerator.Listener, error) {
	portRanges := make(globalaccelerator.ListenerPortRangeArray, len(spec.PortRanges))
	for i, pr := range spec.PortRanges {
		portRanges[i] = &globalaccelerator.ListenerPortRangeArgs{
			FromPort: pulumi.IntPtr(int(pr.FromPort)),
			ToPort:   pulumi.IntPtr(int(pr.ToPort)),
		}
	}

	args := &globalaccelerator.ListenerArgs{
		AcceleratorArn: accel.ID().ToStringOutput(),
		Protocol:       pulumi.String(spec.Protocol),
		PortRanges:     portRanges,
	}

	if spec.ClientAffinity != nil {
		args.ClientAffinity = pulumi.StringPtr(spec.GetClientAffinity())
	}

	return globalaccelerator.NewListener(ctx, spec.Name, args, pulumi.Provider(provider))
}
