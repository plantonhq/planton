package aa_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/plantonhq/planton/catalog/gcp/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// fakePeeringNetworks serves two networks' peerings and records each removal
// request; a request on one side moves both sides' delete status the way
// Google does, and the second request acknowledges the deletion.
type fakePeeringNetworks struct {
	mu       sync.Mutex
	networks map[string]*compute.Network // by network name
	requests []string                    // "network/peering"
}

func (f *fakePeeringNetworks) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/compute/v1/"), "/")
	// projects/P/global/networks/N[/requestRemovePeering] or projects/P/global/operations/O[/wait]
	switch {
	case len(parts) >= 5 && parts[3] == "operations":
		_ = json.NewEncoder(w).Encode(compute.Operation{Name: parts[4], Status: "DONE"})
	case len(parts) == 5 && parts[3] == "networks":
		network, ok := f.networks[parts[4]]
		if !ok {
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(network)
	case len(parts) == 6 && parts[5] == "requestRemovePeering":
		var body compute.NetworksRequestRemovePeeringRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.requests = append(f.requests, parts[4]+"/"+body.Name)
		f.markRequested(parts[4], body.Name)
		_ = json.NewEncoder(w).Encode(compute.Operation{Name: "op-1", Status: "DONE"})
	default:
		http.Error(w, "unexpected "+r.URL.Path, http.StatusBadRequest)
	}
}

func (f *fakePeeringNetworks) markRequested(network, peering string) {
	requested := 0
	for _, n := range f.networks {
		for _, p := range n.Peerings {
			if p.ConnectionStatus == nil || p.ConnectionStatus.ConsensusState == nil {
				continue
			}
			if n.Name == network && p.Name == peering {
				p.ConnectionStatus.ConsensusState.DeleteStatus = "LOCAL_DELETE_REQUESTED"
			}
			if p.ConnectionStatus.ConsensusState.DeleteStatus == "LOCAL_DELETE_REQUESTED" {
				requested++
			}
		}
	}
	if requested == 2 {
		for _, n := range f.networks {
			for _, p := range n.Peerings {
				p.ConnectionStatus.ConsensusState.DeleteStatus = "DELETE_ACKNOWLEDGED"
			}
		}
	}
}

func peeringOf(name, peerNetwork, strategy, deleteStatus string) *compute.NetworkPeering {
	peering := &compute.NetworkPeering{
		Name:             name,
		Network:          "https://www.googleapis.com/compute/v1/projects/p/global/networks/" + peerNetwork,
		ConnectionStatus: &compute.NetworkPeeringConnectionStatus{UpdateStrategy: strategy},
	}
	if strategy == "CONSENSUS" {
		peering.ConnectionStatus.ConsensusState = &compute.NetworkPeeringConnectionStatusConsensusState{DeleteStatus: deleteStatus}
	}
	return peering
}

func peeringHarness(t *testing.T, fake *fakePeeringNetworks) (*Harness, context.Context) {
	t.Helper()
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	ctx := context.Background()
	service, err := compute.NewService(ctx, option.WithEndpoint(server.URL+"/compute/v1/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(t.TempDir(), "peering.yaml")
	body := "spec:\n  peeringName: side-a\n" +
		"  network:\n    value: https://www.googleapis.com/compute/v1/projects/p/global/networks/net-a\n" +
		"  peerNetwork:\n    value: https://www.googleapis.com/compute/v1/projects/p/global/networks/net-b\n"
	if err := os.WriteFile(manifest, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return &Harness{services: &verify.Services{Compute: service}}, context.WithValue(ctx, provider.ManifestPathKey{}, manifest)
}

// The owners' step: both sides of a CONSENSUS peering are asked to remove it
// and the connection reads DELETE_ACKNOWLEDGED; a side already asked is not
// asked again; an INDEPENDENT peering and one never created need nothing.
func TestConsensusPeeringRemovalAsksEachOwnerOnce(t *testing.T) {
	tc := &provider.KindTestContext{T: t}
	cases := []struct {
		name     string
		networks map[string]*compute.Network
		want     string
	}{
		{"consensus", map[string]*compute.Network{
			"net-a": {Name: "net-a", Peerings: []*compute.NetworkPeering{peeringOf("side-a", "net-b", "CONSENSUS", "")}},
			"net-b": {Name: "net-b", Peerings: []*compute.NetworkPeering{peeringOf("side-b", "net-a", "CONSENSUS", "")}},
		}, "net-a/side-a,net-b/side-b"},
		{"ours already requested", map[string]*compute.Network{
			"net-a": {Name: "net-a", Peerings: []*compute.NetworkPeering{peeringOf("side-a", "net-b", "CONSENSUS", "LOCAL_DELETE_REQUESTED")}},
			"net-b": {Name: "net-b", Peerings: []*compute.NetworkPeering{peeringOf("side-b", "net-a", "CONSENSUS", "PEER_DELETE_REQUESTED")}},
		}, "net-b/side-b"},
		{"independent", map[string]*compute.Network{
			"net-a": {Name: "net-a", Peerings: []*compute.NetworkPeering{peeringOf("side-a", "net-b", "INDEPENDENT", "")}},
			"net-b": {Name: "net-b", Peerings: []*compute.NetworkPeering{peeringOf("side-b", "net-a", "INDEPENDENT", "")}},
		}, ""},
		{"never created", map[string]*compute.Network{
			"net-a": {Name: "net-a"},
		}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fake := &fakePeeringNetworks{networks: c.networks}
			h, ctx := peeringHarness(t, fake)
			if err := h.PrepareDestroy(ctx, tc, prepareConsensusPeering); err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(fake.requests, ","); got != c.want {
				t.Fatalf("removal requests: got %q, want %q", got, c.want)
			}
		})
	}
}

func TestPrepareDestroyRefusesAnUnknownClass(t *testing.T) {
	h, ctx := peeringHarness(t, &fakePeeringNetworks{})
	if err := h.PrepareDestroy(ctx, &provider.KindTestContext{T: t}, "consensus-peerings"); err == nil {
		t.Fatal("an unknown class must be refused")
	}
}
