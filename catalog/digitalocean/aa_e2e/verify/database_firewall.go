package verify

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseFirewallVerifier verifies a DigitalOceanDatabaseFirewall via
// GET /v2/databases/{cluster_id}/firewall. The rule set is a PROPERTY of
// its cluster, not a standalone object:
//
//   - EXISTS means the cluster's live rule list is non-empty (the spec
//     requires at least one rule, so an empty list after deploy is a
//     failure, not a variant).
//   - ABSENT means the live rule list is EMPTY -- destroy PUTs an empty
//     list; there is no object to 404 on. A 404 (the cluster itself gone,
//     as happens when a composed teardown removes the fixture) also counts
//     as absent.
//
// The identity is the cluster UUID (the Terraform state id is a random
// unique string), which is exactly what the kind's outputs carry.
type databaseFirewallVerifier struct{}

func (*databaseFirewallVerifier) IDOutputKey() string { return "cluster_id" }

func (v *databaseFirewallVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	rules, err := v.liveRules(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandatabasefirewall verify-exists failed for cluster %q", id)
	}
	if len(rules) == 0 {
		return pkgerrors.Errorf("digitaloceandatabasefirewall on cluster %q has no rules after deploy", id)
	}
	return nil
}

func (v *databaseFirewallVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	rules, err := v.liveRules(ctx, client, id)
	if err != nil {
		if isNotFound(err) {
			// The whole cluster is gone -- its rule set with it.
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceandatabasefirewall verify-absent failed for cluster %q", id)
	}
	if len(rules) > 0 {
		return &StillExistsError{Component: "digitaloceandatabasefirewall", ID: id, Detail: fmt.Sprintf("still has %d rules after destroy (destroy must clear the set)", len(rules))}
	}
	return nil
}

func (v *databaseFirewallVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing after deploy")
	}
	return v.VerifyExists(ctx, client, id)
}

func (v *databaseFirewallVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func (*databaseFirewallVerifier) liveRules(ctx context.Context, client *godo.Client, clusterID string) ([]godo.DatabaseFirewallRule, error) {
	rules, _, err := client.Databases.GetFirewallRules(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	return rules, nil
}
