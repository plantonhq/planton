package verify

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// vectorSearchCollectionVerifier probes a Vector Search collection and its
// folded indexes through the Vector Search REST API on its global host
// (the pinned client library carries no Vector Search client). Existence
// of the collection and of every index the module exported, plus the
// platform attribution labels, are asserted from the JSON bodies.
type vectorSearchCollectionVerifier struct{}

// IDOutputKey is the collection's full resource name
// (projects/{p}/locations/{l}/collections/{id}).
func (v *vectorSearchCollectionVerifier) IDOutputKey() string { return "name" }

// vectorSearchNamed is the subset of the API's Collection and Index objects
// the verifier asserts on.
type vectorSearchNamed struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

func (v *vectorSearchCollectionVerifier) get(ctx context.Context, svc *Services, name string) (*vectorSearchNamed, int, error) {
	obj := &vectorSearchNamed{}
	status, err := googleRestGet(ctx, svc, "vector search resource", fmt.Sprintf("https://vectorsearch.googleapis.com/v1/%s", name), obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// indexNames collects the index_names list output as the outputs
// transformer flattens it -- one dot-indexed key per element
// (index_names.0, index_names.1, ...), in manifest order.
func indexNames(outputs map[string]string) []string {
	var names []string
	for i := 0; ; i++ {
		name, ok := outputs[fmt.Sprintf("index_names.%d", i)]
		if !ok {
			break
		}
		if strings.TrimSpace(name) != "" {
			names = append(names, name)
		}
	}
	return names
}

// VerifyExists confirms the collection reads back under the exported name
// with the attribution labels, and that every index the module exported
// exists and the count agrees.
func (v *vectorSearchCollectionVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	collection, _, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "vector search collection %s not found after deploy", name)
	}
	if collection.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("vector search collection %s missing the planton-ai_resource attribution label after deploy", name)
	}
	if got := outputs["collection_id"]; got != "" && lastPathSegment(collection.Name) != got {
		return errors.Errorf("vector search collection %s collection_id output %q does not match live name %q", name, got, lastPathSegment(collection.Name))
	}

	names := indexNames(outputs)
	if count := outputs["index_count"]; count != "" {
		if want, err := strconv.Atoi(count); err == nil && want != len(names) {
			return errors.Errorf("vector search collection %s exports index_count %d but %d index names", name, want, len(names))
		}
	}
	for _, indexName := range names {
		index, _, err := v.get(ctx, svc, indexName)
		if err != nil {
			return errors.Wrapf(err, "vector search index %s not found after deploy", indexName)
		}
		if index.Labels["planton-ai_resource"] != "true" {
			return errors.Errorf("vector search index %s missing the planton-ai_resource attribution label after deploy", indexName)
		}
	}
	return nil
}

// VerifyAbsent confirms the collection is gone (its indexes cannot outlive
// it).
func (v *vectorSearchCollectionVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, name)
	return restAbsent("vector search collection", name, status, err)
}
