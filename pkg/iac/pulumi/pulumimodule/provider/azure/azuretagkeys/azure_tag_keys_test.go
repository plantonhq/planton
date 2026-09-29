package azuretagkeys

import "testing"

// The keys are live tag names on deployed Azure resources, and every OpenTofu
// module spells them as these exact literals. Changing one here would re-tag
// the Pulumi-deployed fleet and split it from the OpenTofu-deployed one.
func TestIdentityTagKeysMatchTheOpenTofuModules(t *testing.T) {
	want := map[string]string{
		"Resource":     "resource",
		"ResourceName": "resource_name",
		"ResourceKind": "resource_kind",
		"ResourceId":   "resource_id",
		"Organization": "organization",
		"Environment":  "environment",
	}
	got := map[string]string{
		"Resource":     Resource,
		"ResourceName": ResourceName,
		"ResourceKind": ResourceKind,
		"ResourceId":   ResourceId,
		"Organization": Organization,
		"Environment":  Environment,
	}
	for name, key := range want {
		if got[name] != key {
			t.Errorf("azuretagkeys.%s = %q, want %q", name, got[name], key)
		}
	}
}
