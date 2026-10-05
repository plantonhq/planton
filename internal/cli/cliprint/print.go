package cliprint

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintDefaultSuccess() {
	fmt.Printf("success %s\n", GreenTick)
}

func PrintSuccessMessage(msg string) {
	fmt.Printf("%s %s\n", msg, GreenTick)
}

func PrintError(error string) {
	fmt.Printf("%s %s\n", error, RedTick)
}

// PrintStep prints a step in the process with a blue dot
func PrintStep(msg string) {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s %s\n", BlueDot, cyan(msg))
}

// PrintSuccess prints a success message with a green checkmark
func PrintSuccess(msg string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s %s\n", CheckMark, green(msg))
}

// PrintInfo prints an informational message with a package icon
func PrintInfo(msg string) {
	fmt.Printf("%s %s\n", Package, msg)
}

// PrintWarning prints a warning message with a yellow warning sign
func PrintWarning(msg string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s %s\n", Warning, yellow(msg))
}

// PrintHandoff prints a handoff message when transitioning to external tools
func PrintHandoff(tool string) {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Printf("%s %s\n", Handshake, cyan("Handing off to "+tool+"..."))
	fmt.Printf("   %s\n", yellow("Output below is from "+tool))
	fmt.Println()
}

// PrintPulumiSuccess prints a success message after Pulumi execution completes
func PrintPulumiSuccess() {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Printf("%s %s\n", CheckMark, green("Pulumi execution completed successfully"))
}

// PrintTofuSuccess prints a success message after OpenTofu execution completes
func PrintTofuSuccess() {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Printf("%s %s\n", CheckMark, green("OpenTofu execution completed successfully"))
}

// PrintTofuFailure prints a failure message after OpenTofu execution fails
func PrintTofuFailure() {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println()
	fmt.Printf("%s %s\n", RedTick, red("OpenTofu execution failed"))
	fmt.Printf("   %s\n", yellow("Check the above output from OpenTofu CLI to understand the root cause"))
	fmt.Println()
}

// PrintTerraformSuccess prints a success message after Terraform execution completes
func PrintTerraformSuccess() {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Printf("%s %s\n", CheckMark, green("Terraform execution completed successfully"))
}

// PrintTerraformFailure prints a failure message after Terraform execution fails
func PrintTerraformFailure() {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println()
	fmt.Printf("%s %s\n", RedTick, red("Terraform execution failed"))
	fmt.Printf("   %s\n", yellow("Check the above output from Terraform CLI to understand the root cause"))
	fmt.Println()
}

// PrintBackendConfig displays backend configuration details in a beautiful format.
// Shows backend type, bucket/container, and key (state file path) clearly formatted.
func PrintBackendConfig(backendType, bucket, key string) {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Println()
	fmt.Printf("%s %s\n", Package, cyan("State Backend Configuration"))
	fmt.Printf("   %-12s %s\n", white("Type:"), blue(backendType))
	fmt.Printf("   %-12s %s\n", white("Bucket:"), blue(bucket))
	fmt.Printf("   %-12s %s\n", white("Key:"), blue(key))
	fmt.Println()
}

// PrintModulePath displays the module path being used.
func PrintModulePath(modulePath string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s %s: %s\n", Gear, "Module path", yellow(modulePath))
}

// PrintProviderDetected prints a message showing the detected provider.
func PrintProviderDetected(kindName, providerName string) {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s Detected resource: %s (requires %s credentials)\n",
		BlueDot, cyan(kindName), cyan(providerName))
}

// PrintMissingProviderConfig prints an error when provider config is missing.
func PrintMissingProviderConfig(title string, guidance string) {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Printf("%s %s\n", RedTick, red(title))
	fmt.Println()
	fmt.Println(yellow(guidance))
}

// PrintKindDetectionError prints an error when catalog kind cannot be detected.
func PrintKindDetectionError(guidance string) {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Printf("%s %s\n", RedTick, red("Could not detect catalog kind from manifest"))
	fmt.Println()
	fmt.Println(yellow(guidance))
}

// PrintInvalidProviderConfig prints an error when provider config is invalid.
func PrintInvalidProviderConfig(title string, guidance string) {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Printf("%s %s\n", RedTick, red(title))
	fmt.Println()
	fmt.Println(yellow(guidance))
}

// PrintProviderConfigLoaded prints a success message when provider config is loaded.
func PrintProviderConfigLoaded(providerName string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s %s\n", CheckMark, green("Loaded "+providerName+" provider credentials"))
}
