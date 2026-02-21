package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudcontainerregistryv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudcontainerregistry/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cr"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createNamespaces(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	instance *cr.RegistryEnterpriseInstance,
	namespaces []*alicloudcontainerregistryv1.AliCloudContainerRegistryNamespace,
) (pulumi.MapOutput, error) {
	nsIdMap := pulumi.Map{}

	for _, ns := range namespaces {
		resourceName := fmt.Sprintf("ns-%s", ns.Name)

		created, err := cs.NewRegistryEnterpriseNamespace(ctx, resourceName, &cs.RegistryEnterpriseNamespaceArgs{
			InstanceId:        instance.ID(),
			Name:              pulumi.String(ns.Name),
			AutoCreate:        pulumi.Bool(namespaceAutoCreate(ns)),
			DefaultVisibility: pulumi.String(namespaceDefaultVisibility(ns)),
		}, pulumi.Provider(provider), pulumi.Parent(instance))
		if err != nil {
			return pulumi.Map{}.ToMapOutput(), errors.Wrapf(err, "failed to create namespace %s", ns.Name)
		}

		nsIdMap[ns.Name] = created.ID()
	}

	return nsIdMap.ToMapOutput(), nil
}
