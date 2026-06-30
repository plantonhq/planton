package module

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type tableResult struct {
	Table *dynamodb.Table
}

func createTable(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*tableResult, error) {
	// Get billing mode directly from enum (values match AWS API strings)
	var billingMode pulumi.StringPtrInput
	if locals.Spec.BillingMode != 0 {
		billingMode = pulumi.StringPtr(locals.Spec.BillingMode.String())
	}

	// Attributes - get type directly from enum
	var attrDefs dynamodb.TableAttributeArray
	for _, a := range locals.Spec.AttributeDefinitions {
		attrType := "S" // default to String
		if a.Type != 0 {
			attrType = a.Type.String()
		}
		attrDefs = append(attrDefs, dynamodb.TableAttributeArgs{
			Name: pulumi.String(a.Name),
			Type: pulumi.String(attrType),
		})
	}

	// Table keys - get key type directly from enum
	var tableHashKey, tableRangeKey string
	for _, k := range locals.Spec.KeySchema {
		keyType := k.KeyType.String()
		if keyType == "HASH" {
			tableHashKey = k.AttributeName
		}
		if keyType == "RANGE" {
			tableRangeKey = k.AttributeName
		}
	}

	// LSI
	var lsiArgs dynamodb.TableLocalSecondaryIndexArray
	for _, l := range locals.Spec.LocalSecondaryIndexes {
		var lsiRangeKey string
		for _, lk := range l.KeySchema {
			if lk.KeyType.String() == "RANGE" {
				lsiRangeKey = lk.AttributeName
			}
		}
		lsi := dynamodb.TableLocalSecondaryIndexArgs{
			Name:     pulumi.String(l.Name),
			RangeKey: pulumi.String(lsiRangeKey),
		}
		// Default projection ALL; override per spec
		lsi.ProjectionType = pulumi.String("ALL")
		if l.Projection != nil {
			projType := l.Projection.Type.String()
			switch projType {
			case "KEYS_ONLY_PROJECTION":
				lsi.ProjectionType = pulumi.String("KEYS_ONLY")
			case "INCLUDE":
				lsi.ProjectionType = pulumi.String("INCLUDE")
				var nonKeys pulumi.StringArray
				for _, n := range l.Projection.NonKeyAttributes {
					nonKeys = append(nonKeys, pulumi.String(n))
				}
				lsi.NonKeyAttributes = nonKeys
			case "ALL":
				lsi.ProjectionType = pulumi.String("ALL")
			}
		}
		lsiArgs = append(lsiArgs, lsi)
	}

	// GSI
	var gsiArgs dynamodb.TableGlobalSecondaryIndexArray
	for _, g := range locals.Spec.GlobalSecondaryIndexes {
		var gsiHashKey, gsiRangeKey string
		for _, gk := range g.KeySchema {
			keyType := gk.KeyType.String()
			if keyType == "HASH" {
				gsiHashKey = gk.AttributeName
			}
			if keyType == "RANGE" {
				gsiRangeKey = gk.AttributeName
			}
		}
		gsi := dynamodb.TableGlobalSecondaryIndexArgs{
			Name:    pulumi.String(g.Name),
			HashKey: pulumi.String(gsiHashKey),
		}
		if gsiRangeKey != "" {
			gsi.RangeKey = pulumi.StringPtr(gsiRangeKey)
		}
		// Default projection ALL; override per spec
		gsi.ProjectionType = pulumi.String("ALL")
		if g.Projection != nil {
			projType := g.Projection.Type.String()
			switch projType {
			case "KEYS_ONLY_PROJECTION":
				gsi.ProjectionType = pulumi.String("KEYS_ONLY")
			case "INCLUDE":
				gsi.ProjectionType = pulumi.String("INCLUDE")
				var nonKeys pulumi.StringArray
				for _, n := range g.Projection.NonKeyAttributes {
					nonKeys = append(nonKeys, pulumi.String(n))
				}
				gsi.NonKeyAttributes = nonKeys
			case "ALL":
				gsi.ProjectionType = pulumi.String("ALL")
			}
		}
		if g.ProvisionedThroughput != nil && locals.Spec.BillingMode.String() == "PROVISIONED" {
			gsi.ReadCapacity = pulumi.IntPtr(int(g.ProvisionedThroughput.ReadCapacityUnits))
			gsi.WriteCapacity = pulumi.IntPtr(int(g.ProvisionedThroughput.WriteCapacityUnits))
		}
		gsiArgs = append(gsiArgs, gsi)
	}

	// SSE
	var sseArgs *dynamodb.TableServerSideEncryptionArgs
	if locals.Spec.ServerSideEncryption != nil && locals.Spec.ServerSideEncryption.Enabled {
		sseArgs = &dynamodb.TableServerSideEncryptionArgs{
			Enabled:   pulumi.Bool(true),
			KmsKeyArn: pulumi.StringPtr(locals.Spec.ServerSideEncryption.KmsKeyArn),
		}
	}

	args := &dynamodb.TableArgs{
		Attributes:             attrDefs,
		HashKey:                pulumi.String(tableHashKey),
		BillingMode:            billingMode,
		LocalSecondaryIndexes:  lsiArgs,
		GlobalSecondaryIndexes: gsiArgs,
		ServerSideEncryption:   sseArgs,
		Tags:                   pulumi.ToStringMap(locals.AwsTags),
		StreamEnabled:          pulumi.BoolPtr(locals.Spec.StreamEnabled),
	}
	if tableRangeKey != "" {
		args.RangeKey = pulumi.StringPtr(tableRangeKey)
	}

	if locals.Spec.StreamEnabled && locals.Spec.StreamViewType != 0 {
		args.StreamViewType = pulumi.StringPtr(locals.Spec.StreamViewType.String())
	}

	if locals.Spec.TableClass != 0 && locals.Spec.TableClass.String() != "STANDARD" {
		args.TableClass = pulumi.StringPtr(locals.Spec.TableClass.String())
	}
	if locals.Spec.PointInTimeRecoveryEnabled {
		args.PointInTimeRecovery = &dynamodb.TablePointInTimeRecoveryArgs{Enabled: pulumi.Bool(true)}
	}
	if locals.Spec.BillingMode.String() == "PROVISIONED" && locals.Spec.ProvisionedThroughput != nil {
		args.ReadCapacity = pulumi.IntPtr(int(locals.Spec.ProvisionedThroughput.ReadCapacityUnits))
		args.WriteCapacity = pulumi.IntPtr(int(locals.Spec.ProvisionedThroughput.WriteCapacityUnits))
	}

	table, err := dynamodb.NewTable(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	return &tableResult{Table: table}, nil
}
