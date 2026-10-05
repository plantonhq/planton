//go:build !codegen
// +build !codegen

package importmap

import (
	"errors"
	"reflect"
	"testing"

	awsecrrepov1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecrrepo/v1alpha1"
	kindv1 "github.com/plantonhq/planton/iac/catalogkindimportmap/v1"
)

func TestPlaceholders(t *testing.T) {
	cases := []struct {
		idFormat string
		want     []string
	}{
		{"{bucket}", []string{"bucket"}},
		{"{bucket}:{intelligent_tiering_name}", []string{"bucket", "intelligent_tiering_name"}},
		{"{vpc_id}_{association_id}", []string{"vpc_id", "association_id"}},
		// Optional placeholders are still placeholders -- they need a
		// declaration and a resolution attempt.
		{"name:{table_name}/index:{index_name?}/{account_id}",
			[]string{"table_name", "index_name", "account_id"}},
		{"literal-id", nil},
	}
	for _, tc := range cases {
		got := Placeholders(tc.idFormat)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Placeholders(%q) = %v, want %v", tc.idFormat, got, tc.want)
		}
	}
}

func TestRequiredPlaceholders(t *testing.T) {
	got := RequiredPlaceholders("name:{table_name}/index:{index_name?}/{account_id}")
	want := []string{"table_name", "account_id"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RequiredPlaceholders = %v, want %v", got, want)
	}

	// A bracketed segment group's placeholder is optional by construction --
	// the group (delimiters included) is dropped when it does not resolve.
	got = RequiredPlaceholders("{api_version}//{kind}//{name}[//{namespace}]")
	want = []string{"api_version", "kind", "name"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RequiredPlaceholders with segment group = %v, want %v", got, want)
	}
}

func TestRenderID(t *testing.T) {
	id, err := RenderID("{bucket}:{name}", map[string]string{"bucket": "my-bucket", "name": "archive"})
	if err != nil {
		t.Fatalf("RenderID: %v", err)
	}
	if id != "my-bucket:archive" {
		t.Fatalf("RenderID = %q, want my-bucket:archive", id)
	}

	// A missing value must be an ERROR, never an empty substitution -- a
	// partially-filled ID would import the wrong resource.
	if _, err := RenderID("{bucket}:{name}", map[string]string{"bucket": "my-bucket"}); err == nil {
		t.Fatal("RenderID with a missing value did not error")
	}

	// An OPTIONAL placeholder renders as the empty string when unresolved --
	// the provider-documented shape for variants like table-level DynamoDB
	// contributor insights ("name:tbl/index:/123456789012").
	id, err = RenderID("name:{table_name}/index:{index_name?}/{account_id}",
		map[string]string{"table_name": "tbl", "account_id": "123456789012"})
	if err != nil {
		t.Fatalf("RenderID with optional placeholder: %v", err)
	}
	if id != "name:tbl/index:/123456789012" {
		t.Fatalf("RenderID = %q, want name:tbl/index:/123456789012", id)
	}

	// A bracketed segment group renders WITH its literal delimiters when the
	// placeholder resolves (a namespaced kubectl_manifest CR)...
	composed := "{api_version}//{kind}//{name}[//{namespace}]"
	id, err = RenderID(composed, map[string]string{
		"api_version": "cert-manager.io/v1", "kind": "Certificate",
		"name": "app-cert", "namespace": "prod",
	})
	if err != nil {
		t.Fatalf("RenderID with resolved segment group: %v", err)
	}
	if id != "cert-manager.io/v1//Certificate//app-cert//prod" {
		t.Fatalf("RenderID = %q, want cert-manager.io/v1//Certificate//app-cert//prod", id)
	}

	// ...and disappears wholesale when it does not (a cluster-scoped CR):
	// the kubectl importer rejects a trailing "//", so the delimiter must
	// vanish with the value.
	id, err = RenderID(composed, map[string]string{
		"api_version": "cert-manager.io/v1", "kind": "ClusterIssuer",
		"name": "letsencrypt",
	})
	if err != nil {
		t.Fatalf("RenderID with unresolved segment group: %v", err)
	}
	if id != "cert-manager.io/v1//ClusterIssuer//letsencrypt" {
		t.Fatalf("RenderID = %q, want cert-manager.io/v1//ClusterIssuer//letsencrypt", id)
	}
}

func TestParseTofuAddress(t *testing.T) {
	cases := []struct {
		address string
		resType string
		name    string
		key     string
		ok      bool
	}{
		{"aws_s3_bucket.this", "aws_s3_bucket", "this", "", true},
		{"aws_s3_bucket_intelligent_tiering_configuration.this[\"archive\"]",
			"aws_s3_bucket_intelligent_tiering_configuration", "this", "archive", true},
		{"aws_vpc_ipv4_cidr_block_association.secondary[\"10.1.0.0/16\"]",
			"aws_vpc_ipv4_cidr_block_association", "secondary", "10.1.0.0/16", true},
		// A count index is positional, never identity -- it must not leak
		// into an import ID as if it were a for_each discriminator.
		{"aws_instance.web[0]", "aws_instance", "web", "", true},
		{"module.nested.aws_s3_bucket.this", "", "", "", false},
		{"data.aws_ami.ubuntu", "", "", "", false},
	}
	for _, tc := range cases {
		resType, name, key, ok := ParseTofuAddress(tc.address)
		if ok != tc.ok || resType != tc.resType || name != tc.name || key != tc.key {
			t.Errorf("ParseTofuAddress(%q) = (%q,%q,%q,%v), want (%q,%q,%q,%v)",
				tc.address, resType, name, key, ok, tc.resType, tc.name, tc.key, tc.ok)
		}
	}
}

func TestResolveValues(t *testing.T) {
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "bucket",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromMetadataName{FromMetadataName: true}},
					},
				},
				{
					Name: "repository_name",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromSpecField{FromSpecField: "repository_name"}},
					},
				},
				{
					Name: "vpc_id",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromOutput{FromOutput: "vpc_id"}},
						{Source: &kindv1.ImportValueDerivation_FromArnPart{FromArnPart: "resource_id"}},
					},
				},
				{
					Name: "tiering_name",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromAddressKey{FromAddressKey: true}},
					},
				},
				{Name: "association_id", WhereToFind: "describe-vpcs"},
				{
					// A literal derivation: constants the module hardcodes
					// (a typed-CR module's apiVersion/kind).
					Name: "api_version",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_Literal{Literal: "cert-manager.io/v1"}},
					},
				},
				{
					// The prefix mirror of from_metadata_name_suffix:
					// provider-composed identifiers built by prefixing the
					// parent's name (App Auto Scaling's "table/<table_name>").
					Name: "scaling_resource_id",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromMetadataNamePrefix{FromMetadataNamePrefix: "table/"}},
					},
				},
			},
		},
	}

	rctx := ResolveContext{
		MetadataName: "my-bucket",
		Spec:         &awsecrrepov1alpha1.AwsEcrRepoSpec{RepositoryName: "team/app"},
		Outputs:      map[string]string{},
		AddressKey:   "archive",
		ArnParts:     map[string]string{"resource_id": "vpc-0abc"},
	}

	resolved, unresolved := ResolveValues(
		m,
		[]string{"bucket", "repository_name", "vpc_id", "tiering_name", "association_id", "api_version", "scaling_resource_id"},
		rctx,
	)

	want := map[string]string{
		"bucket":          "my-bucket",
		"repository_name": "team/app",
		// output empty -> falls through to the ARN part.
		"vpc_id":              "vpc-0abc",
		"tiering_name":        "archive",
		"api_version":         "cert-manager.io/v1",
		"scaling_resource_id": "table/my-bucket",
	}
	if !reflect.DeepEqual(resolved, want) {
		t.Errorf("resolved = %v, want %v", resolved, want)
	}
	if !reflect.DeepEqual(unresolved, []string{"association_id"}) {
		t.Errorf("unresolved = %v, want [association_id]", unresolved)
	}
}

func TestResolveValues_FromOutputKeyedByAddress(t *testing.T) {
	// Cloud-generated per-instance IDs of keyed satellites: the module
	// exports a map keyed by the SAME key as the resource's for_each
	// instances, and the enumerated address selects the entry. Map keys may
	// themselves contain dots (a CIDR) -- the lookup composes the flattened
	// dot-path string exactly, never re-parses it.
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "cidr_association_id",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromOutputKeyedByAddress{
							FromOutputKeyedByAddress: "secondary_ipv4_cidr_association_ids",
						}},
					},
					WhereToFind: "describe-vpcs",
				},
			},
		},
	}

	outputs := map[string]string{
		"secondary_ipv4_cidr_association_ids.10.1.0.0/16": "vpc-cidr-assoc-0abc",
		"secondary_ipv4_cidr_association_ids.ipam-1":      "vpc-cidr-assoc-0def",
	}

	resolved, unresolved := ResolveValues(m, []string{"cidr_association_id"}, ResolveContext{
		Outputs:    outputs,
		AddressKey: "10.1.0.0/16",
	})
	if len(unresolved) != 0 || resolved["cidr_association_id"] != "vpc-cidr-assoc-0abc" {
		t.Errorf("dotted-key entry: resolved = %v, unresolved = %v", resolved, unresolved)
	}

	resolved, unresolved = ResolveValues(m, []string{"cidr_association_id"}, ResolveContext{
		Outputs:    outputs,
		AddressKey: "ipam-1",
	})
	if len(unresolved) != 0 || resolved["cidr_association_id"] != "vpc-cidr-assoc-0def" {
		t.Errorf("ipam-keyed entry: resolved = %v, unresolved = %v", resolved, unresolved)
	}

	// An address without an instance key (a singular resource reaching this
	// arm by mistake) resolves empty and falls back to unresolved -- the
	// ask-the-user contract, never a wrong composed lookup.
	resolved, unresolved = ResolveValues(m, []string{"cidr_association_id"}, ResolveContext{
		Outputs: outputs,
	})
	if len(resolved) != 0 || !reflect.DeepEqual(unresolved, []string{"cidr_association_id"}) {
		t.Errorf("empty address key: resolved = %v, unresolved = %v", resolved, unresolved)
	}
}

func TestResolveValues_TofuResourceNameScoping(t *testing.T) {
	// A module with SEVERAL resources of one type whose placeholder carries a
	// different value per resource (a control-plane module installing multiple
	// Helm releases): the declaration scoped to the address's logical name
	// wins; addresses without a scoped declaration fall back to the unscoped
	// one; scoped declarations for OTHER resources are never consulted.
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "release_name",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_Literal{Literal: "istio-base"}},
					},
				},
				{
					Name:             "release_name",
					TofuResourceName: "istiod",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromOutput{FromOutput: "istiod_service_name"}},
					},
				},
			},
		},
	}

	outputs := map[string]string{"istiod_service_name": "istiod"}

	// Scoped address: the istiod declaration wins.
	resolved, unresolved := ResolveValues(m, []string{"release_name"},
		ResolveContext{Outputs: outputs, LogicalName: "istiod"})
	if len(unresolved) != 0 || resolved["release_name"] != "istiod" {
		t.Errorf("scoped: resolved = %v, unresolved = %v; want release_name=istiod", resolved, unresolved)
	}

	// Unscoped address: falls back to the unscoped declaration.
	resolved, unresolved = ResolveValues(m, []string{"release_name"},
		ResolveContext{Outputs: outputs, LogicalName: "base"})
	if len(unresolved) != 0 || resolved["release_name"] != "istio-base" {
		t.Errorf("fallback: resolved = %v, unresolved = %v; want release_name=istio-base", resolved, unresolved)
	}
}

func TestResolveValues_FromAddressKeySegment(t *testing.T) {
	// A module that applies a MULTI-GVK manifest bundle through one for_each
	// resource keyed by each document's own composed identity
	// (apiVersion//kind//name[//namespace]): every composed-ID placeholder is
	// one "//"-delimited segment of the key itself. An index past the key's
	// segment count resolves to "" so the optional namespace segment drops
	// for cluster-scoped documents.
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "api_version",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromAddressKeySegment{FromAddressKeySegment: 0}},
					},
				},
				{
					Name: "kind",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromAddressKeySegment{FromAddressKeySegment: 1}},
					},
				},
				{
					Name: "name",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromAddressKeySegment{FromAddressKeySegment: 2}},
					},
				},
				{
					Name: "namespace",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromAddressKeySegment{FromAddressKeySegment: 3}},
					},
				},
			},
		},
	}

	names := []string{"api_version", "kind", "name", "namespace"}

	// A namespaced document: all four segments resolve — apiVersion keeps
	// its own single slash intact.
	resolved, unresolved := ResolveValues(m, names,
		ResolveContext{AddressKey: "apps/v1//Deployment//rabbitmq-cluster-operator//rabbitmq-system"})
	want := map[string]string{
		"api_version": "apps/v1",
		"kind":        "Deployment",
		"name":        "rabbitmq-cluster-operator",
		"namespace":   "rabbitmq-system",
	}
	if len(unresolved) != 0 {
		t.Errorf("namespaced: unresolved = %v; want none", unresolved)
	}
	for k, v := range want {
		if resolved[k] != v {
			t.Errorf("namespaced: resolved[%q] = %q; want %q", k, resolved[k], v)
		}
	}

	// A cluster-scoped document: the namespace segment is past the key's
	// count and stays unresolved (the ID's bracketed group drops it).
	resolved, unresolved = ResolveValues(m, names,
		ResolveContext{AddressKey: "apiextensions.k8s.io/v1//CustomResourceDefinition//rabbitmqclusters.rabbitmq.com"})
	if resolved["kind"] != "CustomResourceDefinition" || resolved["name"] != "rabbitmqclusters.rabbitmq.com" {
		t.Errorf("cluster-scoped: resolved = %v", resolved)
	}
	if len(unresolved) != 1 || unresolved[0] != "namespace" {
		t.Errorf("cluster-scoped: unresolved = %v; want [namespace]", unresolved)
	}

	// An empty address key (a non-repeated resource) resolves nothing.
	_, unresolved = ResolveValues(m, names, ResolveContext{})
	if len(unresolved) != len(names) {
		t.Errorf("empty key: unresolved = %v; want all %d names", unresolved, len(names))
	}
}

func TestResolveValues_FromClusterSecretKey_KeyFromAddressKey(t *testing.T) {
	// Keyed resource collections whose per-instance credentials live
	// under manifest-driven Secret keys (one random_password per
	// declared user, keyed by username): the Secret KEY is the
	// address's own instance key, so one declaration serves every
	// instance of the collection.
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "user_password_value",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromClusterSecretKey{
							FromClusterSecretKey: &kindv1.FromClusterSecretKey{
								NameSuffix:        "-auth",
								KeyFromAddressKey: true,
							},
						}},
					},
				},
			},
		},
	}
	names := []string{"user_password_value"}

	var gotSecret, gotKey string
	resolved, unresolved := ResolveValues(m, names, ResolveContext{
		MetadataName: "messaging",
		AddressKey:   "orders-service",
		ReadClusterSecret: func(secretName, key string) (string, error) {
			gotSecret, gotKey = secretName, key
			return "u53r-s3cr3t", nil
		},
	})
	if len(unresolved) != 0 {
		t.Errorf("keyed: unresolved = %v; want none", unresolved)
	}
	if resolved["user_password_value"] != "u53r-s3cr3t" {
		t.Errorf("keyed: resolved = %q; want the reader's value", resolved["user_password_value"])
	}
	if gotSecret != "messaging-auth" || gotKey != "orders-service" {
		t.Errorf("keyed: reader called with (%q, %q); want (messaging-auth, orders-service)", gotSecret, gotKey)
	}

	// Without an address key the flag cannot resolve — the value falls
	// back to unresolved (the ask-the-user path), never to a wrong key.
	_, unresolved = ResolveValues(m, names, ResolveContext{
		MetadataName: "messaging",
		ReadClusterSecret: func(secretName, key string) (string, error) {
			t.Errorf("reader must not be consulted without an address key (called with %q/%q)", secretName, key)
			return "", nil
		},
	})
	if len(unresolved) != 1 {
		t.Errorf("no address key: unresolved = %v; want the one name", unresolved)
	}
}

func TestResolveValues_FromClusterSecretKey(t *testing.T) {
	// An import ID that IS secret material (a random_password's value): the
	// arm reads the module-materialized Secret through the context's
	// cluster reader. Contexts without a reader (a disconnected wizard)
	// leave the value unresolved so the ask-the-user fallback carries it.
	m := &kindv1.CatalogKindImportMap{
		Spec: &kindv1.CatalogKindImportMapSpec{
			Values: []*kindv1.ImportValue{
				{
					Name: "admin_password_value",
					Derivations: []*kindv1.ImportValueDerivation{
						{Source: &kindv1.ImportValueDerivation_FromClusterSecretKey{
							FromClusterSecretKey: &kindv1.FromClusterSecretKey{
								NameSuffix: "-admin-auth",
								Key:        "password",
							},
						}},
					},
				},
			},
		},
	}
	names := []string{"admin_password_value"}

	// A cluster-connected context: the reader is consulted with the
	// convention-composed Secret name and the declared key.
	var gotSecret, gotKey string
	resolved, unresolved := ResolveValues(m, names, ResolveContext{
		MetadataName: "artifacts",
		ReadClusterSecret: func(secretName, key string) (string, error) {
			gotSecret, gotKey = secretName, key
			return "s3cr3t-value", nil
		},
	})
	if len(unresolved) != 0 {
		t.Errorf("connected: unresolved = %v; want none", unresolved)
	}
	if resolved["admin_password_value"] != "s3cr3t-value" {
		t.Errorf("connected: resolved = %q; want the reader's value", resolved["admin_password_value"])
	}
	if gotSecret != "artifacts-admin-auth" || gotKey != "password" {
		t.Errorf("connected: reader called with (%q, %q); want (artifacts-admin-auth, password)", gotSecret, gotKey)
	}

	// No reader bound (a disconnected context): unresolved, never an error.
	_, unresolved = ResolveValues(m, names, ResolveContext{MetadataName: "artifacts"})
	if len(unresolved) != 1 {
		t.Errorf("disconnected: unresolved = %v; want the one name", unresolved)
	}

	// A failing read resolves empty -- the Secret may legitimately not
	// exist for variants referencing user-provided credentials.
	_, unresolved = ResolveValues(m, names, ResolveContext{
		MetadataName: "artifacts",
		ReadClusterSecret: func(string, string) (string, error) {
			return "", errors.New("not found")
		},
	})
	if len(unresolved) != 1 {
		t.Errorf("read failure: unresolved = %v; want the one name", unresolved)
	}
}
