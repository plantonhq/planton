package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// documentAiGet reads a Document AI resource on the location-prefixed host
// Google's provider uses (https://{location}-documentai.googleapis.com/v1/).
func documentAiGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "document ai processor",
		fmt.Sprintf("https://%s-documentai.googleapis.com/v1/%s", regionFromVertexResource(name), name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// documentAiProcessorVerifier probes a Document AI processor: it reads back
// ENABLED with a type, processor_id is the last segment of its name, and
// the process endpoint output points at that name on the processor's
// regional host.
type documentAiProcessorVerifier struct{}

// IDOutputKey is the processor's full resource name
// (projects/{p}/locations/{l}/processors/{id}).
func (v *documentAiProcessorVerifier) IDOutputKey() string { return "name" }

func (v *documentAiProcessorVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	processor, _, err := documentAiGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "document ai processor %s not found after deploy", name)
	}
	if state := stringField(processor, "state"); state != "ENABLED" {
		return errors.Errorf("document ai processor %s is %q after deploy, expected ENABLED", name, state)
	}
	if stringField(processor, "type") == "" {
		return errors.Errorf("document ai processor %s has no type after deploy", name)
	}
	if got := outputs["processor_id"]; got != lastPathSegment(stringField(processor, "name")) {
		return errors.Errorf("document ai processor %s processor_id output %q does not match the live id", name, got)
	}
	if endpoint := outputs["process_endpoint"]; !strings.HasSuffix(endpoint, "/v1/"+name+":process") {
		return errors.Errorf("document ai processor %s process_endpoint output %q does not target it", name, endpoint)
	}
	return nil
}

func (v *documentAiProcessorVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := documentAiGet(ctx, svc, name)
	return restAbsent("document ai processor", name, status, err)
}
