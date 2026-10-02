package module

import (
	"regexp"
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdatastreamstreamv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamstream/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpDatastreamStream *gcpdatastreamstreamv1alpha1.GcpDatastreamStream
	GcpLabels           map[string]string

	// StreamId is spec.stream_id when set, otherwise metadata.name --
	// identical to the Terraform module's locals.stream_id.
	StreamId string

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// identical to the Terraform module's locals.display_name.
	DisplayName string

	// DesiredState is spec.desired_state when set, otherwise NOT_STARTED --
	// always sent, identical to the Terraform module's locals.desired_state.
	DesiredState string
}

// bigQueryApiPrefix is what a GcpBigQueryDataset self link starts with;
// Datastream wants the projects/{project}/datasets/{dataset} form.
const bigQueryApiPrefix = "https://bigquery.googleapis.com/bigquery/v2/"

// bigQueryConnectionPath matches a GcpBigQueryConnection's full name.
var bigQueryConnectionPath = regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/connections/([^/]+)$`)

// datasetId trims the API prefix from a dataset self link; a literal in
// either of Google's forms passes through -- identical to the Terraform
// module's locals.single_target_dataset_id.
func datasetId(value string) string {
	return strings.TrimPrefix(value, bigQueryApiPrefix)
}

// bigQueryConnectionName converts a connection's full name to the
// {project}.{location}.{connection_id} form BigLake managed tables want; a
// literal in the dotted form passes through -- identical to the Terraform
// module's locals.blmt_connection_name.
func bigQueryConnectionName(value string) string {
	if match := bigQueryConnectionPath.FindStringSubmatch(value); match != nil {
		return strings.Join(match[1:], ".")
	}
	return value
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDatastreamStream = stackInput.Target
	metadata := locals.GcpDatastreamStream.Metadata
	spec := locals.GcpDatastreamStream.Spec

	locals.StreamId = spec.StreamId
	if locals.StreamId == "" {
		locals.StreamId = metadata.Name
	}
	locals.DisplayName = spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = metadata.Name
	}
	locals.DesiredState = spec.DesiredState
	if locals.DesiredState == "" {
		locals.DesiredState = "NOT_STARTED"
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDatastreamStream.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
