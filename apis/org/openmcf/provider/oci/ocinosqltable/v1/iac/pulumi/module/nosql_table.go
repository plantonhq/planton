package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocinosqltablev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocinosqltable/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/nosql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func nosqlTable(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciNosqlTable.Spec

	tableLimitsArgs := &nosql.TableTableLimitsArgs{
		MaxReadUnits:    pulumi.Int(int(spec.TableLimits.MaxReadUnits)),
		MaxWriteUnits:   pulumi.Int(int(spec.TableLimits.MaxWriteUnits)),
		MaxStorageInGbs: pulumi.Int(int(spec.TableLimits.MaxStorageInGbs)),
	}

	if spec.TableLimits.CapacityMode != ocinosqltablev1.OciNosqlTableSpec_TableLimits_capacity_mode_unspecified {
		tableLimitsArgs.CapacityMode = pulumi.StringPtr(strings.ToUpper(spec.TableLimits.CapacityMode.String()))
	}

	tableArgs := &nosql.TableArgs{
		CompartmentId:    pulumi.String(spec.CompartmentId.GetValue()),
		Name:             pulumi.String(spec.Name),
		DdlStatement:     pulumi.String(spec.DdlStatement),
		TableLimits:      tableLimitsArgs,
		IsAutoReclaimable: pulumi.Bool(spec.IsAutoReclaimable),
		FreeformTags:     pulumi.ToStringMap(locals.FreeformTags),
	}

	table, err := nosql.NewTable(ctx, locals.TableName, tableArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create nosql table")
	}

	ctx.Export(OpTableId, table.ID())

	for _, idx := range spec.Indexes {
		keys := make(nosql.IndexKeyArray, len(idx.Keys))
		for i, key := range idx.Keys {
			keyArgs := nosql.IndexKeyArgs{
				ColumnName: pulumi.String(key.ColumnName),
			}
			if key.JsonFieldType != "" {
				keyArgs.JsonFieldType = pulumi.StringPtr(key.JsonFieldType)
			}
			if key.JsonPath != "" {
				keyArgs.JsonPath = pulumi.StringPtr(key.JsonPath)
			}
			keys[i] = keyArgs
		}

		indexName := fmt.Sprintf("%s-%s", locals.TableName, idx.Name)
		_, err := nosql.NewIndex(ctx, indexName, &nosql.IndexArgs{
			TableNameOrId: table.ID(),
			Name:          pulumi.String(idx.Name),
			Keys:          keys,
		}, pulumiOciOpt(provider), pulumi.DependsOn([]pulumi.Resource{table}))
		if err != nil {
			return errors.Wrapf(err, "failed to create nosql index %s", idx.Name)
		}
	}

	return nil
}
