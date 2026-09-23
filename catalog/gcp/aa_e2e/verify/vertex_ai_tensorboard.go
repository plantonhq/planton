package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiTensorboardVerifier probes a Vertex AI TensorBoard and the
// experiments and runs declared in it through the REST API: the
// TensorBoard reads back with the attribution labels, tensorboard_id is
// the numeric id at the end of its name, and every exported experiment and
// run exists with the labels the modules merged onto it.
type vertexAiTensorboardVerifier struct{}

// IDOutputKey is the TensorBoard's full resource name
// (projects/{p}/locations/{l}/tensorboards/{id}).
func (v *vertexAiTensorboardVerifier) IDOutputKey() string { return "name" }

func (v *vertexAiTensorboardVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	tensorboard, err := vertexAiExistsWithLabel(ctx, svc, "vertex ai tensorboard", name)
	if err != nil {
		return err
	}
	if got := outputs["tensorboard_id"]; got != lastPathSegment(stringField(tensorboard, "name")) {
		return errors.Errorf("vertex ai tensorboard %s tensorboard_id output %q does not match the live id %q", name, got, lastPathSegment(stringField(tensorboard, "name")))
	}
	for _, experiment := range listOutput(outputs, "experiment_names") {
		if _, err := vertexAiExistsWithLabel(ctx, svc, "tensorboard experiment", experiment); err != nil {
			return err
		}
	}
	for _, run := range listOutput(outputs, "run_names") {
		if _, err := vertexAiExistsWithLabel(ctx, svc, "tensorboard run", run); err != nil {
			return err
		}
	}
	return nil
}

// VerifyAbsent confirms the TensorBoard is gone (its experiments and runs
// cannot outlive it).
func (v *vertexAiTensorboardVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "vertex ai tensorboard", outputs["name"])
}
