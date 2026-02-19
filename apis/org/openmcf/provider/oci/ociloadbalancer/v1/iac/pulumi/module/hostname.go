package module

import (
	"fmt"

	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createHostnames(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, lb *loadbalancer.LoadBalancer) ([]*loadbalancer.Hostname, error) {
	spec := locals.OciLoadBalancer.Spec
	var created []*loadbalancer.Hostname

	for _, hostSpec := range spec.Hostnames {
		host, err := loadbalancer.NewHostname(ctx, hostSpec.Name, &loadbalancer.HostnameArgs{
			LoadBalancerId: lb.ID(),
			Name:           pulumi.String(hostSpec.Name),
			Hostname:       pulumi.String(hostSpec.Hostname),
		}, pulumiOciOpt(provider), pulumi.Parent(lb))
		if err != nil {
			return nil, fmt.Errorf("failed to create hostname %s: %w", hostSpec.Name, err)
		}
		created = append(created, host)
	}

	return created, nil
}
