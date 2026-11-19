//go:build e2e

package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"golang.org/x/oauth2/google"
	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/clouddeploy/v1"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/discoveryengine/v1"
	"google.golang.org/api/googleapi"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// TestSweep proves the test projects hold nothing a lane left behind. It runs
// only when asked (PLANTON_E2E_GCP_SWEEP=1), at a session's end, after every
// lane: an orphan is anything in the project that is not standing, whatever
// its name. That reading is only sound in a project dedicated to the lanes, so
// the sweep is meant for one; in a shared project it reports what the other
// tenants own too.
//
// It reads every resource Cloud Asset Inventory finds in the project
// (searchAllResources) and the project's IAM policy, and fails on each line
// that is not standing. Asset Inventory indexes most resource types, not all,
// and its index keeps a deleted resource for minutes to hours, so a candidate
// whose type the sweep can read live (liveReaders) is read through its own API
// and reported only if it is still there; any other type is reported as found.
// A type Asset Inventory does not index relies on each lane's own VERIFY-CLEAN
// phase, which reads the resource itself as gone. TestMain
// has already removed crashed lanes' identities (CleanLaneIdentities) before
// this runs, and says so on its own lines.
//
// The second project, PLANTON_E2E_GCP_SECOND_PROJECT_ID, is swept too when set.
func TestSweep(t *testing.T) {
	if os.Getenv("PLANTON_E2E_GCP_SWEEP") != "1" {
		t.Skip("the sweep runs at a session's end, when asked: PLANTON_E2E_GCP_SWEEP=1")
	}
	ctx := context.Background()
	harness := harnessAccountOf(ctx, t)
	projects := []string{os.Getenv("GOOGLE_PROJECT")}
	if second := os.Getenv("PLANTON_E2E_GCP_SECOND_PROJECT_ID"); second != "" {
		projects = append(projects, second)
	}
	var leftovers []string
	for _, project := range projects {
		leftovers = append(leftovers, sweepProject(ctx, t, project, harness)...)
	}
	for _, line := range leftovers {
		t.Log(line)
	}
	if len(leftovers) > 0 {
		t.Fatalf("%d leftover(s) in %s", len(leftovers), strings.Join(projects, ", "))
	}
	t.Logf("the sweep reads 0 in %s", strings.Join(projects, ", "))
}

func sweepProject(ctx context.Context, t *testing.T, project string, harness harnessAccount) []string {
	t.Helper()
	opts := []option.ClientOption{option.WithQuotaProject(project)}
	assets, err := cloudasset.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	readers := newLiveReaders(ctx, t, opts)
	var leftovers []string
	err = assets.V1.SearchAllResources("projects/"+project).Context(ctx).Pages(ctx, func(page *cloudasset.SearchAllResourcesResponse) error {
		for _, resource := range page.Results {
			if standingResource(project, harness, resource) {
				continue
			}
			if read, ok := readers.lookup(resource.AssetType); ok {
				exists, err := read(ctx, resourcePath(resource.Name))
				if err != nil {
					return fmt.Errorf("reading %s live: %w", resource.Name, err)
				}
				if !exists {
					continue // deleted; Asset Inventory has not dropped it yet
				}
			}
			leftovers = append(leftovers, fmt.Sprintf("%s: %s %s", project, resource.AssetType, resource.Name))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("listing %s's resources through Cloud Asset Inventory: %v", project, err)
	}

	crm, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		t.Fatalf("reading %s's IAM policy: %v", project, err)
	}
	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			if strings.HasPrefix(member, "serviceAccount:pe2e-") || strings.HasPrefix(member, "deleted:") {
				leftovers = append(leftovers, fmt.Sprintf("%s: IAM binding %s on %s", project, member, binding.Role))
			}
		}
	}
	return leftovers
}

// standingResource is what a lane project holds on purpose: the project and
// its enabled services, the harness's own account and key, the standing
// declared roles, Google's default accounts and agents, and what Google
// creates when an API is turned on. Each entry says why it stands.
func standingResource(project string, harness harnessAccount, resource *cloudasset.ResourceSearchResult) bool {
	name := resource.Name
	switch resource.AssetType {
	case "cloudresourcemanager.googleapis.com/Project",
		"serviceusage.googleapis.com/Service",
		// The project's link to its billing account.
		"cloudbilling.googleapis.com/ProjectBillingInfo",
		// Compute Engine's per-project settings, present once its API is on.
		"compute.googleapis.com/Project":
		return true
	case "iam.googleapis.com/ServiceAccount", "iam.googleapis.com/ServiceAccountKey":
		// The harness's own account and its key (Asset Inventory names an
		// account by its email and a key by the account's unique id); Google's
		// default accounts.
		return strings.Contains(name, "/serviceAccounts/"+harness.email) ||
			strings.Contains(name, "/serviceAccounts/"+harness.uniqueID+"/keys/") ||
			strings.Contains(name, "-compute@developer.gserviceaccount.com") ||
			strings.Contains(name, "/serviceAccounts/"+project+"@appspot.gserviceaccount.com")
	case "iam.googleapis.com/Role":
		// The declared identities' standing roles (rewritten, never deleted).
		return strings.Contains(name, "/roles/plantonE2eDeclared")
	case "logging.googleapis.com/LogBucket", "logging.googleapis.com/LogSink":
		// Every project's two log routes.
		return strings.HasSuffix(name, "/_Default") || strings.HasSuffix(name, "/_Required")
	case "storage.googleapis.com/Bucket":
		// Cloud Build's per-project source bucket, made on its first build.
		return strings.HasSuffix(name, "/"+project+"_cloudbuild")
	case "cloudkms.googleapis.com/KeyRing", "cloudkms.googleapis.com/CryptoKey",
		"cloudkms.googleapis.com/KeyHandle", "cloudkms.googleapis.com/AutokeyConfig":
		// Cloud KMS deletes none of these: a key ring, a key and an Autokey
		// key handle keep their names for good, and a project's Autokey
		// configuration is cleared, never removed. Each run's names carry
		// its run id. What can still cost or encrypt is a key VERSION, which
		// the live reader requires destroyed or scheduled for destruction.
		return true
	}
	return false
}

// harnessAccount is the harness credential's own account, which stands in the
// project it administers: its email, and its unique id (the key file's
// client_id), by which Asset Inventory names its key.
type harnessAccount struct {
	email    string
	uniqueID string
}

func harnessAccountOf(ctx context.Context, t *testing.T) harnessAccount {
	t.Helper()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var key struct {
		ClientEmail string `json:"client_email"`
		ClientID    string `json:"client_id"`
	}
	if err := json.Unmarshal(creds.JSON, &key); err != nil || key.ClientEmail == "" || key.ClientID == "" {
		t.Fatal("the sweep runs under the harness's service-account key (GOOGLE_APPLICATION_CREDENTIALS), whose account it must leave standing")
	}
	return harnessAccount{email: key.ClientEmail, uniqueID: key.ClientID}
}

// liveReaders read a candidate through its own API, by its resource name (the
// Asset Inventory name without its //service/ host), and say whether it still
// exists. Google answers a deleted resource with 404, and with 403 where the
// parent it would be read through is gone too (a tag value under a deleted
// key, a key under a deleted account), so both read as gone. A type lands here
// when a lane first leaves it in Asset Inventory's index.
type liveReader func(ctx context.Context, name string) (bool, error)

// liveReaders maps an asset type to its reader. Every Compute Engine type
// shares one reader (computeReaderKey): Compute names a resource by the same
// path in Asset Inventory and in its REST API, so a single GET by that path
// reads any of them, from networks and tunnels to security policies.
type liveReaders map[string]liveReader

const computeReaderKey = "compute.googleapis.com/*"

func (r liveReaders) lookup(assetType string) (liveReader, bool) {
	if read, ok := r[assetType]; ok {
		return read, true
	}
	if strings.HasPrefix(assetType, "compute.googleapis.com/") {
		read, ok := r[computeReaderKey]
		return read, ok
	}
	return nil, false
}

func newLiveReaders(ctx context.Context, t *testing.T, opts []option.ClientOption) liveReaders {
	t.Helper()
	iamService, err := iam.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	crm, err := crmv3.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	rest, _, err := htransport.NewClient(ctx, append([]option.ClientOption{option.WithScopes(compute.CloudPlatformScope)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	kms, err := cloudkms.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	deploy, err := clouddeploy.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	search, err := discoveryengine.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return liveReaders{
		// projects/{p}/global/{collection}/{n}, projects/{p}/regions/{r}/...,
		// projects/{p}/zones/{z}/...: the same path under compute/v1.
		computeReaderKey: func(ctx context.Context, name string) (bool, error) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://compute.googleapis.com/compute/v1/"+name, nil)
			if err != nil {
				return false, err
			}
			response, err := rest.Do(request)
			if err != nil {
				return false, err
			}
			defer response.Body.Close()
			return stillThere(googleapi.CheckResponse(response))
		},
		// A key version that can no longer encrypt or bill reads as gone.
		"cloudkms.googleapis.com/CryptoKeyVersion": func(ctx context.Context, name string) (bool, error) {
			version, err := kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Get(name).Context(ctx).Do()
			if exists, err := stillThere(err); !exists || err != nil {
				return exists, err
			}
			return version.State != "DESTROY_SCHEDULED" && version.State != "DESTROYED", nil
		},
		"clouddeploy.googleapis.com/DeliveryPipeline": func(ctx context.Context, name string) (bool, error) {
			_, err := deploy.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"discoveryengine.googleapis.com/DataStore": func(ctx context.Context, name string) (bool, error) {
			_, err := search.Projects.Locations.Collections.DataStores.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"discoveryengine.googleapis.com/Engine": func(ctx context.Context, name string) (bool, error) {
			_, err := search.Projects.Locations.Collections.Engines.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"iam.googleapis.com/ServiceAccount": func(ctx context.Context, name string) (bool, error) {
			_, err := iamService.Projects.ServiceAccounts.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"iam.googleapis.com/ServiceAccountKey": func(ctx context.Context, name string) (bool, error) {
			_, err := iamService.Projects.ServiceAccounts.Keys.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"cloudresourcemanager.googleapis.com/TagKey": func(ctx context.Context, name string) (bool, error) {
			_, err := crm.TagKeys.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
		"cloudresourcemanager.googleapis.com/TagValue": func(ctx context.Context, name string) (bool, error) {
			_, err := crm.TagValues.Get(name).Context(ctx).Do()
			return stillThere(err)
		},
	}
}

// resourcePath turns Asset Inventory's full resource name
// (//iam.googleapis.com/projects/p/serviceAccounts/a) into the name its own API
// takes (projects/p/serviceAccounts/a).
func resourcePath(fullName string) string {
	_, path, _ := strings.Cut(strings.TrimPrefix(fullName, "//"), "/")
	return path
}

func stillThere(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
		return false, nil
	}
	return false, err
}
