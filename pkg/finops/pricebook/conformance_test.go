package pricebook

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var (
	decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	datePattern    = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	slugPattern    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// TestPriceBookConformance holds every price book to its contract, offline:
//
//  1. The document parses strictly, names its provider (metadata.name
//     equals the filename), and the provider exists under catalog/.
//  2. Entry slugs are well-formed and unique within the book -- estimate
//     models reference entries by slug, so a duplicate would make a
//     reference ambiguous.
//  3. Every entry carries the full price identity: service_name,
//     pricing_unit, region (a region code or "global"), currency, a plain
//     decimal unit price, a source URL, and a YYYY-MM-DD retrieval date.
//     A price without a source and date is a rumor, not a fact.
//  4. An entry's selector, when present, matches the book's own provider
//     (an aws_selector belongs only in the aws book, and so on -- the
//     oneof already guarantees at most one selector per entry) and is
//     complete enough to re-fetch deterministically: AWS needs offer code,
//     region code, and usage type; Azure needs a meter identity (a meter
//     ID, or a service plus at least one meter-identifying name filter);
//     GCP needs the service ID and SKU ID.
//  5. An entry's attributes, when present, carry non-empty keys and
//     values, and no two entries wear the same lookup identity
//     (service_name, pricing_unit, region, currency, attributes) -- a
//     value-keyed lookup must resolve to exactly one entry, so a
//     duplicate identity would turn every match into an ambiguity
//     refusal.
//
// Whether entries are actually referenced (and agree with kind cost
// profiles on units) is the estimate generator's cross-artifact check --
// this gate keeps each book internally sound.
func TestPriceBookConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	providers, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering price books: %v", err)
	}
	if len(providers) == 0 {
		t.Skip("no price books authored yet")
	}

	for _, provider := range providers {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			book, err := Load(root, provider)
			if err != nil {
				t.Fatalf("price book: %v", err)
			}
			if book.GetKind() != "PriceBook" {
				t.Fatalf("kind is %q, want PriceBook", book.GetKind())
			}
			if book.GetMetadata().GetName() != provider {
				t.Errorf("metadata.name is %q, want %q (the filename is the provider's identity)",
					book.GetMetadata().GetName(), provider)
			}
			if _, err := os.Stat(filepath.Join(root, "catalog", provider)); err != nil {
				t.Errorf("price book names provider %q, which exists nowhere under catalog/", provider)
			}
			if len(book.GetSpec().GetEntries()) == 0 {
				t.Fatal("price book declares no entries")
			}

			seen := map[string]bool{}
			identities := map[string]string{}
			for _, entry := range book.GetSpec().GetEntries() {
				name := entry.GetName()
				if !slugPattern.MatchString(name) {
					t.Errorf("entry name %q is not a lowercase hyphenated slug", name)
				}
				if seen[name] {
					t.Errorf("entry %q declared more than once -- estimate models reference entries by slug", name)
				}
				seen[name] = true

				if strings.TrimSpace(entry.GetServiceName()) == "" {
					t.Errorf("entry %q has no service_name", name)
				}
				if strings.TrimSpace(entry.GetPricingUnit()) == "" {
					t.Errorf("entry %q has no pricing_unit", name)
				}
				if strings.TrimSpace(entry.GetRegion()) == "" {
					t.Errorf("entry %q has no region -- pin the region code, or %q for region-independent prices", name, GlobalRegion)
				}
				if strings.TrimSpace(entry.GetCurrency()) == "" {
					t.Errorf("entry %q has no currency", name)
				}
				if !decimalPattern.MatchString(entry.GetListUnitPrice()) {
					t.Errorf("entry %q list_unit_price %q is not a plain decimal string (money is never a YAML float)",
						name, entry.GetListUnitPrice())
				}
				if strings.TrimSpace(entry.GetPriceSource()) == "" {
					t.Errorf("entry %q has no price_source", name)
				}
				if !datePattern.MatchString(entry.GetRetrievedOn()) {
					t.Errorf("entry %q retrieved_on %q is not a YYYY-MM-DD date -- a dated price is a fact, an undated price is a rumor",
						name, entry.GetRetrievedOn())
				}

				if attributes := entry.GetAttributes(); len(attributes) > 0 {
					keys := make([]string, 0, len(attributes))
					for key, value := range attributes {
						if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
							t.Errorf("entry %q has an attribute with an empty key or value", name)
						}
						keys = append(keys, key)
					}
					sort.Strings(keys)
					identity := entry.GetServiceName() + "|" + entry.GetPricingUnit() + "|" + entry.GetRegion() + "|" + entry.GetCurrency()
					for _, key := range keys {
						identity += "|" + key + "=" + attributes[key]
					}
					if holder, taken := identities[identity]; taken {
						t.Errorf("entries %q and %q wear the same lookup identity -- a value-keyed lookup could never resolve between them", holder, name)
					}
					identities[identity] = name
				}

				if selector := entry.GetAwsSelector(); selector != nil {
					if provider != "aws" {
						t.Errorf("entry %q carries an aws_selector in the %q book", name, provider)
					}
					if strings.TrimSpace(selector.GetOfferCode()) == "" {
						t.Errorf("entry %q aws_selector has no offer_code", name)
					}
					if strings.TrimSpace(selector.GetRegionCode()) == "" {
						t.Errorf("entry %q aws_selector has no region_code", name)
					}
					if strings.TrimSpace(selector.GetUsageType()) == "" {
						t.Errorf("entry %q aws_selector has no usage_type", name)
					}
				}

				if selector := entry.GetAzureSelector(); selector != nil {
					if provider != "azure" {
						t.Errorf("entry %q carries an azure_selector in the %q book", name, provider)
					}
					hasMeterIdentity := strings.TrimSpace(selector.GetMeterId()) != "" ||
						strings.TrimSpace(selector.GetArmSkuName()) != "" ||
						strings.TrimSpace(selector.GetMeterName()) != "" ||
						strings.TrimSpace(selector.GetProductName()) != ""
					if !hasMeterIdentity {
						t.Errorf("entry %q azure_selector identifies no meter -- set meter_id, or arm_sku_name/meter_name/product_name", name)
					}
					if strings.TrimSpace(selector.GetMeterId()) == "" && strings.TrimSpace(selector.GetServiceName()) == "" {
						t.Errorf("entry %q azure_selector filters by name without a service_name -- name filters only isolate within a service", name)
					}
				}

				if selector := entry.GetGcpSelector(); selector != nil {
					if provider != "gcp" {
						t.Errorf("entry %q carries a gcp_selector in the %q book", name, provider)
					}
					if strings.TrimSpace(selector.GetServiceId()) == "" {
						t.Errorf("entry %q gcp_selector has no service_id", name)
					}
					if strings.TrimSpace(selector.GetSkuId()) == "" {
						t.Errorf("entry %q gcp_selector has no sku_id", name)
					}
				}
			}
		})
	}
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}
