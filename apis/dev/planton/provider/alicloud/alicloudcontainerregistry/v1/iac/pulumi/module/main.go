package module

import (
	"github.com/pkg/errors"
	alicloudcontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudcontainerregistry/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudcontainerregistryv1.AliCloudContainerRegistryStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudContainerRegistry.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	instanceArgs := &cr.RegistryEnterpriseInstanceArgs{
		InstanceName: pulumi.String(spec.InstanceName),
		InstanceType: pulumi.String(spec.InstanceType),
		PaymentType:  pulumi.String(paymentType(spec)),
	}

	if spec.Period > 0 {
		instanceArgs.Period = pulumi.IntPtr(int(spec.Period))
	}

	if spec.Password != "" {
		instanceArgs.Password = pulumi.String(spec.Password)
	}

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = optionalString(spec.ResourceGroupId)
	}

	instance, err := cr.NewRegistryEnterpriseInstance(ctx, spec.InstanceName, instanceArgs,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create container registry instance %s", spec.InstanceName)
	}

	namespaceIds, err := createNamespaces(ctx, alicloudProvider, instance, spec.Namespaces)
	if err != nil {
		return err
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpInstanceName, instance.InstanceName)

	publicEndpoint := extractEndpointDomain(instance, "internet")
	vpcEndpoint := extractEndpointDomain(instance, "vpc")
	ctx.Export(OpPublicEndpoint, publicEndpoint)
	ctx.Export(OpVpcEndpoint, vpcEndpoint)
	ctx.Export(OpNamespaceIds, namespaceIds)

	return nil
}

// extractEndpointDomain finds the first domain in the instance_endpoints list
// matching the given endpoint type (e.g., "internet", "vpc").
func extractEndpointDomain(instance *cr.RegistryEnterpriseInstance, endpointType string) pulumi.StringOutput {
	return instance.InstanceEndpoints.ApplyT(func(endpoints []cr.RegistryEnterpriseInstanceInstanceEndpoint) string {
		for _, ep := range endpoints {
			if ep.EndpointType == nil {
				continue
			}
			if *ep.EndpointType != endpointType {
				continue
			}
			if ep.Enable != nil && !*ep.Enable {
				continue
			}
			for _, d := range ep.Domains {
				if d.Domain != nil && *d.Domain != "" {
					return *d.Domain
				}
			}
		}
		return ""
	}).(pulumi.StringOutput)
}
