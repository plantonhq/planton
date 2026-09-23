package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiFeatureOnlineStoreVerifier probes a Vertex AI Feature Store
// online store and its feature views through the REST API: the store reads
// back with the attribution labels and a STABLE state, and every exported
// feature view exists with the labels the modules merged onto it.
type vertexAiFeatureOnlineStoreVerifier struct{}

// IDOutputKey is the store's full resource name
// (projects/{p}/locations/{l}/featureOnlineStores/{id}).
func (v *vertexAiFeatureOnlineStoreVerifier) IDOutputKey() string { return "name" }

func (v *vertexAiFeatureOnlineStoreVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	store, err := vertexAiExistsWithLabel(ctx, svc, "vertex ai feature online store", name)
	if err != nil {
		return err
	}
	if state := stringField(store, "state"); state != "STABLE" {
		return errors.Errorf("vertex ai feature online store %s is in state %q after deploy, want STABLE", name, state)
	}
	for _, view := range listOutput(outputs, "feature_view_names") {
		if _, err := vertexAiExistsWithLabel(ctx, svc, "feature view", view); err != nil {
			return err
		}
	}
	return nil
}

// VerifyAbsent confirms the store is gone (its views cannot outlive it).
func (v *vertexAiFeatureOnlineStoreVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "vertex ai feature online store", outputs["name"])
}
