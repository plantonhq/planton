//go:build e2e

package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"golang.org/x/oauth2/google"
	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
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
// and lags writes by minutes; a type it does not index relies on each lane's
// own VERIFY-CLEAN phase, which reads the resource itself as gone. TestMain
// has already removed crashed lanes' identities (CleanLaneIdentities) before
// this runs, and says so on its own lines.
//
// The second project, PLANTON_E2E_GCP_SECOND_PROJECT_ID, is swept too when set.
func TestSweep(t *testing.T) {
	if os.Getenv("PLANTON_E2E_GCP_SWEEP") != "1" {
		t.Skip("the sweep runs at a session's end, when asked: PLANTON_E2E_GCP_SWEEP=1")
	}
	ctx := context.Background()
	harnessEmail := harnessAccountEmail(ctx, t)
	projects := []string{os.Getenv("GOOGLE_PROJECT")}
	if second := os.Getenv("PLANTON_E2E_GCP_SECOND_PROJECT_ID"); second != "" {
		projects = append(projects, second)
	}
	var leftovers []string
	for _, project := range projects {
		leftovers = append(leftovers, sweepProject(ctx, t, project, harnessEmail)...)
	}
	for _, line := range leftovers {
		t.Log(line)
	}
	if len(leftovers) > 0 {
		t.Fatalf("%d leftover(s) in %s", len(leftovers), strings.Join(projects, ", "))
	}
	t.Logf("the sweep reads 0 in %s", strings.Join(projects, ", "))
}

func sweepProject(ctx context.Context, t *testing.T, project, harnessEmail string) []string {
	t.Helper()
	opts := []option.ClientOption{option.WithQuotaProject(project)}
	assets, err := cloudasset.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	var leftovers []string
	err = assets.V1.SearchAllResources("projects/"+project).Context(ctx).Pages(ctx, func(page *cloudasset.SearchAllResourcesResponse) error {
		for _, resource := range page.Results {
			if standingResource(project, harnessEmail, resource) {
				continue
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
func standingResource(project, harnessEmail string, resource *cloudasset.ResourceSearchResult) bool {
	name := resource.Name
	switch resource.AssetType {
	case "cloudresourcemanager.googleapis.com/Project",
		"serviceusage.googleapis.com/Service",
		// Compute Engine's per-project settings, present once its API is on.
		"compute.googleapis.com/Project":
		return true
	case "iam.googleapis.com/ServiceAccount", "iam.googleapis.com/ServiceAccountKey":
		// The harness's own account and its key; Google's default accounts.
		return strings.Contains(name, "/serviceAccounts/"+harnessEmail) ||
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
	}
	return false
}

// harnessAccountEmail is the harness credential's own account, which stands in
// the project it administers.
func harnessAccountEmail(ctx context.Context, t *testing.T) string {
	t.Helper()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var key struct {
		ClientEmail string `json:"client_email"`
	}
	if err := json.Unmarshal(creds.JSON, &key); err != nil || key.ClientEmail == "" {
		t.Fatal("the sweep runs under the harness's service-account key (GOOGLE_APPLICATION_CREDENTIALS), whose account it must leave standing")
	}
	return key.ClientEmail
}
