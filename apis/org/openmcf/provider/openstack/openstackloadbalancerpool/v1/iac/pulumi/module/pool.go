package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// pool provisions the OpenStack Octavia pool and exports outputs.
func pool(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackLoadBalancerPool.Spec
	poolName := locals.OpenStackLoadBalancerPool.Metadata.Name

	poolArgs := &loadbalancer.PoolArgs{
		Name:       pulumi.String(poolName),
		ListenerId: pulumi.StringPtr(locals.ListenerId),
		Protocol:   pulumi.String(spec.Protocol),
		LbMethod:   pulumi.String(spec.LbMethod),
	}

	// Set description if provided.
	if spec.Description != "" {
		poolArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set admin_state_up if explicitly provided.
	if spec.AdminStateUp != nil {
		poolArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set persistence if provided.
	if spec.Persistence != nil {
		persistenceArgs := &loadbalancer.PoolPersistenceArgs{
			Type: pulumi.String(spec.Persistence.Type),
		}
		if spec.Persistence.CookieName != "" {
			persistenceArgs.CookieName = pulumi.StringPtr(spec.Persistence.CookieName)
		}
		poolArgs.Persistence = persistenceArgs
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		poolArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		poolArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdPool, err := loadbalancer.NewPool(
		ctx,
		strings.ToLower(poolName),
		poolArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack load balancer pool")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpPoolId, createdPool.ID())
	ctx.Export(OpName, createdPool.Name)
	ctx.Export(OpProtocol, createdPool.Protocol)
	ctx.Export(OpLbMethod, createdPool.LbMethod)
	ctx.Export(OpRegion, createdPool.Region)

	return nil
}
