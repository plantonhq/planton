package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// kafkaSchema registers the schema subject and exports its outputs. Every
// argument is create-only upstream: any change replaces the subject and
// drops all prior versions.
func kafkaSchema(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseKafkaSchemaRegistry, error) {
	spec := locals.DigitalOceanDatabaseKafkaSchema.Spec

	schema, err := canonicalSchema(spec.SchemaType, spec.Schema)
	if err != nil {
		return nil, err
	}

	createdSchema, err := digitalocean.NewDatabaseKafkaSchemaRegistry(
		ctx,
		"schema",
		&digitalocean.DatabaseKafkaSchemaRegistryArgs{
			ClusterId:   pulumi.String(spec.Cluster.GetValue()),
			SubjectName: pulumi.String(spec.SubjectName),
			SchemaType:  pulumi.String(spec.SchemaType),
			Schema:      pulumi.String(schema),
		},
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register digitalocean kafka schema subject")
	}

	ctx.Export(OpClusterId, createdSchema.ClusterId)
	ctx.Export(OpSubjectName, createdSchema.SubjectName)

	return createdSchema, nil
}

// canonicalSchema renders a JSON-typed schema (avro, json) into the form the
// registry stores it in -- object keys sorted, no whitespace -- so the
// manifest and the provider's read-back agree on every plan. The registry
// canonicalizes every JSON schema it accepts and the provider's Read stores
// that text verbatim, so a schema sent in any other key order would re-plan
// a REPLACE forever. encoding/json's Marshal sorts map keys and emits compact
// output -- byte-identical to the Terraform module's jsonencode(jsondecode())
// (both also escape `<`, `>`, `&` as \u003c/\u003e/\u0026 and keep non-ASCII
// text raw, so the two engines stay twins). Protobuf schemas are text, not
// JSON, and pass through untouched.
func canonicalSchema(schemaType, schema string) (string, error) {
	if schemaType == "protobuf" {
		return schema, nil
	}
	var document interface{}
	if err := json.Unmarshal([]byte(schema), &document); err != nil {
		return "", errors.Wrapf(err, "%s schema is not valid JSON", schemaType)
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "failed to render the schema in canonical form")
	}
	return string(canonical), nil
}
