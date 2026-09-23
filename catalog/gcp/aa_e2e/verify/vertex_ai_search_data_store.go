package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiSearchDataStoreVerifier probes a Vertex AI Search data store and
// its folded schema, target sites, and sitemaps through the Discovery
// Engine REST API. Discovery Engine resources carry no labels, so the
// verifier asserts existence and the display name instead of the
// attribution-label canary, plus every companion the module exported.
type vertexAiSearchDataStoreVerifier struct{}

// IDOutputKey is the store's full resource name
// (projects/{p}/locations/{l}/collections/default_collection/dataStores/{id}).
func (v *vertexAiSearchDataStoreVerifier) IDOutputKey() string { return "name" }

// VerifyExists confirms the store reads back under the exported name with
// the exported id and default schema, and that every exported schema,
// target site, and sitemap exists.
func (v *vertexAiSearchDataStoreVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	store, _, err := discoveryEngineGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "vertex ai search data store %s not found after deploy", name)
	}
	if got := outputs["data_store_id"]; got != "" && lastPathSegment(stringField(store, "name")) != got {
		return errors.Errorf("vertex ai search data store %s data_store_id output %q does not match live name %q", name, got, lastPathSegment(stringField(store, "name")))
	}
	if stringField(store, "displayName") == "" {
		return errors.Errorf("vertex ai search data store %s has no display name after deploy", name)
	}
	if got := outputs["default_schema_id"]; got != "" && stringField(store, "defaultSchemaId") != got {
		return errors.Errorf("vertex ai search data store %s default_schema_id output %q does not match live %q", name, got, stringField(store, "defaultSchemaId"))
	}

	if schemaName := outputs["schema_name"]; schemaName != "" {
		if _, _, err := discoveryEngineGet(ctx, svc, schemaName); err != nil {
			return errors.Wrapf(err, "vertex ai search schema %s not found after deploy", schemaName)
		}
	}
	for _, siteName := range listOutput(outputs, "target_site_names") {
		if _, _, err := discoveryEngineGet(ctx, svc, siteName); err != nil {
			return errors.Wrapf(err, "vertex ai search target site %s not found after deploy", siteName)
		}
	}
	for _, sitemapName := range listOutput(outputs, "sitemap_names") {
		if _, _, err := discoveryEngineGet(ctx, svc, sitemapName); err != nil {
			return errors.Wrapf(err, "vertex ai search sitemap %s not found after deploy", sitemapName)
		}
	}
	return nil
}

// VerifyAbsent confirms the store is gone (its schema, target sites, and
// sitemaps cannot outlive it).
func (v *vertexAiSearchDataStoreVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := discoveryEngineGet(ctx, svc, name)
	return discoveryEngineAbsent("vertex ai search data store", name, status, err)
}
