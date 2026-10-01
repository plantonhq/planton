package manifestgraph

import (
	"testing"

	"github.com/plantonhq/planton/shared"
	"github.com/stretchr/testify/assert"
)

// TestGenerateSlug pins the port to the platform's ApiRequestResourceSlugGenerator:
// this table is copied from its Java suite. Node identity on both lanes
// derives through these exact rules, so a divergence here is an identity
// divergence everywhere.
func TestGenerateSlug(t *testing.T) {
	cases := []struct{ name, want string }{
		{"Sandeep's AWS", "sandeep-s-aws"},
		{"Sandeep’s AWS", "sandeep-s-aws"},
		{"My Resource Name", "my-resource-name"},
		{"Test@Resource#123", "test-resource-123"},
		{"Too   Many   Spaces", "too-many-spaces"},
		{"test---multiple---hyphens", "test-multiple-hyphens"},
		{"-test-name-", "test-name"},
		{"  test name  ", "test-name"},
		{"", ""},
		{"   ", ""},
		{"\t", ""},
		{"\n", ""},
		{"Café Résumé", "cafe-resume"},
		{"test_with_underscores", "test-with-underscores"},
		{"example.com", "example-com"},
		{"Resource 123 Test", "resource-123-test"},
		{"MixedCASE Name", "mixedcase-name"},
		{"Simple Name", "simple-name"},
		{"Name With 123", "name-with-123"},
		{"UPPERCASE", "uppercase"},
		{"lowercase", "lowercase"},
		{"a", "a"},
		{"123", "123"},
		{"test-already-hyphenated", "test-already-hyphenated"},
		{"DB_PASSWORD", "db-password"},
		{"_sip._tcp", "sip-tcp"},
		{"İstanbul", "istanbul"},
		{"Straße 9", "stra-e-9"},
		{"@#$%^&*()", ""},
		{"Test 🚀 Rocket", "test-rocket"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, GenerateSlug(c.name), "name %q", c.name)
	}
}

// TestGenerateSlug_AlwaysLawful: a name that keeps a letter or digit always
// generates a lawful slug, whatever else it carries.
func TestGenerateSlug_AlwaysLawful(t *testing.T) {
	names := []string{
		"DB_PASSWORD", "_sip._tcp", "a__b", "Straße 9", "x.-_y", "--Ünïcode--",
		"Test 🚀 Rocket", "aws.profile.Prod_2", "Sandeep’s AWS", "  a  b  ",
	}
	for _, name := range names {
		slug := GenerateSlug(name)
		assert.True(t, IsLawfulSlug(slug), "name %q generated unlawful slug %q", name, slug)
	}
}

func TestIsLawfulSlug(t *testing.T) {
	for _, s := range []string{"my-app-2", "a", "123", "sandeep-s-aws"} {
		assert.True(t, IsLawfulSlug(s), "%q", s)
	}
	for _, s := range []string{"", "My-App", "a_b", "example.com", "-a", "a-", "a--b", "../etc", "a b"} {
		assert.False(t, IsLawfulSlug(s), "%q", s)
	}
}

func TestResolveSlug_ExplicitSlugPassesThrough(t *testing.T) {
	meta := &shared.CloudResourceMetadata{Name: "My Shared Producer", Slug: "authored-slug"}
	assert.Equal(t, "authored-slug", ResolveSlug(meta))

	meta = &shared.CloudResourceMetadata{Name: "My Shared Producer"}
	assert.Equal(t, "my-shared-producer", ResolveSlug(meta))

	assert.Equal(t, "", ResolveSlug(nil))
}
