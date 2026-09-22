package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// cloudRunWorkerPoolVerifier probes a Cloud Run worker pool through the Run
// Admin API v2 REST surface. The pinned google.golang.org/api line predates
// the typed worker-pool client, so the probe is a plain authenticated GET on
// the pool's documented resource path -- existence, a Ready terminal
// condition, and the revision pointers are asserted from the JSON body.
type cloudRunWorkerPoolVerifier struct{}

// IDOutputKey is the pool's full resource name
// (projects/{p}/locations/{l}/workerPools/{name}).
func (v *cloudRunWorkerPoolVerifier) IDOutputKey() string { return "name" }

// cloudRunWorkerPool is the subset of the API's WorkerPool object the
// verifier asserts on.
type cloudRunWorkerPool struct {
	Name                  string `json:"name"`
	Uid                   string `json:"uid"`
	LatestReadyRevision   string `json:"latestReadyRevision"`
	LatestCreatedRevision string `json:"latestCreatedRevision"`
	Reconciling           bool   `json:"reconciling"`
	TerminalCondition     struct {
		Type    string `json:"type"`
		State   string `json:"state"`
		Message string `json:"message"`
	} `json:"terminalCondition"`
}

func (v *cloudRunWorkerPoolVerifier) get(ctx context.Context, svc *Services, name string) (*cloudRunWorkerPool, int, error) {
	url := fmt.Sprintf("https://run.googleapis.com/v2/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to build run worker pool GET request")
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "run worker pool GET request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to read run worker pool response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.Errorf("run worker pool GET %s returned %d: %s", name, resp.StatusCode, string(body))
	}

	pool := &cloudRunWorkerPool{}
	if err := json.Unmarshal(body, pool); err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to decode run worker pool")
	}
	return pool, resp.StatusCode, nil
}

func (v *cloudRunWorkerPoolVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	pool, _, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "run worker pool %s not found after deploy", name)
	}
	if pool.TerminalCondition.State != "CONDITION_SUCCEEDED" {
		return errors.Errorf("run worker pool %s terminal condition is %q (%s), want CONDITION_SUCCEEDED", name, pool.TerminalCondition.State, pool.TerminalCondition.Message)
	}
	if uid := outputs["uid"]; uid != "" && pool.Uid != uid {
		return errors.Errorf("run worker pool %s uid mismatch: output %q, live %q", name, uid, pool.Uid)
	}
	if rev := outputs["latest_ready_revision"]; rev != "" && pool.LatestReadyRevision != rev {
		return errors.Errorf("run worker pool %s latest_ready_revision mismatch: output %q, live %q", name, rev, pool.LatestReadyRevision)
	}
	if pool.LatestReadyRevision == "" {
		return errors.Errorf("run worker pool %s has no ready revision", name)
	}
	return nil
}

func (v *cloudRunWorkerPoolVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, name)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing run worker pool %s after destroy", name)
	}
	return errors.Errorf("run worker pool %s still exists after destroy", name)
}
