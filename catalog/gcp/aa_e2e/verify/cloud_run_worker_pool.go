package verify

import (
	"context"
	"fmt"

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
	pool := &cloudRunWorkerPool{}
	status, err := googleRestGet(ctx, svc, "run worker pool", fmt.Sprintf("https://run.googleapis.com/v2/%s", name), pool)
	if err != nil {
		return nil, status, err
	}
	return pool, status, nil
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
	return restAbsent("run worker pool", name, status, err)
}
