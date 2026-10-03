package e2ediscover

import (
	"fmt"
	"io"
	"strings"

	"github.com/plantonhq/planton/pkg/e2e/profile"
	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/shared"
)

// RenderTable writes a plain-text table to w. No ANSI codes, suitable for piping.
func RenderTable(w io.Writer, result *profile.DiscoverResult) error {
	fmt.Fprintf(w, "%-5s %-40s %-10s %-9s %s\n", "TIER", "KIND", "STATUS", "PROV", "TIMEOUT")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 80))

	var currentTier int32
	for _, ce := range result.Kinds {
		spec := ce.Profile.Spec
		if spec == nil {
			continue
		}

		if spec.Tier != currentTier {
			if currentTier != 0 {
				fmt.Fprintln(w)
			}
			currentTier = spec.Tier
		}

		statusStr := statusName(spec.Status)
		provStr := provisionerShorthand(spec.ValidatedProvisioners)
		timeout := fmt.Sprintf("%dm", spec.TimeoutMinutes)

		fmt.Fprintf(w, "%-5d %-40s %-10s %-9s %s\n",
			spec.Tier, ce.Name, statusStr, provStr, timeout)
	}

	fmt.Fprintln(w)
	counts := profile.CountByStatus(result)
	fmt.Fprintf(w, "Summary: %d GREEN, %d DEFERRED, %d SKIP, %d STUB, %d REAL-CLUSTER, %d PENDING-PROOF (%d total)\n",
		counts.Green, counts.Deferred, counts.Skip, counts.Stub, counts.RealCluster, counts.PendingProof, counts.Total)

	return nil
}

func statusName(s kindv1.CatalogKindE2EProfileSpec_Status) string {
	switch s {
	case kindv1.CatalogKindE2EProfileSpec_green:
		return "GREEN"
	case kindv1.CatalogKindE2EProfileSpec_deferred:
		return "DEFERRED"
	case kindv1.CatalogKindE2EProfileSpec_skip:
		return "SKIP"
	case kindv1.CatalogKindE2EProfileSpec_stub:
		return "STUB"
	case kindv1.CatalogKindE2EProfileSpec_real_cluster:
		return "REAL-CLUSTER"
	case kindv1.CatalogKindE2EProfileSpec_pending_proof:
		return "PENDING-PROOF"
	default:
		return "UNKNOWN"
	}
}

func provisionerShorthand(provisioners []sharedpb.IacProvisioner) string {
	var parts []string
	for _, p := range provisioners {
		switch p {
		case sharedpb.IacProvisioner_pulumi:
			parts = append(parts, "P")
		case sharedpb.IacProvisioner_terraform, sharedpb.IacProvisioner_tofu:
			parts = append(parts, "T")
		}
	}
	return strings.Join(parts, " ")
}
