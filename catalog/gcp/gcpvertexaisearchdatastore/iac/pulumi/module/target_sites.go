package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// targetSites creates one `google_discovery_engine_target_site` per
// spec.target_sites[] entry -- the URL patterns a PUBLIC_WEBSITE store
// crawls (INCLUDE) or leaves out (EXCLUDE). Every field is immutable, so
// a change replaces that target site. Resource names are indexed by
// position, the same order the Terraform module keys its for_each by
// pattern, so both engines export the names in manifest order.
func targetSites(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	createdDataStore *discoveryengine.DataStore) (pulumi.StringArray, error) {
	spec := locals.GcpVertexAiSearchDataStore.Spec
	names := pulumi.StringArray{}

	for i, site := range spec.TargetSites {
		args := &discoveryengine.TargetSiteArgs{
			DataStoreId:        createdDataStore.DataStoreId,
			Location:           pulumi.String(spec.Location),
			ProvidedUriPattern: pulumi.String(site.ProvidedUriPattern),
			ExactMatch:         pulumi.BoolPtr(site.ExactMatch),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		// Empty lets Google default to INCLUDE.
		if site.Type != "" {
			args.Type = pulumi.String(site.Type)
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdSite, err := discoveryengine.NewTargetSite(ctx,
			fmt.Sprintf("%s-target-site-%d", locals.GcpVertexAiSearchDataStore.Metadata.Name, i), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdDataStore))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create target site %s", site.ProvidedUriPattern)
		}
		names = append(names, createdSite.Name)
	}
	return names, nil
}
