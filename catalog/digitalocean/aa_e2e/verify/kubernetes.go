package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// kubernetesClusterVerifier verifies a DigitalOceanKubernetesCluster via
// GET /v2/kubernetes/clusters/{id}. Beyond existence, it asserts the live
// cluster is running and checks every identity and wiring value the module
// CLAIMS in its stack outputs -- the API endpoint, the URN, the public IPv4,
// the default pool's id, and the pod and service subnets -- against the
// live cluster. Outputs are contractually identical across both engines, so
// one assertion protects both, and an absent output simply means "not
// claimed" and is skipped. Status is always read live, never from an
// output: an apply-time snapshot goes stale immediately. The kubeconfig is
// deliberately not compared: it is a credential whose bytes are minted per
// fetch, so equality with anything is not a meaningful claim.
type kubernetesClusterVerifier struct{}

func (*kubernetesClusterVerifier) IDOutputKey() string { return "cluster_id" }

func (*kubernetesClusterVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := getKubernetesCluster(ctx, client, id)
	return err
}

func (*kubernetesClusterVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.Kubernetes.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceankubernetescluster verify-absent failed for %q", id)
	}
	return &StillExistsError{Component: "digitaloceankubernetescluster", ID: id}
}

func (v *kubernetesClusterVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing after deploy")
	}

	cluster, err := getKubernetesCluster(ctx, client, id)
	if err != nil {
		return err
	}

	if cluster.Status == nil || cluster.Status.State != godo.KubernetesClusterStatusRunning {
		state := "unknown"
		if cluster.Status != nil {
			state = string(cluster.Status.State)
		}
		return pkgerrors.Errorf("digitaloceankubernetescluster %q status is %q, want running", id, state)
	}

	// The default pool is the one the provider marks with its own tag; a
	// cluster this module created always has exactly one such pool. Its id
	// is what the default_node_pool_id output claims.
	liveDefaultPoolID := ""
	for _, pool := range cluster.NodePools {
		for _, t := range pool.Tags {
			if t == "terraform:default-node-pool" {
				liveDefaultPoolID = pool.ID
			}
		}
	}

	claims := []struct {
		output string
		live   string
	}{
		{"api_server_endpoint", cluster.Endpoint},
		{"urn", cluster.URN()},
		{"ipv4_address", cluster.IPv4},
		{"default_node_pool_id", liveDefaultPoolID},
		{"cluster_subnet", cluster.ClusterSubnet},
		{"service_subnet", cluster.ServiceSubnet},
	}
	for _, c := range claims {
		if claimed := StringOutput(outputs, c.output); claimed != "" && claimed != c.live {
			return pkgerrors.Errorf("digitaloceankubernetescluster %q %s mismatch: output %q, live %q",
				id, c.output, claimed, c.live)
		}
	}

	return nil
}

func (v *kubernetesClusterVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func getKubernetesCluster(ctx context.Context, client *godo.Client, id string) (*godo.KubernetesCluster, error) {
	cluster, _, err := client.Kubernetes.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("digitaloceankubernetescluster %q not found after deploy", id)
		}
		return nil, pkgerrors.Wrapf(err, "digitaloceankubernetescluster verify-exists failed for %q", id)
	}
	return cluster, nil
}

// kubernetesNodePoolVerifier verifies a DigitalOceanKubernetesNodePool. The
// API addresses node pools as /v2/kubernetes/clusters/{cluster_id}/node_pools/{id},
// and the kind's stack outputs carry BOTH ids, so the outputs form addresses
// the pool directly -- one GET validates both claimed outputs against live
// state at once. The plain-id forms keep the account-wide cluster scan (the
// same discovery the upstream provider's own node-pool importer performs)
// for callers that only hold the pool UUID.
type kubernetesNodePoolVerifier struct{}

func (*kubernetesNodePoolVerifier) IDOutputKey() string { return "node_pool_id" }

func (*kubernetesNodePoolVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := kubernetesNodePoolExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceankubernetesnodepool verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceankubernetesnodepool %q not found after deploy", id)
	}
	return nil
}

func (*kubernetesNodePoolVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := kubernetesNodePoolExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceankubernetesnodepool verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceankubernetesnodepool", ID: id}
	}
	return nil
}

func (v *kubernetesNodePoolVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	poolID := StringOutput(outputs, "node_pool_id")
	if poolID == "" {
		return pkgerrors.New("node_pool_id output missing after deploy")
	}
	clusterID := StringOutput(outputs, "cluster_id")
	if clusterID == "" {
		return pkgerrors.New("cluster_id output missing after deploy")
	}

	// Direct addressing: finding the pool under the claimed cluster proves
	// both outputs against live state in one call. Node-level assertions
	// are deliberately absent -- with autoscaling the live node set drifts
	// from the apply-time snapshot by design.
	_, _, err := client.Kubernetes.GetNodePool(ctx, clusterID, poolID)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceankubernetesnodepool %q not found under cluster %q after deploy", poolID, clusterID)
		}
		return pkgerrors.Wrapf(err, "digitaloceankubernetesnodepool verify-exists failed for %q", poolID)
	}

	return nil
}

func (v *kubernetesNodePoolVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	poolID := StringOutput(outputs, "node_pool_id")
	if poolID == "" {
		return pkgerrors.New("node_pool_id output missing for destroy verification")
	}
	clusterID := StringOutput(outputs, "cluster_id")
	if clusterID == "" {
		return pkgerrors.New("cluster_id output missing for destroy verification")
	}

	// A destroyed pool 404s directly; a destroyed owning cluster (torn down
	// as the scenario's fixture) also proves the pool is gone.
	_, _, err := client.Kubernetes.GetNodePool(ctx, clusterID, poolID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceankubernetesnodepool verify-absent failed for %q", poolID)
	}
	return &StillExistsError{Component: "digitaloceankubernetesnodepool", ID: poolID}
}

func kubernetesNodePoolExists(ctx context.Context, client *godo.Client, poolID string) (bool, error) {
	opts := &godo.ListOptions{PerPage: 200}
	for {
		clusters, resp, err := client.Kubernetes.List(ctx, opts)
		if err != nil {
			return false, pkgerrors.Wrap(err, "listing kubernetes clusters to locate the node pool")
		}
		for _, cluster := range clusters {
			for _, pool := range cluster.NodePools {
				if pool.ID == poolID {
					return true, nil
				}
			}
		}
		if resp.Links == nil || resp.Links.IsLastPage() {
			return false, nil
		}
		page, err := resp.Links.CurrentPage()
		if err != nil {
			return false, pkgerrors.Wrap(err, "reading kubernetes cluster list pagination")
		}
		opts.Page = page + 1
	}
}
