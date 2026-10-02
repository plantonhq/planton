package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// sitemaps creates one `google_discovery_engine_sitemap` per
// spec.sitemap_uris[] entry -- the sitemaps an advanced site search store
// reads. A sitemap is fully immutable, so a change replaces it. Names are
// exported in manifest order, as the Terraform module does.
func sitemaps(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	createdDataStore *discoveryengine.DataStore) (pulumi.StringArray, error) {
	spec := locals.GcpVertexAiSearchDataStore.Spec
	names := pulumi.StringArray{}

	for i, uri := range spec.SitemapUris {
		args := &discoveryengine.SitemapArgs{
			DataStoreId: createdDataStore.DataStoreId,
			Location:    pulumi.String(spec.Location),
			Uri:         pulumi.String(uri),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdSitemap, err := discoveryengine.NewSitemap(ctx,
			fmt.Sprintf("%s-sitemap-%d", locals.GcpVertexAiSearchDataStore.Metadata.Name, i), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdDataStore))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create sitemap %s", uri)
		}
		names = append(names, createdSitemap.Name)
	}
	return names, nil
}
