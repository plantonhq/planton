package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureeventhubclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubclusterv1alpha1.AzureEventHubClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubCluster.Spec

	// The ARM sku is the composite string "Dedicated_{CUs}". The tier name
	// is a one-value constant -- Dedicated is the ONLY sku family Azure
	// sells for clusters -- so the module composes the string from the
	// capacity count instead of surfacing it as configuration. Presence-
	// guarded to 1 CU (Azure's entry size): IaC inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	capacityUnits := int32(1)
	if spec.CapacityUnits != nil {
		capacityUnits = spec.GetCapacityUnits()
	}

	// The dedicated Event Hubs cluster: single-tenant capacity units that
	// namespaces are placed on via their dedicated_cluster_id reference.
	// Many namespaces share one cluster, which is why the cluster is its
	// own resource rather than a namespace property.
	//
	// Cost: dedicated clusters bill per capacity unit per hour at
	// dedicated-tier rates -- the most expensive resource in the Event
	// Hubs family. Provision one deliberately.
	//
	// Lifecycle: Azure FORBIDS deleting a cluster for 4 HOURS after
	// creation (the deletion moratorium). A destroy inside that window
	// retries until Azure permits the delete -- expect a destroy of a
	// young cluster to take hours by the service's own rule.
	clusterArgs := &eventhub.ClusterArgs{
		// ForceNew: renaming replaces the cluster (subject to the 4-hour
		// deletion moratorium above).
		Name:              pulumi.String(spec.ClusterName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		SkuName:           pulumi.String(fmt.Sprintf("Dedicated_%d", capacityUnits)),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	createdCluster, err := eventhub.NewCluster(ctx,
		spec.ClusterName,
		clusterArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hubs cluster %s", spec.ClusterName)
	}

	// Export outputs. cluster_id is what an AzureEventHubNamespace's
	// dedicated_cluster_id references to place the namespace on this
	// cluster.
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpClusterName, createdCluster.Name)

	return nil
}
