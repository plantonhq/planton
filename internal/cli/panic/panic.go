package panic

import (
	"fmt"
	"github.com/plantonhq/planton/internal/cli/version"
	"os"
	"runtime"
	"runtime/debug"
)

// Handle displays an emergency error message to the user and a stack trace to
// report the panic.
//
// finished should be set to false when the handler is deferred and set to true as the
// last statement in the scope. This trick is necessary to avoid catching and then
// discarding a panic(nil).
func Handle(finished *bool) {
	if panicPayload := recover(); !*finished {
		stack := string(debug.Stack())
		fmt.Fprintln(os.Stderr, "================================================================================")
		fmt.Fprintln(os.Stderr, "The Pronect Planton CLI encountered a fatal error. This is a bug!")
		fmt.Fprintln(os.Stderr, "We would appreciate a report: https://github.com/plantonhq/planton/issues/")
		fmt.Fprintln(os.Stderr, "Please provide all of the below text in your report.")
		fmt.Fprintln(os.Stderr, "================================================================================")
		fmt.Fprintf(os.Stderr, "CLI Version:   %s\n", version.Version)
		fmt.Fprintf(os.Stderr, "Go Version:       %s\n", runtime.Version())
		fmt.Fprintf(os.Stderr, "Go Compiler:      %s\n", runtime.Compiler)
		fmt.Fprintf(os.Stderr, "Architecture:     %s\n", runtime.GOARCH)
		fmt.Fprintf(os.Stderr, "Operating System: %s\n", runtime.GOOS)
		fmt.Fprintf(os.Stderr, "Panic:            %s\n\n", panicPayload)
		fmt.Fprintln(os.Stderr, stack)
		os.Exit(1)
	}
}
