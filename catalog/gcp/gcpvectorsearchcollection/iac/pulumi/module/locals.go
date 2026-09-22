package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvectorsearchcollectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvectorsearchcollection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig         *gcpprovider.GcpProviderConfig
	GcpVectorSearchCollection *gcpvectorsearchcollectionv1alpha1.GcpVectorSearchCollection
	GcpLabels                 map[string]string

	// CollectionId is the collection's GCP id: spec.collection_id when set,
	// otherwise metadata.name -- the same fallback the Terraform module
	// applies in locals.tf.
	CollectionId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvectorsearchcollectionv1alpha1.GcpVectorSearchCollectionStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVectorSearchCollection = stackInput.Target

	locals.CollectionId = locals.GcpVectorSearchCollection.Spec.CollectionId
	if locals.CollectionId == "" {
		locals.CollectionId = locals.GcpVectorSearchCollection.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module. The same
	// set lands on the collection and on every index.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVectorSearchCollection.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.CollectionId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVectorSearchCollection.String())

	if locals.GcpVectorSearchCollection.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVectorSearchCollection.Metadata.Org
	}
	if locals.GcpVectorSearchCollection.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVectorSearchCollection.Metadata.Env
	}
	if locals.GcpVectorSearchCollection.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVectorSearchCollection.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
