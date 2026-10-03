package module

import (
	azurecosmosdbmongocollectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbmongocollection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbMongoCollection *azurecosmosdbmongocollectionv1alpha1.AzureCosmosdbMongoCollection
	MongoDatabaseId              string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecosmosdbmongocollectionv1alpha1.AzureCosmosdbMongoCollectionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureCosmosdbMongoCollection = iacInput.Target
	locals.MongoDatabaseId = iacInput.Target.Spec.MongoDatabaseId.GetValue()

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.

	return locals
}
