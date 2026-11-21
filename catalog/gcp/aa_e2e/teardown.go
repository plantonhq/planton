package aa_e2e

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

var _ provider.DependencyTeardownPreparer = (*Harness)(nil)

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
