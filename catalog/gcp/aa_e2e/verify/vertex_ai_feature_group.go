package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiFeatureGroupVerifier probes a Vertex AI Feature Store feature
// group and its registered features through the REST API: the group reads
// back with the attribution labels and its BigQuery source, and every
// exported feature exists with the labels the modules merged onto it.
type vertexAiFeatureGroupVerifier struct{}

// IDOutputKey is the group's full resource name
// (projects/{p}/locations/{l}/featureGroups/{id}).
func (v *vertexAiFeatureGroupVerifier) IDOutputKey() string { return "name" }

func (v *vertexAiFeatureGroupVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	group, err := vertexAiExistsWithLabel(ctx, svc, "vertex ai feature group", name)
	if err != nil {
		return err
	}
	if got := outputs["feature_group_id"]; got != lastPathSegment(stringField(group, "name")) {
		return errors.Errorf("vertex ai feature group %s feature_group_id output %q does not match the live id %q", name, got, lastPathSegment(stringField(group, "name")))
	}
	for _, feature := range listOutput(outputs, "feature_names") {
		if _, err := vertexAiExistsWithLabel(ctx, svc, "feature group feature", feature); err != nil {
			return err
		}
	}
	return nil
}

// VerifyAbsent confirms the group is gone (its features cannot outlive it).
func (v *vertexAiFeatureGroupVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "vertex ai feature group", outputs["name"])
}
