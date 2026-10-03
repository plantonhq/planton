// Package importmap loads and resolves the import-ID recipes that make state
// import derivable instead of hand-assembled.
//
// The knowledge is deliberately split in two tiers:
//
//   - Provider tier (ProviderImportCatalog, one per provider at
//     catalog/{provider}/aa_import/catalog.yaml): for each
//     IaC resource type, the import-ID FORMAT the engines expect --
//     "{bucket}", "{vpc_id}", "{bucket}:{intelligent_tiering_name}". Stable,
//     shared by every kind of the provider.
//
//   - Kind tier (CatalogKindImportMap, one per kind at
//     {kind}/v1/iac/import-map.yaml): for each {placeholder} the formats
//     reference, WHERE the value comes from -- the resource's metadata.name, a
//     spec field, an output, a pasted ARN's part, or the enumerated
//     address's own instance key -- plus "where to find this" guidance for the
//     values only the user can supply.
//
// Resource ADDRESSES are never authored anywhere: the engine enumerates real
// addresses per spec at import time (a read-only preview lists them), which
// handles module-constructed names, repeated (for_each/count) resources, and
// conditional resources by construction.
//
// Correctness is machine-proven, not reviewed: the offline conformance test in
// this package validates every recipe against the module sources it maps, and
// the E2E import round-trip (deploy a fixture, set state aside, re-import
// blind through these recipes, plan must show zero diff) proves the IDs
// against the real cloud.
package importmap
