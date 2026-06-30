package pulumigoogleprovider

import (
	"encoding/json"
	"fmt"
	"reflect"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Get(ctx *pulumi.Context, gcpProviderConfig *gcpprovider.GcpProviderConfig,
	nameSuffixes ...string) (*gcp.Provider, error) {
	gcpProviderArgs := &gcp.ProviderArgs{}

	if gcpProviderConfig != nil && gcpProviderConfig.ServiceAccountKey != "" {
		var serviceAccountKeyMap map[string]interface{}
		if err := json.Unmarshal([]byte(gcpProviderConfig.ServiceAccountKey), &serviceAccountKeyMap); err != nil {
			return nil, errors.Wrap(err, "failed to parse service account key JSON. "+
				"Ensure the value is a valid GCP Service Account key file containing fields: "+
				"type, project_id, private_key_id, private_key, client_email, client_id, auth_uri, token_uri")
		}

		requiredFields := []string{"type", "project_id", "private_key", "client_email"}
		for _, field := range requiredFields {
			if _, ok := serviceAccountKeyMap[field]; !ok {
				return nil, errors.Errorf("service account key JSON is missing required field: %s", field)
			}
		}

		privateKey, ok := serviceAccountKeyMap["private_key"].(string)
		if !ok {
			return nil, errors.New("service account key 'private_key' field must be a string")
		}
		if len(privateKey) > 11 && privateKey[:11] != "-----BEGIN " {
			return nil, errors.New("service account key 'private_key' field must be a PEM-encoded key " +
				"(starting with '-----BEGIN PRIVATE KEY-----'). " +
				"Ensure you're using a JSON key file from GCP, not a P12/PKCS12 key")
		}

		gcpProviderArgs = &gcp.ProviderArgs{Credentials: pulumi.String(gcpProviderConfig.ServiceAccountKey)}
	}

	googleProvider, err := gcp.NewProvider(ctx, ProviderResourceName(nameSuffixes), gcpProviderArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create google provider")
	}

	return googleProvider, nil
}

func ProviderResourceName(suffixes []string) string {
	name := "google"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	outputName := fmt.Sprintf("gcp_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		outputName = fmt.Sprintf("%s_%s", outputName, s)
	}
	return outputName
}
