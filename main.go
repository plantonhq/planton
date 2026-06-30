package main

import (
	"github.com/plantonhq/planton/cmd/planton"
	clipanic "github.com/plantonhq/planton/internal/cli/panic"
)

func main() {
	finished := new(bool)
	defer clipanic.Handle(finished)
	planton.Execute()
	*finished = true
}
