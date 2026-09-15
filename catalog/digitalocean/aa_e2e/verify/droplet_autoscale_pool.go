package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// dropletAutoscalePoolVerifier verifies a DigitalOceanDropletAutoscalePool
// via GET /v2/droplets/autoscale_pools/{id}. The absence check matters
// doubly here: the pool's delete destroys its member droplets, so a pool
// that lingers after destroy keeps real droplets billing. The provider's
// own code matches this endpoint's not-found by error STRING ("autoscale
// group with id ... not found"); the typed 404 check below covers the same
// response -- if the proof lane ever sees an absence reported as a non-404,
// that finding belongs in the provider defect register.
type dropletAutoscalePoolVerifier struct{}

func (*dropletAutoscalePoolVerifier) IDOutputKey() string { return "pool_id" }

func (*dropletAutoscalePoolVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := dropletAutoscalePoolExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandropletautoscalepool %q not found after deploy", id)
	}
	return nil
}

func (*dropletAutoscalePoolVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := dropletAutoscalePoolExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandropletautoscalepool verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandropletautoscalepool", ID: id}
	}
	return nil
}

func dropletAutoscalePoolExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.DropletAutoscale.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
