package verify

import (
	"context"
	"strconv"
	"sync"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// dropletAutoscalePoolVerifier verifies a DigitalOceanDropletAutoscalePool
// via GET /v2/droplets/autoscale/{id} and its members endpoint.
//
// The kind's headline promise is a fleet that exists: the deploy-side check
// asserts the pool is live `active` AND that it has at least one member, every
// member itself `active` (the API reports the pool active before its first
// member finishes provisioning -- measured: pool active at +1s, member active
// at +33s -- so the pool status alone proves nothing about the fleet).
//
// The absence check matters doubly here: the pool's only delete is the
// "dangerous" variant that terminates every member, so a pool that lingers
// keeps real droplets billing -- and a pool that is gone while a member
// lingers is the same leak wearing a 404. Member droplets carry nothing that
// links them back to their pool (their name is `<pool-name>-<uuid>-NNN` and
// their tags are the template's), and member ids are deliberately not
// outputs (they churn by design), so the verifier remembers the member ids
// it saw at deploy and probes each one at destroy. This is the harness's one
// stateful verifier; if the Kubernetes cluster or node-pool lanes need the
// same destroy-side member check, lift the memo into a harness-level
// observation hook rather than adding a second memo.
//
// Measured live: after the dangerous delete both the pool and its member
// answer a real HTTP 404 within ten seconds (the pool's body reads
// "autoscale group with id ... not found" -- the string the upstream provider
// matches on; the typed status check below covers the same response).
type dropletAutoscalePoolVerifier struct {
	mu      sync.Mutex
	members map[string][]int // pool id -> member droplet ids seen at deploy
}

func (*dropletAutoscalePoolVerifier) IDOutputKey() string { return "pool_id" }

func (v *dropletAutoscalePoolVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandropletautoscalepool requires the full outputs map; " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (v *dropletAutoscalePoolVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandropletautoscalepool requires the full outputs map; " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *dropletAutoscalePoolVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "pool_id")
	if id == "" {
		return pkgerrors.New("pool_id output missing after deploy")
	}
	pool, _, err := client.DropletAutoscale.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceandropletautoscalepool %q not found after deploy", id)
		}
		return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool verify-exists failed for %q", id)
	}
	if pool.Status != "active" {
		return pkgerrors.Errorf("digitaloceandropletautoscalepool %q live status is %q, want %q", id, pool.Status, "active")
	}

	members, err := listPoolMembers(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool %q: listing members failed", id)
	}
	if len(members) == 0 {
		return pkgerrors.Errorf("digitaloceandropletautoscalepool %q is active but has no members", id)
	}
	ids := make([]int, 0, len(members))
	for _, m := range members {
		if m.Status != "active" {
			return pkgerrors.Errorf("digitaloceandropletautoscalepool %q member droplet %d status is %q, want %q", id, m.DropletID, m.Status, "active")
		}
		ids = append(ids, int(m.DropletID))
	}

	v.mu.Lock()
	if v.members == nil {
		v.members = map[string][]int{}
	}
	v.members[id] = ids
	v.mu.Unlock()
	return nil
}

func (v *dropletAutoscalePoolVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "pool_id")
	if id == "" {
		return pkgerrors.New("pool_id output missing for destroy verification")
	}
	_, _, err := client.DropletAutoscale.Get(ctx, id)
	if err == nil {
		return &StillExistsError{Component: "digitaloceandropletautoscalepool", ID: id}
	}
	if !isNotFound(err) {
		return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool verify-absent failed for %q", id)
	}

	// The pool is gone; every member it had at deploy must be gone too.
	v.mu.Lock()
	members := v.members[id]
	v.mu.Unlock()
	for _, dropletID := range members {
		_, _, err := client.Droplets.Get(ctx, dropletID)
		if err == nil {
			return &StillExistsError{
				Component: "digitaloceandropletautoscalepool",
				ID:        id,
				Detail:    "is gone but its member droplet " + strconv.Itoa(dropletID) + " still exists after destroy",
			}
		}
		if !isNotFound(err) {
			return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool %q: probing member droplet %d failed", id, dropletID)
		}
	}
	return nil
}

// listPoolMembers pages through the pool's members exactly as the upstream
// provider's create waiter does.
func listPoolMembers(ctx context.Context, client *godo.Client, poolID string) ([]*godo.DropletAutoscaleResource, error) {
	var members []*godo.DropletAutoscaleResource
	opts := &godo.ListOptions{Page: 1, PerPage: 100}
	for {
		page, resp, err := client.DropletAutoscale.ListMembers(ctx, poolID, opts)
		if err != nil {
			return nil, err
		}
		members = append(members, page...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			return members, nil
		}
		current, err := resp.Links.CurrentPage()
		if err != nil {
			return members, nil
		}
		opts.Page = current + 1
	}
}
