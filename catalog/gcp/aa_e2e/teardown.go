package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/yaml"
)

var (
	_ provider.DependencyTeardownPreparer = (*Harness)(nil)
	_ provider.DestroyPreparer            = (*Harness)(nil)
)

// The classes of owner step a GCP scenario may name with
// planton.dev/e2e-prepare-destroy (provider.DestroyPreparer).
const (
	// prepareConsensusPeering: Google deletes a VPC peering whose update
	// strategy is CONSENSUS only after the owners of both networks have
	// called networks.requestRemovePeering, and neither provider calls it,
	// so both engines' destroys are refused ("cannot be deleted unless the
	// connection has delete_status=DELETE_ACKNOWLEDGED"). The kind's GUIDE
	// tells the customer to request removal on both networks before the
	// destroy; the harness makes those two requests in the owners' place.
	prepareConsensusPeering = "consensus-peering"
)

// consensusRemovalBudget bounds the wait for Google to acknowledge both
// removal requests.
const consensusRemovalBudget = 5 * time.Minute

// networkPath reads project and name from a network's self link or path
// (https://www.googleapis.com/compute/v1/projects/P/global/networks/N).
var networkPath = regexp.MustCompile(`projects/([^/]+)/global/networks/([^/]+)$`)

// subnetworkPath reads project, region and name from a subnetwork's self
// link or path (https://www.googleapis.com/compute/v1/projects/P/regions/R/subnetworks/N).
var subnetworkPath = regexp.MustCompile(`projects/([^/]+)/regions/([^/]+)/subnetworks/([^/]+)$`)

// PrepareDependencyDestroy implements provider.DependencyTeardownPreparer.
// Deleting a Managed Kafka Connect cluster leaves the network attachment
// Google made for its workers (connect-na-*) on the cluster's subnetwork, with
// no connection endpoints, and the subnetwork refuses deletion while it
// exists -- measured holding it for more than 25 minutes. Before a subnetwork
// fixture's destroy, the harness deletes every network attachment on that
// subnetwork that serves no endpoint; one still serving is left alone, so the
// destroy reports why it cannot proceed.
func (h *Harness) PrepareDependencyDestroy(ctx context.Context, kindSlug string, outputs map[string]interface{}) error {
	if kindSlug != "gcpsubnetwork" {
		return nil
	}
	link, _ := outputs["subnetwork_self_link"].(string)
	match := subnetworkPath.FindStringSubmatch(link)
	if match == nil {
		return nil
	}
	project, region, name := match[1], match[2], match[3]
	base := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/regions/%s/networkAttachments", project, region)
	for pageToken := ""; ; {
		var page struct {
			Items []struct {
				Name                string        `json:"name"`
				SelfLink            string        `json:"selfLink"`
				Subnetworks         []string      `json:"subnetworks"`
				ConnectionEndpoints []interface{} `json:"connectionEndpoints"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}
		if _, err := h.restGet(ctx, base+"?maxResults=500&pageToken="+pageToken, &page); err != nil {
			return errors.Wrapf(err, "listing the network attachments in %s/%s", project, region)
		}
		for _, attachment := range page.Items {
			if len(attachment.ConnectionEndpoints) > 0 || !onSubnetwork(attachment.Subnetworks, region, name) {
				continue
			}
			if err := h.restDelete(ctx, attachment.SelfLink); err != nil {
				return errors.Wrapf(err, "deleting the network attachment %s that still holds the subnetwork %s", attachment.Name, name)
			}
		}
		if pageToken = page.NextPageToken; pageToken == "" {
			return nil
		}
	}
}

func onSubnetwork(subnetworks []string, region, name string) bool {
	suffix := "/regions/" + region + "/subnetworks/" + name
	for _, subnetwork := range subnetworks {
		if strings.HasSuffix(subnetwork, suffix) {
			return true
		}
	}
	return false
}

// PrepareDestroy implements provider.DestroyPreparer.
func (h *Harness) PrepareDestroy(ctx context.Context, tc *provider.KindTestContext, class string) error {
	manifestPath, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	if manifestPath == "" {
		return errors.New("no manifest on the context; the runner passes the resolved first act's")
	}
	switch class {
	case prepareConsensusPeering:
		return h.requestConsensusPeeringRemoval(ctx, tc, manifestPath)
	default:
		return errors.Errorf("unknown prepare-destroy class %q: the GCP harness prepares %q", class, prepareConsensusPeering)
	}
}

// requestConsensusPeeringRemoval requests the removal of the lane's peering
// on its own network and of the matching peering on the peer network, then
// waits until Google reads the connection DELETE_ACKNOWLEDGED. It reads the
// peering's name and both networks from the resolved manifest, so it holds
// when the deploy failed before any output existed. A peering that was never
// created, one that is INDEPENDENT, and a side whose request is already made
// need nothing.
func (h *Harness) requestConsensusPeeringRemoval(ctx context.Context, tc *provider.KindTestContext, manifestPath string) error {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the peering manifest")
	}
	var manifest struct {
		Spec struct {
			PeeringName string `json:"peeringName"`
			Network     struct {
				Value string `json:"value"`
			} `json:"network"`
			PeerNetwork struct {
				Value string `json:"value"`
			} `json:"peerNetwork"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the peering manifest")
	}
	local := networkPath.FindStringSubmatch(manifest.Spec.Network.Value)
	peer := networkPath.FindStringSubmatch(manifest.Spec.PeerNetwork.Value)
	if manifest.Spec.PeeringName == "" || local == nil || peer == nil {
		return errors.Errorf("the manifest must name the peering and both networks by self link (spec.peeringName %q, spec.network %q, spec.peerNetwork %q)",
			manifest.Spec.PeeringName, manifest.Spec.Network.Value, manifest.Spec.PeerNetwork.Value)
	}

	ours, err := h.networkPeering(ctx, local[1], local[2], func(p *compute.NetworkPeering) bool { return p.Name == manifest.Spec.PeeringName })
	if err != nil {
		return err
	}
	if ours == nil {
		lanef(tc, "  [prepare-destroy] peering %s on %s was never created or is already gone", manifest.Spec.PeeringName, local[2])
		return nil
	}
	if peeringStrategy(ours) != "CONSENSUS" {
		lanef(tc, "  [prepare-destroy] peering %s is %s: either side's destroy removes it", ours.Name, peeringStrategy(ours))
		return nil
	}
	theirs, err := h.networkPeering(ctx, peer[1], peer[2], func(p *compute.NetworkPeering) bool { return sameNetwork(p.Network, local[1], local[2]) })
	if err != nil {
		return err
	}

	sides := []struct {
		project, network string
		peering          *compute.NetworkPeering
	}{{local[1], local[2], ours}}
	if theirs != nil {
		sides = append(sides, struct {
			project, network string
			peering          *compute.NetworkPeering
		}{peer[1], peer[2], theirs})
	}
	for _, side := range sides {
		switch peeringDeleteStatus(side.peering) {
		case "LOCAL_DELETE_REQUESTED", "DELETE_ACKNOWLEDGED":
			continue
		}
		op, err := h.services.Compute.Networks.RequestRemovePeering(side.project, side.network,
			&compute.NetworksRequestRemovePeeringRequest{Name: side.peering.Name}).Context(ctx).Do()
		if err != nil {
			return errors.Wrapf(err, "requesting the removal of peering %s on %s", side.peering.Name, side.network)
		}
		if err := h.waitGlobalOperation(ctx, side.project, op); err != nil {
			return errors.Wrapf(err, "requesting the removal of peering %s on %s", side.peering.Name, side.network)
		}
		lanef(tc, "  [prepare-destroy] removal of peering %s requested on %s, as its network's owner would", side.peering.Name, side.network)
	}
	if theirs == nil {
		// One side alone is inactive and holds no agreement to wait for.
		return nil
	}

	deadline := time.Now().Add(consensusRemovalBudget)
	for {
		ours, err = h.networkPeering(ctx, local[1], local[2], func(p *compute.NetworkPeering) bool { return p.Name == manifest.Spec.PeeringName })
		if err != nil {
			return err
		}
		if ours == nil || peeringDeleteStatus(ours) == "DELETE_ACKNOWLEDGED" {
			lanef(tc, "  [prepare-destroy] peering %s reads DELETE_ACKNOWLEDGED: both owners agreed", manifest.Spec.PeeringName)
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Errorf("peering %s still reads %q %s after both removal requests", manifest.Spec.PeeringName, peeringDeleteStatus(ours), consensusRemovalBudget)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}

// networkPeering returns the network's peering that match accepts, or nil.
func (h *Harness) networkPeering(ctx context.Context, project, network string, match func(*compute.NetworkPeering) bool) (*compute.NetworkPeering, error) {
	got, err := h.services.Compute.Networks.Get(project, network).Context(ctx).Do()
	if err != nil {
		if isStatus(err, 404) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading the network %s", network)
	}
	for _, peering := range got.Peerings {
		if match(peering) {
			return peering, nil
		}
	}
	return nil, nil
}

// waitGlobalOperation waits for a global Compute operation and returns its
// error, if any.
func (h *Harness) waitGlobalOperation(ctx context.Context, project string, op *compute.Operation) error {
	for op.Status != "DONE" {
		var err error
		op, err = h.services.Compute.GlobalOperations.Wait(project, op.Name).Context(ctx).Do()
		if err != nil {
			return err
		}
	}
	if op.Error != nil && len(op.Error.Errors) > 0 {
		return errors.Errorf("%s: %s", op.Error.Errors[0].Code, op.Error.Errors[0].Message)
	}
	return nil
}

func peeringStrategy(peering *compute.NetworkPeering) string {
	if peering.ConnectionStatus != nil && peering.ConnectionStatus.UpdateStrategy != "" {
		return peering.ConnectionStatus.UpdateStrategy
	}
	if peering.UpdateStrategy != "" {
		return peering.UpdateStrategy
	}
	return "INDEPENDENT"
}

func peeringDeleteStatus(peering *compute.NetworkPeering) string {
	if peering.ConnectionStatus == nil || peering.ConnectionStatus.ConsensusState == nil {
		return ""
	}
	return peering.ConnectionStatus.ConsensusState.DeleteStatus
}

// sameNetwork reports whether a peering's network link names project/network.
func sameNetwork(link, project, network string) bool {
	match := networkPath.FindStringSubmatch(link)
	return match != nil && match[1] == project && match[2] == network
}
