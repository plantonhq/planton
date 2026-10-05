package module

import (
	"strings"

	gcploggingsinkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcploggingsink/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs. (Sinks carry no label
// surface in GCP, so there is no platform-label merge here.)
type Locals struct {
	GcpLoggingSink *gcploggingsinkv1alpha1.GcpLoggingSink

	// The cloud-side sink name defaults to metadata.name when the spec
	// leaves sink_name empty — the same naming basis every kind uses.
	SinkName string

	// The rendered destination URI (see renderDestination) — the exact
	// string the Logging API expects, assembled from whichever destination
	// arm the spec set.
	Destination string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcploggingsinkv1alpha1.GcpLoggingSinkIacInput) *Locals {
	target := iacInput.Target

	sinkName := target.Spec.SinkName
	if sinkName == "" {
		sinkName = target.Metadata.Name
	}

	return &Locals{
		GcpLoggingSink: target,
		SinkName:       sinkName,
		Destination:    renderDestination(target.Spec.Destination),
	}
}

// renderDestination assembles the service-scheme destination URI from the
// spec's destination arm — the hand-assembly this kind exists to remove.
// Exactly one arm is set (proto-CEL-enforced):
//
//	gcs_bucket       my-bucket                         -> storage.googleapis.com/my-bucket
//	bigquery_dataset projects/p/datasets/d (or the
//	                 dataset's https self_link)        -> bigquery.googleapis.com/projects/p/datasets/d
//	pubsub_topic     projects/p/topics/t               -> pubsub.googleapis.com/projects/p/topics/t
//	raw_uri          anything                          -> passed through verbatim
func renderDestination(destination *gcploggingsinkv1alpha1.GcpLoggingSinkDestination) string {
	switch {
	case destination.GetGcsBucket().GetValue() != "":
		return "storage.googleapis.com/" + destination.GetGcsBucket().GetValue()
	case destination.GetBigqueryDataset().GetValue() != "":
		// Accept both the GcpBigQueryDataset self_link output
		// (https://bigquery.googleapis.com/bigquery/v2/projects/p/datasets/d)
		// and a bare projects/p/datasets/d path — normalize either into the
		// Logging destination form.
		dataset := destination.GetBigqueryDataset().GetValue()
		dataset = strings.TrimPrefix(dataset, "https://bigquery.googleapis.com/bigquery/v2/")
		dataset = strings.TrimPrefix(dataset, "bigquery.googleapis.com/")
		return "bigquery.googleapis.com/" + dataset
	case destination.GetPubsubTopic().GetValue() != "":
		return "pubsub.googleapis.com/" + destination.GetPubsubTopic().GetValue()
	default:
		return destination.GetRawUri()
	}
}
