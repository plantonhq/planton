package setdeploy

import (
	"strings"
	"testing"
)

func TestParseNodeOverride(t *testing.T) {
	parsed, err := ParseNodeOverride("TestCatalogKindGeneric/consumer:spec.displayName", "v2")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Kind != "TestCatalogKindGeneric" || parsed.Name != "consumer" ||
		parsed.FieldPath != "spec.displayName" || parsed.Value != "v2" {
		t.Fatalf("wrong parse: %+v", parsed)
	}

	for _, bad := range []string{"spec.displayName", "kind/name", "/name:path", "kind/:path", "kind/name:"} {
		if _, err := ParseNodeOverride(bad, "v"); err == nil {
			t.Fatalf("%q must refuse as not node-addressed", bad)
		} else if !strings.Contains(err.Error(), "node-addressed") {
			t.Fatalf("the refusal must teach the form, got: %v", err)
		}
	}
}

func TestApplyNodeOverride_SetsTheNamedDocumentOnly(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml, "02-consumer.yaml": consumerYaml})
	out, err := ApplyNodeOverride(docs, NodeOverride{
		Kind: "TestCatalogKindGeneric", Name: "consumer",
		FieldPath: "spec.displayName", Value: "storefront-v2",
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !strings.Contains(string(out[1].Bytes), "storefront-v2") {
		t.Fatalf("the named document must carry the value; got:\n%s", out[1].Bytes)
	}
	if strings.Contains(string(out[0].Bytes), "storefront-v2") {
		t.Fatalf("the sibling document must be untouched")
	}
	// The input slice is never mutated — callers own their docs.
	if strings.Contains(string(docs[1].Bytes), "storefront-v2") {
		t.Fatalf("the input docs must not be mutated in place")
	}
}

func TestApplyNodeOverride_MissRefusesNamingTheSet(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml})
	_, err := ApplyNodeOverride(docs, NodeOverride{
		Kind: "TestCatalogKindGeneric", Name: "ghost",
		FieldPath: "spec.displayName", Value: "v",
	})
	if err == nil {
		t.Fatalf("a miss must refuse")
	}
	if !strings.Contains(err.Error(), "TestCatalogKindGeneric/producer") {
		t.Fatalf("the refusal must name what the set holds, got: %v", err)
	}
}
