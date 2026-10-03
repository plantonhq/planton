// service_group_test.go — the service-group registry gates.
//
// WHY THIS EXISTS:
// Every kind of a grouped provider carries exactly one service_group on its
// kind_meta — the provider-console category catalog surfaces browse it under.
// The taxonomy only works if it is total and coherent: a kind without a group
// vanishes from grouped explorers, and a kind carrying another provider's
// group renders under the wrong console's taxonomy. These gates make both a
// committed-test failure, in both directions:
//
//   - every kind of a grouped provider carries a non-unspecified group
//     whose owning provider IS the kind's provider (coverage + coherence),
//   - kinds of providers without a service taxonomy carry none (strictness),
//   - every group enum value carries service_group_meta with a grouped
//     provider and a display name (label surfaces render from it).
//
// The grouped-provider set is pinned here deliberately: which providers HAVE
// a service taxonomy is a product-level decision, so growing or shrinking the
// set must show up as a conscious edit to this test, not as a silent side
// effect of enum churn.
package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

const unspecifiedServiceGroup = catalogkind.CatalogProviderServiceGroup_catalog_provider_service_group_unspecified

// groupedProviders pins the set of providers whose kinds carry service
// groups. Managed-service providers (auth0, openfga, vault) and the _test
// provider have no service taxonomy.
var groupedProviders = map[catalogkind.CatalogProvider]bool{
	catalogkind.CatalogProvider_aws:           true,
	catalogkind.CatalogProvider_azure:         true,
	catalogkind.CatalogProvider_gcp:           true,
	catalogkind.CatalogProvider_kubernetes:    true,
	catalogkind.CatalogProvider_cloudflare:    true,
	catalogkind.CatalogProvider_digital_ocean: true,
}

// TestServiceGroupExtensionNumberIsPinned guards the wire-level contract:
// downstream platforms read service_group_meta by number from bundle
// descriptors, so the number is load-bearing API surface (same guarantee as
// containment_decisions_test.go pins for its extensions).
func TestServiceGroupExtensionNumberIsPinned(t *testing.T) {
	if n := catalogkind.E_ServiceGroupMeta.TypeDescriptor().Number(); n != 81102 {
		t.Fatalf("service_group_meta extension number changed: got %d, want 81102", n)
	}
}

// TestServiceGroupValuesCarryCoherentMeta walks every CatalogProviderServiceGroup
// value: each must carry service_group_meta naming a grouped provider and a
// non-empty display name, and every grouped provider must declare at least
// one group (a provider with zero groups could never satisfy the coverage
// gate below).
func TestServiceGroupValuesCarryCoherentMeta(t *testing.T) {
	values := unspecifiedServiceGroup.Descriptor().Values()
	checked := 0
	providersWithGroups := map[catalogkind.CatalogProvider]bool{}
	for i := 0; i < values.Len(); i++ {
		valueDescriptor := values.Get(i)
		if valueDescriptor.Number() == 0 {
			continue
		}
		group := catalogkind.CatalogProviderServiceGroup(valueDescriptor.Number())
		meta, err := ServiceGroupMetaOf(group)
		if err != nil {
			t.Errorf("service group %s carries no service_group_meta: %v", valueDescriptor.Name(), err)
			continue
		}
		if !groupedProviders[meta.Provider] {
			t.Errorf("service group %s declares provider %s, which has no service taxonomy", valueDescriptor.Name(), meta.Provider)
		}
		if meta.DisplayName == "" {
			t.Errorf("service group %s has no display_name — label surfaces render from it", valueDescriptor.Name())
		}
		providersWithGroups[meta.Provider] = true
		checked++
	}
	if checked == 0 {
		t.Fatal("walked zero service-group values — the gate has gone vacuous")
	}
	for provider := range groupedProviders {
		if !providersWithGroups[provider] {
			t.Errorf("grouped provider %s declares no service groups", provider)
		}
	}
}

// TestKindServiceGroupsAreCoherent is the per-kind gate: coverage and
// provider-coherence for grouped providers, strictness for the rest.
func TestKindServiceGroupsAreCoherent(t *testing.T) {
	groupedKinds, ungroupedKinds := 0, 0
	for _, kind := range KindsList() {
		meta, err := KindMeta(kind)
		if err != nil {
			t.Errorf("%s: no kind_meta: %v", kind, err)
			continue
		}
		group := meta.ServiceGroup
		if !groupedProviders[meta.Provider] {
			if group != unspecifiedServiceGroup {
				t.Errorf("%s (provider %s) carries service_group %s, but its provider has no service taxonomy", kind, meta.Provider, group)
			}
			ungroupedKinds++
			continue
		}
		if group == unspecifiedServiceGroup {
			t.Errorf("%s (provider %s) carries no service_group — every grouped-provider kind needs exactly one", kind, meta.Provider)
			continue
		}
		groupMeta, err := ServiceGroupMetaOf(group)
		if err != nil {
			t.Errorf("%s carries service_group %s which has no service_group_meta: %v", kind, group, err)
			continue
		}
		if groupMeta.Provider != meta.Provider {
			t.Errorf("%s (provider %s) carries service_group %s, which belongs to provider %s", kind, meta.Provider, group, groupMeta.Provider)
			continue
		}
		groupedKinds++
	}
	if groupedKinds == 0 || ungroupedKinds == 0 {
		t.Fatalf("vacuous walk: grouped=%d ungrouped=%d — both classes must be exercised", groupedKinds, ungroupedKinds)
	}
}

// TestServiceGroupFixtures pins a few representative assignments so an
// accidental mass re-stamp cannot pass silently, plus the accessor's honest
// absence for an ungrouped provider.
func TestServiceGroupFixtures(t *testing.T) {
	fixtures := []struct {
		kind catalogkind.CatalogKind
		want catalogkind.CatalogProviderServiceGroup
	}{
		{catalogkind.CatalogKind_AwsEksCluster, catalogkind.CatalogProviderServiceGroup_aws_containers},
		{catalogkind.CatalogKind_AzureKeyVault, catalogkind.CatalogProviderServiceGroup_azure_security},
		{catalogkind.CatalogKind_GcpCloudSql, catalogkind.CatalogProviderServiceGroup_gcp_databases},
		{catalogkind.CatalogKind_KubernetesPostgres, catalogkind.CatalogProviderServiceGroup_kubernetes_open_source_software},
		{catalogkind.CatalogKind_CloudflareWorker, catalogkind.CatalogProviderServiceGroup_cloudflare_developer_platform},
		{catalogkind.CatalogKind_DigitalOceanDroplet, catalogkind.CatalogProviderServiceGroup_digital_ocean_compute},
		{catalogkind.CatalogKind_TestCatalogKindGeneric, unspecifiedServiceGroup},
	}
	for _, fixture := range fixtures {
		got, err := ServiceGroup(fixture.kind)
		if err != nil {
			t.Errorf("ServiceGroup(%s): %v", fixture.kind, err)
			continue
		}
		if got != fixture.want {
			t.Errorf("ServiceGroup(%s) = %s, want %s", fixture.kind, got, fixture.want)
		}
	}
}
