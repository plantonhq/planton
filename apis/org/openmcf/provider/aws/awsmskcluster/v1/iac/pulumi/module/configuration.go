package module

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/msk"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configuration creates an inline MSK Configuration when server_properties are provided.
// The configuration holds Kafka server.properties overrides that are applied to the cluster.
func configuration(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*msk.Configuration, error) {
	spec := locals.AwsMskCluster.Spec
	if spec == nil || len(spec.ServerProperties) == 0 {
		return nil, nil
	}

	serverProps := serializeProperties(spec.ServerProperties)

	config, err := msk.NewConfiguration(ctx, "kafka-config", &msk.ConfigurationArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-config", locals.AwsMskCluster.Metadata.Id)),
		KafkaVersions:    pulumi.StringArray{pulumi.String(spec.KafkaVersion)},
		ServerProperties: pulumi.String(serverProps),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create msk configuration")
	}

	return config, nil
}

// serializeProperties converts a map to Apache Kafka .properties format (key=value per line).
// Keys are sorted for deterministic output.
func serializeProperties(props map[string]string) string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s = %s", k, props[k]))
	}
	return strings.Join(lines, "\n")
}
