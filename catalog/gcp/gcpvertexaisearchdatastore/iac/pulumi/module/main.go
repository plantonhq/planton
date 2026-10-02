package module

import (
	"github.com/pkg/errors"
	gcpvertexaisearchdatastorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchdatastore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpvertexaisearchdatastorev1alpha1.GcpVertexAiSearchDataStoreStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdDataStore, err := dataStore(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai search data store")
	}

	createdSchema, err := schema(ctx, locals, gcpProvider, createdDataStore)
	if err != nil {
		return errors.Wrap(err, "failed to create data store schema")
	}

	targetSiteNames, err := targetSites(ctx, locals, gcpProvider, createdDataStore)
	if err != nil {
		return errors.Wrap(err, "failed to create data store target sites")
	}

	sitemapNames, err := sitemaps(ctx, locals, gcpProvider, createdDataStore)
	if err != nil {
		return errors.Wrap(err, "failed to create data store sitemaps")
	}

	ctx.Export(OpName, createdDataStore.Name)
	ctx.Export(OpDataStoreId, createdDataStore.DataStoreId)
	ctx.Export(OpLocation, createdDataStore.Location)
	ctx.Export(OpDefaultSchemaId, createdDataStore.DefaultSchemaId)
	if createdSchema != nil {
		ctx.Export(OpSchemaName, createdSchema.Name)
	} else {
		ctx.Export(OpSchemaName, pulumi.String(""))
	}
	ctx.Export(OpTargetSiteNames, targetSiteNames.ToStringArrayOutput())
	ctx.Export(OpSitemapNames, sitemapNames.ToStringArrayOutput())
	return nil
}
