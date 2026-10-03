package setdeploy

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/manifestgraph"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// TestNodeWorkspaceDir_RefusesUnlawfulSlug: the slug and env are joined into
// a filesystem path, so one that is not a lawful slug is refused before any
// directory is touched.
func TestNodeWorkspaceDir_RefusesUnlawfulSlug(t *testing.T) {
	kind := catalogkind.CatalogKind_KubernetesNamespace
	for _, id := range []manifestgraph.Identity{
		{Kind: kind, Slug: "../escape"},
		{Kind: kind, Slug: "a_b"},
		{Kind: kind, Slug: "example.com"},
		{Kind: kind, Slug: ""},
		{Kind: kind, Slug: "fine", Env: "../prod"},
		{Kind: kind, Slug: "fine", Env: "prod_1"},
	} {
		_, err := nodeWorkspaceDir(id)
		if err == nil {
			t.Fatalf("%+v must be refused", id)
		}
		if !strings.Contains(err.Error(), manifestgraph.SlugRule) {
			t.Fatalf("%+v refusal must state the slug rule, got %v", id, err)
		}
	}
}
