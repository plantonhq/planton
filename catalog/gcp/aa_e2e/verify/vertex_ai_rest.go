package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// vertexAiGet reads one Vertex AI resource by its full name on the regional
// host Google's provider uses (https://{region}-aiplatform.googleapis.com/v1/)
// and returns the decoded body as a generic map, so each verifier asserts
// the fields it cares about. Shared by the Vertex AI and Colab Enterprise
// kinds; the status code tells a 404 from any other error.
func vertexAiGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "vertex ai resource",
		fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", regionFromVertexResource(name), name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// vertexAiExistsWithLabel reads a Vertex AI resource and asserts the
// platform attribution label landed -- the cross-engine label-parity canary
// every labelled Vertex AI resource shares.
func vertexAiExistsWithLabel(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, error) {
	obj, _, err := vertexAiGet(ctx, svc, name)
	if err != nil {
		return nil, errors.Wrapf(err, "%s %s not found after deploy", what, name)
	}
	labels, _ := obj["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return nil, errors.Errorf("%s %s missing the planton-ai_resource attribution label after deploy", what, name)
	}
	return obj, nil
}

// vertexAiAbsent is the destroy verdict every Vertex AI verifier shares: a
// 404 means gone, any other error is unexpected, and a 200 means the
// resource survived.
func vertexAiAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := vertexAiGet(ctx, svc, name)
	return restAbsent(what, name, status, err)
}
