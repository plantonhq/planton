package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiDatasetVerifier probes a Vertex AI managed dataset through the
// REST API: the dataset reads back under its exported name with the
// attribution labels, and the dataset_id output is the numeric id at the
// end of that name -- the id both engines derive the same way.
type vertexAiDatasetVerifier struct{}

// IDOutputKey is the dataset's full resource name
// (projects/{p}/locations/{l}/datasets/{id}).
func (v *vertexAiDatasetVerifier) IDOutputKey() string { return "name" }

func (v *vertexAiDatasetVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	dataset, err := vertexAiExistsWithLabel(ctx, svc, "vertex ai dataset", name)
	if err != nil {
		return err
	}
	if got := outputs["dataset_id"]; got != lastPathSegment(stringField(dataset, "name")) {
		return errors.Errorf("vertex ai dataset %s dataset_id output %q does not match the live id %q", name, got, lastPathSegment(stringField(dataset, "name")))
	}
	if stringField(dataset, "displayName") == "" {
		return errors.Errorf("vertex ai dataset %s has no display name after deploy", name)
	}
	return nil
}

func (v *vertexAiDatasetVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "vertex ai dataset", outputs["name"])
}
