package verify

import (
	"context"

	"github.com/pkg/errors"
)

// Colab Enterprise resources live on the Vertex AI API, so the three Colab
// verifiers read them through vertexAiGet.

// colabRuntimeTemplateVerifier probes a Colab Enterprise runtime template:
// it reads back with the attribution labels, and runtime_template_id is
// the last segment of its name.
type colabRuntimeTemplateVerifier struct{}

// IDOutputKey is the template's full resource name
// (projects/{p}/locations/{l}/notebookRuntimeTemplates/{id}).
func (v *colabRuntimeTemplateVerifier) IDOutputKey() string { return "name" }

func (v *colabRuntimeTemplateVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	if _, err := vertexAiExistsWithLabel(ctx, svc, "colab runtime template", name); err != nil {
		return err
	}
	if got := outputs["runtime_template_id"]; got != lastPathSegment(name) {
		return errors.Errorf("colab runtime template %s runtime_template_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *colabRuntimeTemplateVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "colab runtime template", outputs["name"])
}

// colabRuntimeVerifier probes a Colab Enterprise runtime: it reads back
// assigned to a user, built from a template, in a RUNNING or STOPPED state.
type colabRuntimeVerifier struct{}

// IDOutputKey is the runtime's full resource name
// (projects/{p}/locations/{l}/notebookRuntimes/{id}).
func (v *colabRuntimeVerifier) IDOutputKey() string { return "name" }

func (v *colabRuntimeVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	runtime, _, err := vertexAiGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "colab runtime %s not found after deploy", name)
	}
	if stringField(runtime, "runtimeUser") == "" {
		return errors.Errorf("colab runtime %s has no runtime user after deploy", name)
	}
	if state := stringField(runtime, "runtimeState"); state != "RUNNING" && state != "STOPPED" {
		return errors.Errorf("colab runtime %s is %q after deploy, expected RUNNING or STOPPED", name, state)
	}
	if got := outputs["runtime_id"]; got != lastPathSegment(name) {
		return errors.Errorf("colab runtime %s runtime_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *colabRuntimeVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "colab runtime", outputs["name"])
}

// colabScheduleVerifier probes a Vertex AI schedule: it reads back with its
// cron in an ACTIVE or PAUSED state, and schedule_id is the last segment
// of its name.
type colabScheduleVerifier struct{}

// IDOutputKey is the schedule's full resource name
// (projects/{p}/locations/{l}/schedules/{id}).
func (v *colabScheduleVerifier) IDOutputKey() string { return "name" }

func (v *colabScheduleVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	schedule, _, err := vertexAiGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "colab schedule %s not found after deploy", name)
	}
	if stringField(schedule, "cron") == "" {
		return errors.Errorf("colab schedule %s has no cron after deploy", name)
	}
	if state := stringField(schedule, "state"); state != "ACTIVE" && state != "PAUSED" {
		return errors.Errorf("colab schedule %s is %q after deploy, expected ACTIVE or PAUSED", name, state)
	}
	if got := outputs["schedule_id"]; got != lastPathSegment(name) {
		return errors.Errorf("colab schedule %s schedule_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *colabScheduleVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "colab schedule", outputs["name"])
}
