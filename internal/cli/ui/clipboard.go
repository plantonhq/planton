package ui

import (
	"fmt"
	"strings"
)

const (
	maxPreviewLines = 8
	maxLineLength   = 60
)

// ClipboardEmpty displays a helpful message when clipboard is empty
func ClipboardEmpty() {
	sep := separator(infoIcon)

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("%s  %s\n", infoIcon.Render("ℹ️"), infoTitle.Render("Clipboard is Empty"))
	fmt.Println(sep)

	fmt.Println(infoMessage.Render("Copy a manifest or stack input YAML to your clipboard, then run the command again."))
	fmt.Println()

	fmt.Println(infoMessage.Render("Examples:"))
	fmt.Println()
	fmt.Printf("    %s\n", Dim("# Copy file content to clipboard (macOS)"))
	fmt.Printf("    %s\n", Cmd("pbcopy < manifest.yaml"))
	fmt.Println()
	fmt.Printf("    %s\n", Dim("# Or copy a file path"))
	fmt.Printf("    %s\n", Cmd("echo \"/path/to/manifest.yaml\" | pbcopy"))
	fmt.Println()

	fmt.Printf("%s %s\n", infoIcon.Render(iconTip),
		infoMessage.Render("Tip: You can also copy a file path - planton will read the file automatically."))

	fmt.Println(sep)
}

// ClipboardInvalidYAML displays a formatted error when clipboard content is not valid YAML
func ClipboardInvalidYAML(content []byte, parseErr error) {
	sep := separator(errorIcon)

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("%s  %s\n", errorIcon.Render(iconError), errorTitle.Render("Clipboard Content Error"))
	fmt.Println(sep)

	fmt.Println(errorMessage.Render("The clipboard content is not valid YAML."))
	fmt.Println()

	// Show content preview
	preview := formatContentPreview(content)
	fmt.Println(warningTitle.Render("Content preview:"))
	fmt.Println(preview)
	fmt.Println()

	// Show expected format
	fmt.Println(infoTitle.Render("Expected format:"))
	fmt.Println()
	fmt.Printf("    %s\n", Cmd("apiVersion: kubernetes.planton.com/v1"))
	fmt.Printf("    %s\n", Cmd("kind: PostgresKubernetes"))
	fmt.Printf("    %s\n", Cmd("metadata:"))
	fmt.Printf("    %s\n", Cmd("  name: my-postgres"))
	fmt.Printf("    %s\n", Cmd("spec:"))
	fmt.Printf("    %s\n", Cmd("  ..."))
	fmt.Println()

	// Show parse error
	if parseErr != nil {
		fmt.Printf("%s %s\n", Dim("Parse error:"), dimStyle.Render(parseErr.Error()))
		fmt.Println()
	}

	fmt.Printf("%s %s\n", infoIcon.Render(iconTip),
		infoMessage.Render("Tip: Copy a valid manifest YAML file or a path to a manifest file."))

	fmt.Println(sep)
}

// ClipboardFileNotFound displays an error when clipboard contains a file path but file doesn't exist
func ClipboardFileNotFound(filePath string) {
	sep := separator(errorIcon)

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("%s  %s\n", errorIcon.Render(iconError), errorTitle.Render("File Not Found"))
	fmt.Println(sep)

	fmt.Println(errorMessage.Render("Clipboard contains a file path, but the file doesn't exist:"))
	fmt.Println()
	fmt.Printf("    %s\n", Path(filePath))
	fmt.Println()

	fmt.Println(errorMessage.Render("This typically happens when:"))
	fmt.Printf("  %s %s\n", Dim("-"), errorMessage.Render("The file was moved or deleted"))
	fmt.Printf("  %s %s\n", Dim("-"), errorMessage.Render("The path contains a typo"))
	fmt.Printf("  %s %s\n", Dim("-"), errorMessage.Render("You're in a different directory context"))
	fmt.Println()

	fmt.Printf("%s %s\n", infoIcon.Render(iconTip),
		infoMessage.Render(fmt.Sprintf("Tip: Verify the file exists with: %s", Cmd("ls "+filePath))))

	fmt.Println(sep)
}

// ClipboardNotStackInput displays a formatted error when clipboard content
// is valid YAML but not a stack input (missing "target" field).
func ClipboardNotStackInput(content []byte) {
	sep := separator(infoIcon)

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("%s  %s\n", infoIcon.Render("ℹ️"), infoTitle.Render("Not a Stack Input"))
	fmt.Println(sep)

	fmt.Println(infoMessage.Render("The clipboard content is valid YAML but not a stack input."))
	fmt.Println(infoMessage.Render("Stack input files must have a \"target\" field at the root level."))
	fmt.Println()

	// Show content preview
	preview := formatContentPreview(content)
	fmt.Println(warningTitle.Render("Content preview:"))
	fmt.Println(preview)
	fmt.Println()

	// Show expected stack input format
	fmt.Println(infoTitle.Render("Expected stack input format:"))
	fmt.Println()
	fmt.Printf("    %s\n", Cmd("target:"))
	fmt.Printf("    %s\n", Cmd("  apiVersion: kubernetes.planton.com/v1"))
	fmt.Printf("    %s\n", Cmd("  kind: PostgresKubernetes"))
	fmt.Printf("    %s\n", Cmd("  metadata:"))
	fmt.Printf("    %s\n", Cmd("    name: my-postgres"))
	fmt.Printf("    %s\n", Cmd("  spec:"))
	fmt.Printf("    %s\n", Cmd("    ..."))
	fmt.Printf("    %s\n", Cmd("provider_config:"))
	fmt.Printf("    %s\n", Cmd("  ..."))
	fmt.Println()

	fmt.Printf("%s %s\n", infoIcon.Render(iconTip),
		infoMessage.Render("Tip: If your clipboard contains a raw manifest (not stack input),"))
	fmt.Println(infoMessage.Render("     use '--clip' without '-i' and it will be detected automatically."))

	fmt.Println(sep)
}

// ClipboardFileLoaded displays a success message when loading from a file path in clipboard
func ClipboardFileLoaded(filePath string) {
	fmt.Printf("%s  %s: %s\n",
		successIcon.Render(iconSuccess),
		successTitle.Render("Loaded from clipboard path"),
		Path(filePath))
}

// formatContentPreview creates a formatted preview of clipboard content
func formatContentPreview(content []byte) string {
	lines := strings.Split(string(content), "\n")

	// Truncate to max lines
	truncated := false
	if len(lines) > maxPreviewLines {
		lines = lines[:maxPreviewLines]
		truncated = true
	}

	var sb strings.Builder

	// Content lines with indentation
	for _, line := range lines {
		// Truncate long lines
		displayLine := line
		if len(line) > maxLineLength {
			displayLine = line[:maxLineLength-3] + "..."
		}
		sb.WriteString("   ")
		sb.WriteString(dimStyle.Render(displayLine))
		sb.WriteString("\n")
	}

	// Show truncation indicator
	if truncated {
		sb.WriteString("   ")
		sb.WriteString(dimStyle.Render("..."))
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
