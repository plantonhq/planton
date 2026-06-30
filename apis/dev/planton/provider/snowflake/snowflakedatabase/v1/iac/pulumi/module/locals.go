package module

import (
	snowflakedatabasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/snowflake/snowflakedatabase/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	SnowflakeDatabase *snowflakedatabasev1.SnowflakeDatabase
}

func initializeLocals(ctx *pulumi.Context, stackInput *snowflakedatabasev1.SnowflakeDatabaseStackInput) *Locals {
	locals := &Locals{}

	//assign value for the locals variable to make it available across the project
	locals.SnowflakeDatabase = stackInput.Target

	return locals
}
