package verify

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// vertexAiRagEngineConfigVerifier probes the RAG Engine configuration of one
// location through the Vertex AI REST API (the pinned client library has no
// typed surface for it). The configuration is a per-location singleton
// Google owns: it never returns 404, so "absent" after destroy means the
// tier has been PATCHed to UNPROVISIONED -- the provider's delete semantics
// -- rather than the resource being gone.
type vertexAiRagEngineConfigVerifier struct{}

// IDOutputKey is the singleton's full resource name
// (projects/{p}/locations/{l}/ragEngineConfig).
func (v *vertexAiRagEngineConfigVerifier) IDOutputKey() string { return "name" }

// vertexAiRagEngineConfig is the subset of the API's RagEngineConfig the
// verifier asserts on: which of the three tier blocks is present.
type vertexAiRagEngineConfig struct {
	Name               string `json:"name"`
	RagManagedDbConfig struct {
		Basic         *struct{} `json:"basic"`
		Scaled        *struct{} `json:"scaled"`
		Unprovisioned *struct{} `json:"unprovisioned"`
	} `json:"ragManagedDbConfig"`
}

func (c *vertexAiRagEngineConfig) tier() string {
	switch {
	case c.RagManagedDbConfig.Basic != nil:
		return "BASIC"
	case c.RagManagedDbConfig.Scaled != nil:
		return "SCALED"
	case c.RagManagedDbConfig.Unprovisioned != nil:
		return "UNPROVISIONED"
	}
	return ""
}

func (v *vertexAiRagEngineConfigVerifier) get(ctx context.Context, svc *Services, name string) (*vertexAiRagEngineConfig, int, error) {
	cfg := &vertexAiRagEngineConfig{}
	status, err := googleRestGet(ctx, svc, "rag engine config", fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", regionFromVertexResource(name), name), cfg)
	if err != nil {
		return nil, status, err
	}
	return cfg, status, nil
}

// VerifyExists confirms the configuration reads back at a provisioned tier
// (BASIC or SCALED) under the name both engines export.
func (v *vertexAiRagEngineConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	cfg, _, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "rag engine config %s not readable after deploy", name)
	}
	if tier := cfg.tier(); tier == "" || tier == "UNPROVISIONED" {
		return errors.Errorf("rag engine config %s reads back as %q after deploy; expected a provisioned tier", name, tier)
	}
	if location := outputs["location"]; location != "" && regionFromVertexResource(name) != location {
		return errors.Errorf("rag engine config %s location output %q does not match the resource path", name, location)
	}
	return nil
}

// VerifyAbsent confirms the destroy PATCHed the location back to
// UNPROVISIONED. The singleton itself never disappears.
func (v *vertexAiRagEngineConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	cfg, status, err := v.get(ctx, svc, name)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing rag engine config %s after destroy", name)
	}
	if tier := cfg.tier(); tier != "UNPROVISIONED" {
		return errors.Errorf("rag engine config %s still reads %q after destroy; expected UNPROVISIONED", name, tier)
	}
	return nil
}
