package verify

import (
	"context"

	"github.com/pkg/errors"
)

// tpuGet reads a Cloud TPU resource on the TPU API's GA surface
// (https://tpu.googleapis.com/v2/), which serves the same nodes and queued
// resources Google's beta provider manages.
func tpuGet(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, what, "https://tpu.googleapis.com/v2/"+name, &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// tpuVmVerifier probes a Cloud TPU VM: it reads back READY with the
// attribution labels and at least one network endpoint, and node_id is the
// last segment of its name.
type tpuVmVerifier struct{}

// IDOutputKey is the TPU's full resource name
// (projects/{p}/locations/{zone}/nodes/{id}).
func (v *tpuVmVerifier) IDOutputKey() string { return "name" }

func (v *tpuVmVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	node, _, err := tpuGet(ctx, svc, "tpu vm", name)
	if err != nil {
		return errors.Wrapf(err, "tpu vm %s not found after deploy", name)
	}
	if state := stringField(node, "state"); state != "READY" {
		return errors.Errorf("tpu vm %s is %q after deploy, expected READY", name, state)
	}
	labels, _ := node["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return errors.Errorf("tpu vm %s missing the planton-ai_resource attribution label after deploy", name)
	}
	if endpoints, _ := node["networkEndpoints"].([]interface{}); len(endpoints) == 0 {
		return errors.Errorf("tpu vm %s has no network endpoints after deploy", name)
	}
	if got := outputs["node_id"]; got != lastPathSegment(name) {
		return errors.Errorf("tpu vm %s node_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *tpuVmVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := tpuGet(ctx, svc, "tpu vm", name)
	return restAbsent("tpu vm", name, status, err)
}

// tpuQueuedResourceVerifier probes a Cloud TPU queued resource: it reads
// back with a queue state and a node spec, and queued_resource_id is the
// last segment of its name. Capacity may keep it WAITING_FOR_RESOURCES --
// accepted as a live state; the request existing is what the block
// declares.
type tpuQueuedResourceVerifier struct{}

// IDOutputKey is the request's full resource name
// (projects/{p}/locations/{zone}/queuedResources/{id}).
func (v *tpuQueuedResourceVerifier) IDOutputKey() string { return "name" }

func (v *tpuQueuedResourceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	request, _, err := tpuGet(ctx, svc, "tpu queued resource", name)
	if err != nil {
		return errors.Wrapf(err, "tpu queued resource %s not found after deploy", name)
	}
	state, _ := request["state"].(map[string]interface{})
	switch s, _ := state["state"].(string); s {
	case "ACCEPTED", "PROVISIONING", "WAITING_FOR_RESOURCES", "ACTIVE", "CREATING":
	default:
		return errors.Errorf("tpu queued resource %s is %q after deploy", name, s)
	}
	tpuSpec, _ := request["tpu"].(map[string]interface{})
	if nodeSpecs, _ := tpuSpec["nodeSpec"].([]interface{}); len(nodeSpecs) == 0 {
		return errors.Errorf("tpu queued resource %s carries no node spec after deploy", name)
	}
	if got := outputs["queued_resource_id"]; got != lastPathSegment(name) {
		return errors.Errorf("tpu queued resource %s queued_resource_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *tpuQueuedResourceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := tpuGet(ctx, svc, "tpu queued resource", name)
	return restAbsent("tpu queued resource", name, status, err)
}
