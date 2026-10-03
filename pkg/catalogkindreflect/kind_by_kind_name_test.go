package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

func TestKindByKindName(t *testing.T) {
	tests := []struct {
		name     string
		kindName string
		want     catalogkind.CatalogKind
		wantErr  bool
	}{
		{
			name:     "resolves by kind name",
			kindName: "AwsVpc",
			want:     catalogkind.CatalogKind_AwsVpc,
			wantErr:  false,
		},
		{
			name:     "resolves kubernetes kind",
			kindName: "KubernetesDeployment",
			want:     catalogkind.CatalogKind_KubernetesDeployment,
			wantErr:  false,
		},
		{
			name:     "exact match only - lowercase does not resolve",
			kindName: "awsvpc",
			want:     catalogkind.CatalogKind_unspecified,
			wantErr:  true,
		},
		{
			name:     "unknown kind name",
			kindName: "NoSuchKind",
			want:     catalogkind.CatalogKind_unspecified,
			wantErr:  true,
		},
		{
			name:     "empty kind name",
			kindName: "",
			want:     catalogkind.CatalogKind_unspecified,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KindByKindName(tt.kindName)
			if (err != nil) != tt.wantErr {
				t.Errorf("KindByKindName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KindByKindName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The registry contract: every kind resolves to exactly one name and one id
// prefix. A collision makes resolution ambiguous, so index construction must
// succeed against the live registry — this test is the CI gate that catches a
// duplicate the moment it is introduced.
func TestKindRegistryIsUnambiguous(t *testing.T) {
	if _, err := buildKindNameIndex(); err != nil {
		t.Errorf("kind name registry has ambiguous entries: %v", err)
	}
	if _, err := buildKindIdPrefixIndex(); err != nil {
		t.Errorf("kind id-prefix registry has ambiguous entries: %v", err)
	}
}

// Enum value names must also stay canonically unique: tolerant resolution
// surfaces compare them lowercased with separators stripped, so two enum
// names differing only in case or separators would be ambiguous there even
// though protobuf accepts them as distinct identifiers.
func TestEnumValueNamesAreCanonicallyUnique(t *testing.T) {
	seen := make(map[string]catalogkind.CatalogKind)
	for _, kind := range KindsList() {
		canonicalName := canonicalKindName(kind.String())
		if owner, taken := seen[canonicalName]; taken {
			t.Errorf("enum value names %s and %s collide canonically (%q)", owner, kind, canonicalName)
			continue
		}
		seen[canonicalName] = kind
	}
}

func TestKindsListIsSortedAndStable(t *testing.T) {
	first := KindsList()
	if len(first) == 0 {
		t.Fatal("KindsList() returned no kinds")
	}
	for i := 1; i < len(first); i++ {
		if first[i-1] >= first[i] {
			t.Fatalf("KindsList() not sorted by enum number: %v before %v", first[i-1], first[i])
		}
	}
	second := KindsList()
	if len(first) != len(second) {
		t.Fatalf("KindsList() length changed between calls: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("KindsList() order changed between calls at index %d: %v vs %v", i, first[i], second[i])
		}
	}
}
